package image_processor

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"fmt"
	"image/gif"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SevenTV/Common/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"github.com/seventv/image-processor/go/container"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/task"
	"go.uber.org/multierr"
	"golang.org/x/crypto/sha3"
)

type Worker struct{}

func (w Worker) Work(ctx global.Context, task task.Task, result *Result) (err error) {
	if result == nil {
		return fmt.Errorf("nil for result")
	}
	finish := ctx.Inst().Prometheus.StartTask()

	result.StartedAt = time.Now()
	defer func() {
		if pnk := recover(); pnk != nil {
			err = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
		}

		result.FinishedAt = time.Now()
		finish(err == nil)
	}()

	id := uuid.New().String()
	tmpDir := path.Join(ctx.Config().Worker.TempDir, id)
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	done := ctx.Inst().Prometheus.DownloadFile()
	raw, match, inputFile, err := w.downloadFile(ctx, task, tmpDir, result)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at download file"), err)
	}
	done()

	ctx.Inst().Prometheus.InputFileType(match.MIME.Value)

	ctx.Inst().Prometheus.TotalBytesDownloaded(len(raw))

	done = ctx.Inst().Prometheus.ExportFrames()
	delays, inputDir, err := w.exportFrames(ctx, tmpDir, inputFile, match, raw)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at export frames"), err)
	}
	done()

	ctx.Inst().Prometheus.TotalFramesProcessed(len(delays))

	done = ctx.Inst().Prometheus.ResizeFrames()
	variantsDir, err := w.resizeFrames(ctx, inputDir, tmpDir, task, delays)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at resize file"), err)
	}
	done()

	done = ctx.Inst().Prometheus.MakeResults()
	resultsDir, err := w.makeResults(tmpDir, delays, task, variantsDir, ctx, inputDir, inputFile)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at make results"), err)
	}
	done()

	done = ctx.Inst().Prometheus.UploadResults()
	if err = multierr.Append(
		w.uploadResults(tmpDir, resultsDir, variantsDir, task, result, ctx),
		ctx.Err(),
	); err != nil {
		return err
	}
	done()

	return nil
}

func (Worker) downloadFile(ctx global.Context, task task.Task, tmpDir string, result *Result) (raw []byte, match types.Type, inputFile string, err error) {
	defer func() {
		if pnk := recover(); pnk != nil {
			err = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
		}
	}()

	buf := aws.NewWriteAtBuffer([]byte{})

	err = ctx.Inst().S3.DownloadFile(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(task.Input.Bucket),
		Key:    aws.String(task.Input.Key),
	})
	if err != nil {
		return nil, types.Type{}, "", multierr.Append(fmt.Errorf("failed at s3 download"), err)
	}

	match = container.Match(buf.Bytes())
	switch match {
	case matchers.TypeWebp,
		matchers.TypeGif,
		matchers.TypePng,
		matchers.TypeMp4,
		matchers.TypeFlv,
		matchers.TypeAvi,
		matchers.TypeMov,
		matchers.TypeJpeg,
		matchers.TypeTiff,
		matchers.TypeWebm,
		container.TypeAvif:
	default:
		return nil, types.Type{}, "", fmt.Errorf("failed at match: unsupported image format: %v", match.Extension)
	}

	inputFile = path.Join(tmpDir, fmt.Sprintf("input.%s", match.Extension))
	file, err := os.Create(inputFile)
	if err != nil {
		return nil, types.Type{}, "", multierr.Append(fmt.Errorf("failed at create dir"), err)
	}

	_, err = file.Write(buf.Bytes())
	if err != nil {
		return nil, types.Type{}, "", multierr.Append(fmt.Errorf("failed at write file"), multierr.Append(err, file.Close()))
	}

	err = file.Close()
	if err != nil {
		return nil, types.Type{}, "", multierr.Append(fmt.Errorf("failed at close file"), err)
	}

	h := sha3.New512()
	_, err = h.Write(buf.Bytes())
	if err != nil {
		return nil, types.Type{}, "", multierr.Append(fmt.Errorf("failed at hash input file"), err)
	}

	result.InputSHA3 = hex.EncodeToString(h.Sum(nil))
	return buf.Bytes(), match, inputFile, nil
}

func (Worker) uploadResults(tmpDir string, resultsDir string, variantsDir string, task task.Task, result *Result, ctx global.Context) (err error) {
	defer func() {
		if pnk := recover(); pnk != nil {
			err = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
		}
	}()

	zipFilePath := path.Join(tmpDir, "emote.zip")
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at create zip file"), err)
	}

	zipWriter := zip.NewWriter(zipFile)
	walker := func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(pth)
		if err != nil {
			return multierr.Append(fmt.Errorf("failed at open file %s", pth), err)
		}
		defer file.Close()

		f, err := zipWriter.Create(strings.TrimLeft(pth, tmpDir))
		if err != nil {
			return multierr.Append(fmt.Errorf("failed at create zip file"), err)
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return multierr.Append(fmt.Errorf("failed at io copy"), err)
		}

		return nil
	}

	err = filepath.Walk(resultsDir, walker)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at walk resultsDir"), multierr.Append(err, multierr.Append(zipWriter.Close(), zipFile.Close())))
	}

	err = filepath.Walk(variantsDir, walker)
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at walk variantsDir"), multierr.Append(err, multierr.Append(zipWriter.Close(), zipFile.Close())))
	}

	err = multierr.Append(zipWriter.Close(), zipFile.Close())
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at close zip file"), err)
	}

	wg := sync.WaitGroup{}

	var (
		uploadErr error
		mtx       sync.Mutex
	)
	uploadPath := func(pth string) {
		defer wg.Done()
		defer func() {
			if pnk := recover(); pnk != nil {
				mtx.Lock()
				defer mtx.Unlock()
				uploadErr = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
			}
		}()

		h := sha3.New512()
		data, err := os.ReadFile(pth)
		if err != nil {
			mtx.Lock()
			defer mtx.Unlock()
			uploadErr = multierr.Append(fmt.Errorf("failed at readfile %s", pth), multierr.Append(err, uploadErr))
			return
		}
		_, err = h.Write(data)
		if err != nil {
			mtx.Lock()
			defer mtx.Unlock()
			uploadErr = multierr.Append(fmt.Errorf("failed at hash data"), multierr.Append(err, uploadErr))
			return
		}

		ctx.Inst().Prometheus.TotalBytesUploaded(len(data))

		sha3 := hex.EncodeToString(h.Sum(nil))

		t := container.Match(data)
		key := path.Join(task.Output.Prefix, path.Base(pth))
		if t == matchers.TypeZip {
			result.ZipOutput = ResultZipOutput{
				Name:         path.Base(pth),
				Size:         len(data),
				Key:          key,
				Bucket:       task.Output.Bucket,
				ACL:          task.Output.ACL,
				CacheControl: task.Output.CacheControl,
				SHA3:         sha3,
			}
		} else {
			var format ResultOutputFormatType
			switch t {
			case matchers.TypeGif:
				format = ResultOutputFormatTypeGIF
			case matchers.TypePng:
				format = ResultOutputFormatTypePNG
			case matchers.TypeWebp:
				format = ResultOutputFormatTypeWEBP
			case container.TypeAvif:
				format = ResultOutputFormatTypeAVIF
			}

			var (
				width      int
				height     int
				frameCount int
			)
			switch t {
			case matchers.TypeGif, matchers.TypePng:
				output, err := exec.CommandContext(ctx,
					"ffprobe",
					"-v", "error",
					"-select_streams", "v:0",
					"-count_packets",
					"-show_entries", "stream=width,height,nb_read_packets",
					"-of", "csv=p=0",
					pth,
				).CombinedOutput()
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at ffprobe png/gif"), multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr))
					return
				}

				splits := strings.SplitN(strings.TrimSpace(utils.B2S(output)), ",", 3)
				width, err = strconv.Atoi(splits[0])
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at parse width"), multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr))
					return
				}
				height, err = strconv.Atoi(splits[1])
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at parse height"), multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr))
					return
				}
				frameCount, err = strconv.Atoi(splits[2])
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at parse frame count"), multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr))
					return
				}
			case matchers.TypeWebp, container.TypeAvif:
				output, err := exec.CommandContext(ctx,
					"dump_png",
					"--info",
					"-i", pth,
				).CombinedOutput()
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at dump_png"), multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr))
					return
				}

				lines := strings.Split(strings.TrimSpace(utils.B2S(output)), "\n")
				splits := strings.SplitN(lines[1], ",", 3)
				width, err = strconv.Atoi(splits[0])
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at parse width"), multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr))
					return
				}
				height, err = strconv.Atoi(splits[1])
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at parse height"), multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr))
					return
				}
				frameCount, err = strconv.Atoi(splits[2])
				if err != nil {
					mtx.Lock()
					defer mtx.Unlock()
					uploadErr = multierr.Append(fmt.Errorf("failed at parse frameCount"), multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr))
					return
				}
			}

			mtx.Lock()
			defer mtx.Unlock()
			result.ImageOutputs = append(result.ImageOutputs, ResultImageOutput{
				Name:         path.Base(pth),
				Format:       format,
				FrameCount:   frameCount,
				Width:        width,
				Height:       height,
				Key:          key,
				Bucket:       task.Output.Bucket,
				Size:         len(data),
				ContentType:  t.MIME.Value,
				ACL:          task.Output.ACL,
				CacheControl: task.Output.CacheControl,
				SHA3:         sha3,
			})
		}

		if err := ctx.Inst().S3.UploadFile(ctx, &s3manager.UploadInput{
			Body:         bytes.NewReader(data),
			ACL:          aws.String(task.Output.ACL),
			Bucket:       aws.String(task.Output.Bucket),
			CacheControl: aws.String(task.Output.CacheControl),
			ContentType:  aws.String(t.MIME.Value),
			Key:          aws.String(key),
		}); err != nil {
			mtx.Lock()
			uploadErr = multierr.Append(fmt.Errorf("failed at s3 upload"), multierr.Append(err, uploadErr))
			mtx.Unlock()
		}
	}

	err = filepath.Walk(resultsDir, func(pth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		go uploadPath(pth)

		return nil
	})
	if err != nil {
		return multierr.Append(fmt.Errorf("failed at walk resultsDir"), err)
	}

	wg.Add(1)
	uploadPath(zipFilePath)

	wg.Wait()
	return uploadErr
}

func (Worker) makeResults(tmpDir string, delays []int, tsk task.Task, variantsDir string, ctx global.Context, inputDir string, inputFile string) (resultsDir string, err error) {
	// Syntax: convert_png [options] -i input.png -o output.webp -o output.gif -o output.avif
	// Options:
	//   -h,--help                   : Shows syntax help
	//   -i,--input FILENAME         : Input file location (supported types are png).
	//   -o,--output FILENAME        : Output file location (supported types are webp, avif, gif).
	//   -d,--delay D                : Delay of the next frame in 100s of a second. (default 4 = 40ms)
	// the max fps is 50fps

	defer func() {
		if pnk := recover(); pnk != nil {
			err = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
		}
	}()

	resultsDir = path.Join(tmpDir, "results")
	err = os.MkdirAll(resultsDir, 0700)
	if err != nil {
		return "", multierr.Append(fmt.Errorf("failed at mkdir resultsDir"), err)
	}

	if len(delays) > 1 {
		for _, scale := range tsk.Scales {
			convertArgs := []string{}
			for i := 0; i < len(delays); i++ {

				if delays[i] <= 1 {
					delays[i] = 2
				}

				convertArgs = append(convertArgs,
					"-d", strconv.Itoa(delays[i]),
					"-i", path.Join(variantsDir, fmt.Sprintf("%04d_%dx.png", i, scale)),
				)
			}

			outputs := 0

			if tsk.Flags&task.TaskFlagAVIF != 0 {
				convertArgs = append(convertArgs,
					"-o", path.Join(resultsDir, fmt.Sprintf("%dx.avif", scale)),
				)
				outputs++
			}

			if tsk.Flags&task.TaskFlagWEBP != 0 {
				convertArgs = append(convertArgs,
					"-o", path.Join(resultsDir, fmt.Sprintf("%dx.webp", scale)),
				)
				outputs++
			}

			madeGif := false

			if tsk.Flags&task.TaskFlagGIF != 0 {
				convertArgs = append(convertArgs,
					"-o", path.Join(resultsDir, fmt.Sprintf("%dx.gif", scale)),
				)
				madeGif = true
				outputs++
			}

			if outputs > 0 {
				out, err := exec.CommandContext(ctx,
					"convert_png",
					convertArgs...,
				).CombinedOutput()
				if err != nil {
					return "", multierr.Append(fmt.Errorf("failed at convert_png"), multierr.Append(err, fmt.Errorf("convert_png failed: %s", out)))
				}

			}
			if madeGif {
				out, err := exec.CommandContext(ctx,
					"gifsicle",
					"-O3",
					"--colors", "256",
					"-b",
					path.Join(resultsDir, fmt.Sprintf("%dx.gif", scale)),
				).CombinedOutput()
				if err != nil {
					return "", multierr.Append(fmt.Errorf("failed at gifsicle"), multierr.Append(err, fmt.Errorf("gifsicle failed: %s", out)))
				}
			}
		}
	}

	for _, scale := range tsk.Scales {
		convertArgs := []string{
			"-i", path.Join(variantsDir, fmt.Sprintf("0000_%dx.png", scale)),
		}

		static := "_static"
		if len(delays) == 1 {
			static = ""
		}

		outputs := 0

		if (tsk.Flags&task.TaskFlagAVIF_STATIC != 0 && len(delays) > 1) || (tsk.Flags&task.TaskFlagAVIF != 0 && len(delays) == 1) {
			convertArgs = append(convertArgs,
				"-o", path.Join(resultsDir, fmt.Sprintf("%dx%s.avif", scale, static)),
			)
			outputs++
		}

		if (tsk.Flags&task.TaskFlagWEBP_STATIC != 0 && len(delays) > 1) || (tsk.Flags&task.TaskFlagWEBP != 0 && len(delays) == 1) {
			convertArgs = append(convertArgs,
				"-o", path.Join(resultsDir, fmt.Sprintf("%dx%s.webp", scale, static)),
			)
			outputs++
		}

		if (tsk.Flags&task.TaskFlagPNG_STATIC != 0 && len(delays) > 1) || (tsk.Flags&task.TaskFlagPNG != 0 && len(delays) == 1) {
			if _, err := copyFile(path.Join(variantsDir, fmt.Sprintf("0000_%dx.png", scale)), path.Join(resultsDir, fmt.Sprintf("%dx%s.png", scale, static))); err != nil {
				return "", multierr.Append(fmt.Errorf("failed at copy png"), err)
			}

			out, err := exec.CommandContext(ctx,
				"optipng",
				"-o6",
				path.Join(resultsDir, fmt.Sprintf("%dx%s.png", scale, static)),
			).CombinedOutput()
			if err != nil {
				return "", multierr.Append(fmt.Errorf("failed at optipng"), multierr.Append(err, fmt.Errorf("optipng failed: %s", out)))
			}
		}

		if outputs > 0 {
			out, err := exec.CommandContext(ctx,
				"convert_png",
				convertArgs...,
			).CombinedOutput()
			if err != nil {
				return "", multierr.Append(fmt.Errorf("failed at convert_png"), multierr.Append(err, fmt.Errorf("convert_png failed: %s", out)))
			}
		}
	}

	if err = os.RemoveAll(inputDir); err != nil {
		return "", multierr.Append(fmt.Errorf("failed at rmdir inputDir"), err)
	}

	if err = os.RemoveAll(inputFile); err != nil {
		return "", multierr.Append(fmt.Errorf("failed at rmdir inputFile"), err)
	}
	return resultsDir, nil
}

func (Worker) resizeFrames(ctx global.Context, inputDir string, tmpDir string, tsk task.Task, delays []int) (variantsDir string, err error) {
	// Syntax: resize_png [options] -i input.png -r 100 100 -o out.png -r 50 50 -o out2.png
	// Options:
	//	 -h,--help                   : Shows syntax help
	//	 -i,--input FILENAME         : Input file location (supported types are png).
	//	 -r,--resize 100 100         : The width and height
	//	 -o,--output FILENAME        : Output filename (supported types are png).

	defer func() {
		if pnk := recover(); pnk != nil {
			err = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
		}
	}()

	out, err := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "v",
		"-of", "default=noprint_wrappers=1:nokey=1",
		"-show_entries", "stream=width,height",
		path.Join(inputDir, "0000.png"),
	).CombinedOutput()
	if err != nil {
		return "", multierr.Append(fmt.Errorf("failed at ffprobe"), multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out)))
	}

	widthHeight := strings.SplitN(strings.TrimSpace(utils.B2S(out)), "\n", 2)

	width, err := strconv.Atoi(widthHeight[0])
	if err != nil {
		return "", multierr.Append(fmt.Errorf("failed at parse width"), multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out)))
	}
	height, err := strconv.Atoi(widthHeight[1])
	if err != nil {
		return "", multierr.Append(fmt.Errorf("failed at parse height"), multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out)))
	}

	variantsDir = path.Join(tmpDir, "variants")
	err = os.MkdirAll(variantsDir, 0700)
	if err != nil {
		return "", multierr.Append(fmt.Errorf("failed at mkdir variantsDir"), err)
	}

	smwf := float64(tsk.SmallestMaxWidth)
	wf := float64(width)
	smhf := float64(tsk.SmallestMaxHeight)
	hf := float64(height)

	if smwf < wf {
		hf *= smwf / wf
		wf = smwf
	}

	if smhf < hf {
		wf *= smhf / hf
		hf = smhf
	}

	width = int(math.Round(wf))
	height = int(math.Round(hf))

	resizeArgs := []string{}
	for i := 0; i < len(delays); i++ {
		resizeArgs = append(resizeArgs,
			"-i", path.Join(inputDir, fmt.Sprintf("%04d.png", i)),
		)
		for _, scale := range tsk.Scales {
			height := height * scale
			width := width * scale

			resizeArgs = append(resizeArgs,
				"-r", strconv.Itoa(width), strconv.Itoa(height),
				"-o", path.Join(variantsDir, fmt.Sprintf("%04d_%dx.png", i, scale)),
			)
		}
	}

	out, err = exec.CommandContext(ctx,
		"resize_png",
		resizeArgs...,
	).CombinedOutput()
	if err != nil {
		return "", multierr.Append(fmt.Errorf("failed at resize_png"), multierr.Append(err, fmt.Errorf("resize_png failed: %s", out)))
	}

	return variantsDir, nil
}

func (Worker) exportFrames(ctx global.Context, tmpDir string, inputFile string, match types.Type, raw []byte) (delays []int, inputDir string, err error) {
	// Syntax: dump_png -i input.webp -o output
	// Options:
	//	 -h,--help                   : Shows syntax help
	//	 -i,--input FILENAME         : Input file location (supported types are webp and avif).
	//	 -o,--output FOLDER          : Output folder
	//	 --info                      : Only output info dont dump the images

	defer func() {
		if pnk := recover(); pnk != nil {
			err = multierr.Append(fmt.Errorf("panic at runtime: %v", pnk), err)
		}
	}()

	inputDir = path.Join(tmpDir, "input")
	err = os.MkdirAll(inputDir, 0700)
	if err != nil {
		return nil, "", multierr.Append(fmt.Errorf("failed at mkdir inputDir"), err)
	}

	switch match {
	case matchers.TypeWebp, container.TypeAvif:
		// we use dump_png
		out, err := exec.CommandContext(ctx,
			"dump_png",
			"-i", inputFile,
			"-o", inputDir,
		).CombinedOutput()
		if err != nil {
			return nil, "", multierr.Append(fmt.Errorf("failed at dump_png"), multierr.Append(err, fmt.Errorf("dump_png failed: %s", out)))
		}

		lines := strings.Split(utils.B2S(out), "\n")
		for _, line := range lines[3:] {
			line = strings.TrimSpace(line)
			if line != "" {
				splits := strings.SplitN(line, ",", 2)
				delay, err := strconv.Atoi(splits[1])
				if err != nil {
					return nil, "", multierr.Append(fmt.Errorf("failed at parse delay"), multierr.Append(err, fmt.Errorf("dump_png failed: %s", out)))
				}
				delays = append(delays, delay)
			}
		}
	case matchers.TypeGif,
		matchers.TypePng,
		matchers.TypeMp4,
		matchers.TypeFlv,
		matchers.TypeAvi,
		matchers.TypeMov,
		matchers.TypeJpeg,
		matchers.TypeTiff,
		matchers.TypeWebm:
		// we use ffmpeg to get the frames
		if match == matchers.TypeGif {
			// if this is a gif we need to know the per frame timings, we can use the builtin gif decoder to get this
			img, err := gif.DecodeAll(bytes.NewReader(raw))
			if err != nil {
				return nil, "", multierr.Append(fmt.Errorf("failed at gif decode"), err)
			}

			delays = img.Delay
		}

		// now we must use ffmpeg to extract all the frames of the image
		out, err := exec.CommandContext(ctx,
			"ffmpeg",
			"-v", "error",
			"-nostats",
			"-hide_banner",
			"-i", inputFile,
			"-f", "image2",
			"-start_number", "0",
			path.Join(inputDir, "%04d.png"),
		).CombinedOutput()
		if err != nil {
			return nil, "", multierr.Append(fmt.Errorf("failed at ffmpeg"), multierr.Append(err, fmt.Errorf("ffmpeg failed: %s", out)))
		}

		if len(delays) == 0 {
			files, err := ioutil.ReadDir(inputDir)
			if err != nil {
				return nil, "", multierr.Append(fmt.Errorf("failed at ReadDir inputDir"), err)
			}

			// make the array with the total number of files
			delays = make([]int, len(files))
			// we then need to get the framerate of the input if there is more than 1 file
			if len(files) > 1 {
				// ffprobe -v error -select_streams v -of default=noprint_wrappers=1:nokey=1 -show_entries stream=r_frame_rate
				out, err := exec.CommandContext(ctx,
					"ffprobe",
					"-v", "error",
					"-select_streams", "v",
					"-of", "default=noprint_wrappers=1:nokey=1",
					"-show_entries", "stream=r_frame_rate",
					inputFile,
				).CombinedOutput()
				if err != nil {
					return nil, "", multierr.Append(fmt.Errorf("failed at ffprobe"), multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out)))
				}

				fpsArr := strings.SplitN(strings.TrimSpace(utils.B2S(out)), "/", 2)
				numerator, err := strconv.Atoi(fpsArr[0])
				if err != nil {
					return nil, "", multierr.Append(fmt.Errorf("failed at parse numerator fps"), multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out)))
				}
				denominator, err := strconv.Atoi(fpsArr[1])
				if err != nil {
					return nil, "", multierr.Append(fmt.Errorf("failed at parse denominator fps"), multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out)))
				}

				// this is because GIF images can only be a max of 50fps, meaning each frame can only be 2 timescales (0.02s)
				delay := int(math.Max(math.Round(100/(float64(numerator)/float64(denominator))), 2))
				for i := range delays {
					delays[i] = delay
				}
			}
		}
	}

	return delays, inputDir, nil
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
