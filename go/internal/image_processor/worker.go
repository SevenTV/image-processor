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

	"github.com/SevenTV/Common/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/h2non/filetype/matchers"
	"github.com/seventv/image-processor/go/internal/container"
	"github.com/seventv/image-processor/go/internal/global"
	"go.uber.org/multierr"
	"golang.org/x/crypto/sha3"
)

type Worker struct{}

func (Worker) Work(ctx global.Context, task Task, result *Result) error {
	if result == nil {
		return fmt.Errorf("nil for result")
	}

	id := uuid.New().String()
	tmpDir := path.Join(ctx.Config().Worker.TempDir, id)
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	buf := aws.NewWriteAtBuffer([]byte{})

	err := ctx.Inst().S3.DownloadFile(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(task.Input.Bucket),
		Key:    aws.String(task.Input.Key),
	})
	if err != nil {
		return err
	}

	match := container.Match(buf.Bytes())
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
		return fmt.Errorf("unsupported image format: %v", match.Extension)
	}

	inputFile := path.Join(tmpDir, fmt.Sprintf("input.%s", match.Extension))
	file, err := os.Create(inputFile)
	if err != nil {
		return err
	}

	_, err = file.Write(buf.Bytes())
	if err != nil {
		return multierr.Append(err, file.Close())
	}

	err = file.Close()
	if err != nil {
		return err
	}

	h := sha3.New512()
	_, err = h.Write(buf.Bytes())
	if err != nil {
		return err
	}

	result.InputSHA3 = hex.EncodeToString(h.Sum(nil))

	inputDir := path.Join(tmpDir, "input")
	err = os.MkdirAll(inputDir, 0700)
	if err != nil {
		return err
	}

	var delays []int
	switch match {
	case matchers.TypeWebp, container.TypeAvif:
		// we use dump_png
		out, err := exec.CommandContext(ctx,
			"dump_png",
			"-i", inputFile,
			"-o", inputDir,
		).CombinedOutput()
		if err != nil {
			return multierr.Append(err, fmt.Errorf("dump_png failed: %s", out))
		}

		lines := strings.Split(utils.B2S(out), "\n")
		for _, line := range lines[3:] {
			line = strings.TrimSpace(line)
			if line != "" {
				splits := strings.SplitN(line, ",", 2)
				delay, err := strconv.Atoi(splits[1])
				if err != nil {
					return multierr.Append(err, fmt.Errorf("dump_png failed: %s", out))
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
			img, err := gif.DecodeAll(bytes.NewReader(buf.Bytes()))
			if err != nil {
				return err
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
			return multierr.Append(err, fmt.Errorf("ffmpeg failed: %s", out))
		}

		if len(delays) == 0 {
			files, err := ioutil.ReadDir(inputDir)
			if err != nil {
				return err
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
					return multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out))
				}

				fpsArr := strings.SplitN(strings.TrimSpace(utils.B2S(out)), "/", 2)
				numerator, err := strconv.Atoi(fpsArr[0])
				if err != nil {
					return multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out))
				}
				denominator, err := strconv.Atoi(fpsArr[1])
				if err != nil {
					return multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out))
				}

				// this is because GIF images can only be a max of 50fps, meaning each frame can only be 2 timescales (0.02s)
				delay := int(math.Max(math.Round(100/(float64(numerator)/float64(denominator))), 2))
				for i := range delays {
					delays[i] = delay
				}
			}
		}
	}

	out, err := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "v",
		"-of", "default=noprint_wrappers=1:nokey=1",
		"-show_entries", "stream=width,height",
		path.Join(inputDir, "0000.png"),
	).CombinedOutput()
	if err != nil {
		return multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out))
	}

	widthHeight := strings.SplitN(strings.TrimSpace(utils.B2S(out)), "\n", 2)

	width, err := strconv.Atoi(widthHeight[0])
	if err != nil {
		return multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out))
	}
	height, err := strconv.Atoi(widthHeight[1])
	if err != nil {
		return multierr.Append(err, fmt.Errorf("ffprobe failed: %s", out))
	}

	// we then need to resize all the images
	// Syntax: resize_png [options] -i input.png -r 100 100 -o out.png -r 50 50 -o out2.png
	// Options:
	// 	-h,--help                   : Shows syntax help
	// 	-i,--input FILENAME         : Input file location (supported types are png).
	// 	-r,--resize 100 100         : The width and height(supported types are png).
	// 	-o,--output FILENAME        : Output filename(supported types are png).

	variantsDir := path.Join(tmpDir, "variants")
	err = os.MkdirAll(variantsDir, 0700)
	if err != nil {
		return err
	}

	smwf := float64(task.SmallestMaxWidth)
	wf := float64(width)
	smhf := float64(task.SmallestMaxHeight)
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
		for _, scale := range task.Scales {
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
		return multierr.Append(err, fmt.Errorf("convert_png failed: %s", out))
	}

	resultsDir := path.Join(tmpDir, "results")
	err = os.MkdirAll(resultsDir, 0700)
	if err != nil {
		return err
	}

	// Syntax: convert_png [options] -i input.png -o output.webp -o output.gif -o output.avif
	// Options:
	//   -h,--help                   : Shows syntax help
	//   -i,--input FILENAME         : Input file location (supported types are png).
	//   -o,--output FILENAME        : Output file location (supported types are webp, avif, gif).
	//   -d,--delay D                : Delay of the next frame in 100s of a second. (default 4 = 40ms)

	if len(delays) > 1 {
		for _, scale := range task.Scales {
			convertArgs := []string{}
			for i := 0; i < len(delays); i++ {
				// the max fps is 50fps
				if delays[i] <= 1 {
					delays[i] = 2
				}

				convertArgs = append(convertArgs,
					"-d", strconv.Itoa(delays[i]),
					"-i", path.Join(variantsDir, fmt.Sprintf("%04d_%dx.png", i, scale)),
				)
			}

			outputs := 0

			if task.Flags&TaskFlagAVIF != 0 {
				convertArgs = append(convertArgs,
					"-o", path.Join(resultsDir, fmt.Sprintf("%dx.avif", scale)),
				)
				outputs++
			}

			if task.Flags&TaskFlagWEBP != 0 {
				convertArgs = append(convertArgs,
					"-o", path.Join(resultsDir, fmt.Sprintf("%dx.webp", scale)),
				)
				outputs++
			}

			madeGif := false

			if task.Flags&TaskFlagGIF != 0 {
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
					return multierr.Append(err, fmt.Errorf("convert_png failed: %s", out))
				}

			}
			if madeGif {
				out, err = exec.CommandContext(ctx,
					"gifsicle",
					"-O3",
					"--colors", "256",
					"-b",
					path.Join(resultsDir, fmt.Sprintf("%dx.gif", scale)),
				).CombinedOutput()
				if err != nil {
					return multierr.Append(err, fmt.Errorf("gifsicle failed: %s", out))
				}
			}
		}
	}

	for _, scale := range task.Scales {
		convertArgs := []string{
			"-i", path.Join(variantsDir, fmt.Sprintf("0000_%dx.png", scale)),
		}

		static := "_static"
		if len(delays) == 1 {
			static = ""
		}

		outputs := 0

		if (task.Flags&TaskFlagAVIF_STATIC != 0 && len(delays) > 1) || (task.Flags&TaskFlagAVIF != 0 && len(delays) == 1) {
			convertArgs = append(convertArgs,
				"-o", path.Join(resultsDir, fmt.Sprintf("%dx%s.avif", scale, static)),
			)
			outputs++
		}

		if (task.Flags&TaskFlagWEBP_STATIC != 0 && len(delays) > 1) || (task.Flags&TaskFlagWEBP != 0 && len(delays) == 1) {
			convertArgs = append(convertArgs,
				"-o", path.Join(resultsDir, fmt.Sprintf("%dx%s.webp", scale, static)),
			)
			outputs++
		}

		if (task.Flags&TaskFlagPNG_STATIC != 0 && len(delays) > 1) || (task.Flags&TaskFlagPNG != 0 && len(delays) == 1) {
			if _, err := copyFile(path.Join(variantsDir, fmt.Sprintf("0000_%dx.png", scale)), path.Join(resultsDir, fmt.Sprintf("%dx%s.png", scale, static))); err != nil {
				return err
			}

			out, err = exec.CommandContext(ctx,
				"optipng",
				"-o6",
				path.Join(resultsDir, fmt.Sprintf("%dx%s.png", scale, static)),
			).CombinedOutput()
			if err != nil {
				return multierr.Append(err, fmt.Errorf("optipng failed: %s", out))
			}
		}

		if outputs > 0 {
			out, err := exec.CommandContext(ctx,
				"convert_png",
				convertArgs...,
			).CombinedOutput()
			if err != nil {
				return multierr.Append(err, fmt.Errorf("convert_png failed: %s", out))
			}
		}
	}

	if err = os.RemoveAll(inputDir); err != nil {
		return err
	}

	if err = os.RemoveAll(inputFile); err != nil {
		return err
	}

	zipFilePath := path.Join(tmpDir, "emote.zip")
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	w := zip.NewWriter(zipFile)

	walker := func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(pth)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(strings.TrimLeft(pth, tmpDir)[1:])
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}
	err = filepath.Walk(resultsDir, walker)
	if err != nil {
		return multierr.Append(err, multierr.Append(w.Close(), zipFile.Close()))
	}

	err = filepath.Walk(variantsDir, walker)
	if err != nil {
		return multierr.Append(err, multierr.Append(w.Close(), zipFile.Close()))
	}

	err = multierr.Append(w.Close(), zipFile.Close())
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	// we now need to upload all the files to s3
	var (
		uploadErr error
		mtx       sync.Mutex
	)
	uploadPath := func(pth string) {
		defer wg.Done()

		h := sha3.New512()
		data, err := os.ReadFile(pth)
		if err != nil {
			mtx.Lock()
			uploadErr = multierr.Append(err, uploadErr)
			mtx.Unlock()
			return
		}
		_, err = h.Write(data)
		if err != nil {
			mtx.Lock()
			uploadErr = multierr.Append(err, uploadErr)
			mtx.Unlock()
			return
		}

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
				// ffprobe -v error -select_streams v:0 -count_packets -show_entries stream=width,height,nb_read_packets -of csv=p=0 assets/animated-1.gif
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
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}

				splits := strings.SplitN(strings.TrimSpace(utils.B2S(output)), ",", 3)
				width, err = strconv.Atoi(splits[0])
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}
				height, err = strconv.Atoi(splits[1])
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}
				frameCount, err = strconv.Atoi(splits[2])
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("ffprobe failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}
			case matchers.TypeWebp, container.TypeAvif:
				// we need to do
				output, err := exec.CommandContext(ctx,
					"dump_png",
					"--info",
					"-i", pth,
					pth,
				).CombinedOutput()
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}

				lines := strings.Split(strings.TrimSpace(utils.B2S(output)), "\n")

				splits := strings.SplitN(lines[1], ",", 3)
				width, err = strconv.Atoi(splits[0])
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}
				height, err = strconv.Atoi(splits[1])
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}
				frameCount, err = strconv.Atoi(splits[2])
				if err != nil {
					mtx.Lock()
					uploadErr = multierr.Append(multierr.Append(err, fmt.Errorf("dump_png failed: %s", output)), uploadErr)
					mtx.Unlock()
					return
				}
			}

			mtx.Lock()
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
			mtx.Unlock()
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
			uploadErr = multierr.Append(err, uploadErr)
			mtx.Unlock()
			return
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
		return err
	}

	wg.Add(1)
	uploadPath(zipFilePath)

	wg.Wait()

	return ctx.Err()
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
