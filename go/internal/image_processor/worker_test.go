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
	"github.com/seventv/image-processor/go/internal/svc/kubemq"
	"github.com/seventv/image-processor/go/internal/svc/prometheus"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
)

var assets = []string{
	// "animated-1.avif",
	"animated-1.gif",
	"animated-1.png",
	"animated-1.webp",
	"animated-2.gif",
	"animated-2.webp",
	"animated-3.gif",
	"animated.avi",
	"animated.flv",
	"animated.mov",
	"animated.mp4",
	"animated.webm",
	// "static-1.avif",
	"static-1.jpeg",
	"static-1.png",
	"static-1.tiff",
	"static-1.webp",
	// "static-2.avif",
	"static-2.webp",
}

func TestWorker(t *testing.T) {
	t.Parallel()

	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))

	gCtx.Inst().KubeMQ, err = kubemq.NewMock(gCtx)
	testutil.IsNil(t, err, "kubemq init successful")
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
			result := Result{}
			err := worker.Work(gCtx, Task{
				Flags: TaskFlagALL,
				Input: TaskInput{
					Bucket: "input",
					Key:    file,
				},
				Output: TaskOutput{
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
