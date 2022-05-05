package image_processor

import (
	"context"
	"path"
	"runtime"
	"testing"

	"github.com/seventv/image-processor/go/internal/configure"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/svc/kubemq"
	"github.com/seventv/image-processor/go/internal/svc/prometheus"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
)

func setup(t *testing.T, gCtx global.Context, files ...string) {
	var err error
	gCtx.Inst().KubeMQ, err = kubemq.NewMock(gCtx)
	testutil.IsNil(t, err, "kubemq init successful")
	gCtx.Inst().Prometheus = prometheus.New(prometheus.Options{})

	_, cwd, _, _ := runtime.Caller(0)
	assetDir := path.Join(path.Dir(cwd), "..", "..", "..", "assets")

	f := map[string]map[string][]byte{
		"input":  {},
		"output": {},
	}

	for _, file := range files {
		f["input"][file] = testutil.ReadFile(t, path.Join(assetDir, file))
	}

	gCtx.Inst().S3, err = s3.NewMock(gCtx, f)
	testutil.IsNil(t, err, "s3 init successful")
}

// func TestAnimatedAvif1(t *testing.T) {
// 	var err error
// 	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
// 	defer cancel()
// 	setup(t, gCtx, "animated-1.avif")

// 	testutil.IsNil(t, err, "s3 init successful")

// 	Run(gCtx)

// 	worker := Worker{}
// 	result := Result{}
// 	err = worker.Work(gCtx, Task{
// 		Flags: TaskFlagALL,
// 		Input: TaskInput{
// 			Bucket: "input",
// 			Key:    "animated-1.avif",
// 		},
// 		Output: TaskOutput{
// 			Bucket: "output",
// 			Prefix: "",
// 		},
// 		SmallestMaxWidth:  96,
// 		SmallestMaxHeight: 32,
// 		Scales:            []int{1, 2, 3, 4},
// 	}, &result)
// 	testutil.IsNil(t, err, "Convert was successful")
// }

func TestAnimatedGif1(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()
	setup(t, gCtx, "animated-1.gif")

	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	worker := Worker{}
	result := Result{}
	err = worker.Work(gCtx, Task{
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-1.gif",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	}, &result)
	testutil.IsNil(t, err, "Convert was successful")
}

func TestAnimatedPng1(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()
	setup(t, gCtx, "animated-1.png")

	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	worker := Worker{}
	result := Result{}
	err = worker.Work(gCtx, Task{
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-1.png",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	}, &result)
	testutil.IsNil(t, err, "Convert was successful")
}

func TestAnimatedWebp1(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()
	setup(t, gCtx, "animated-1.webp")

	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	worker := Worker{}
	result := Result{}
	err = worker.Work(gCtx, Task{
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-1.webp",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	}, &result)
	testutil.IsNil(t, err, "Convert was successful")
}

func TestAnimatedGif2(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()
	setup(t, gCtx, "animated-2.gif")

	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	worker := Worker{}
	result := Result{}
	err = worker.Work(gCtx, Task{
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-2.gif",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	}, &result)
	testutil.IsNil(t, err, "Convert was successful")
}

func TestAnimatedWebp2(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()
	setup(t, gCtx, "animated-2.webp")

	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	worker := Worker{}
	result := Result{}
	err = worker.Work(gCtx, Task{
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-2.webp",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	}, &result)
	testutil.IsNil(t, err, "Convert was successful")
}

func TestAnimatedGif3(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()
	setup(t, gCtx, "animated-3.gif")

	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	worker := Worker{}
	result := Result{}
	err = worker.Work(gCtx, Task{
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-3.gif",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	}, &result)
	testutil.IsNil(t, err, "Convert was successful")
}
