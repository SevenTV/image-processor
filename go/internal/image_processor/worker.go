package image_processor

import (
	"fmt"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/h2non/filetype/matchers"
	"github.com/seventv/image-processor/go/internal/container"
	"github.com/seventv/image-processor/go/internal/global"
	"go.uber.org/multierr"
)

type Worker struct{}

func (r *Worker) Work(ctx global.Context, task Task, result *Result) error {
	id := uuid.New().String()
	tmpDir := path.Join(ctx.Config().Worker.TempDir, id)
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return err
	}

	defer os.RemoveAll(path.Join(ctx.Config().Worker.TempDir, id))

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
	case matchers.TypeWebp:
		// we use webpdump
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
	case container.TypeAvif:
		// we use avifdump
	default:
		return fmt.Errorf("unsupported image format: %v", match.Extension)
	}

	inputFile := path.Join(tmpDir, path.Base(task.Input.Key))
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

	return ctx.Err()
}
