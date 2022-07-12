package image_processor

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"sync"
	"testing"

	"github.com/seventv/image-processor/go/internal/configure"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/svc/prometheus"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
	"github.com/seventv/image-processor/go/task"
	messagequeue "github.com/seventv/message-queue/go"
)

var assets = []string{
	"animated-1.avif",
	"animated-1.gif",
	"animated-1.png",
	"animated-1.webp",
	"animated-2.gif",
	"animated-2.webp",
	"animated-3.gif",
	"animated-4.gif",
	"animated.avi",
	"animated.flv",
	"animated.mov",
	"animated.mp4",
	"animated.webm",
	"static-1.avif",
	"static-1.jpeg",
	"static-1.png",
	"static-2.png",
	"static-3.png",
	"static-1.tiff",
	"static-1.webp",
	"static-2.avif",
	"static-2.webp",
}

func TestWorker(t *testing.T) {
	t.Parallel()

	var err error

	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))

	gCtx.Inst().MessageQueue, err = messagequeue.New(gCtx, messagequeue.ConfigMock{})
	testutil.IsNil(t, err, "mq init successful")

	gCtx.Inst().Prometheus = prometheus.New(prometheus.Options{})

	_, cwd, _, _ := runtime.Caller(0)
	assetDir := path.Join(path.Dir(cwd), "..", "..", "..", "assets")

	f := map[string]map[string][]byte{
		"input":  {},
		"output": {},
	}

	for _, file := range assets {
		f["input"][file] = testutil.ReadFile(t, path.Join(assetDir, file))
	}

	gCtx.Inst().S3, err = s3.NewMock(gCtx, f)
	testutil.IsNil(t, err, "s3 init successful")

	wg := sync.WaitGroup{}
	wg.Add(len(assets))

	for _, file := range assets {
		file := file
		t.Run(fmt.Sprintf("test %s", file), func(t *testing.T) {
			t.Parallel()

			defer wg.Done()

			worker := Worker{}
			result := task.Result{}
			err := worker.Work(gCtx, task.Task{
				Flags: task.TaskFlagALL,
				Input: task.TaskInput{
					Bucket: "input",
					Key:    file,
				},
				Output: task.TaskOutput{
					Bucket: "output",
					Prefix: file,
				},
				SmallestMaxWidth:  96,
				SmallestMaxHeight: 32,
				Scales:            []int{1, 2, 3, 4},
			}, &result)
			testutil.IsNil(t, err, "Convert was successful")
		})
	}

	go func() {
		wg.Wait()
		cancel()
	}()
}

func TestWorkerFailed(t *testing.T) {
	t.Parallel()

	var err error

	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))

	gCtx.Inst().MessageQueue, err = messagequeue.New(gCtx, messagequeue.ConfigMock{})
	testutil.IsNil(t, err, "mq init successful")

	gCtx.Inst().Prometheus = prometheus.New(prometheus.Options{})

	f := map[string]map[string][]byte{
		"input":  {},
		"output": {},
	}

	gCtx.Inst().S3, err = s3.NewMock(gCtx, f)
	testutil.IsNil(t, err, "s3 init successful")

	wg := sync.WaitGroup{}
	wg.Add(len(assets))

	for _, file := range assets {
		file := file
		t.Run(fmt.Sprintf("test %s", file), func(t *testing.T) {
			t.Parallel()

			defer wg.Done()

			worker := Worker{}
			result := task.Result{}
			err := worker.Work(gCtx, task.Task{
				Flags: task.TaskFlagALL,
				Input: task.TaskInput{
					Bucket: "input",
					Key:    file,
				},
				Output: task.TaskOutput{
					Bucket: "output",
					Prefix: file,
				},
				SmallestMaxWidth:  96,
				SmallestMaxHeight: 32,
				Scales:            []int{1, 2, 3, 4},
			}, &result)
			testutil.IsNotNil(t, err, "Convert was unsuccessful successful")
		})
	}

	go func() {
		wg.Wait()
		cancel()
	}()
}
