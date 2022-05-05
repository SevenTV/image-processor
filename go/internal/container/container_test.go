package container

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"github.com/seventv/image-processor/go/internal/testutil"
)

type testCase struct {
	Filename     string
	Data         []byte
	ExpectedType types.Type
}

func makeCase(t *testing.T, filename string, expected types.Type) testCase {
	_, cwd, _, _ := runtime.Caller(0)
	file := path.Join(path.Dir(cwd), "..", "..", "..", "assets", filename)

	return testCase{
		Filename:     filename,
		Data:         testutil.ReadFile(t, file),
		ExpectedType: expected,
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()

	cases := []testCase{
		makeCase(t, "animated-1.avif", TypeAvif),
		makeCase(t, "animated-1.gif", matchers.TypeGif),
		makeCase(t, "animated-1.webp", matchers.TypeWebp),
		makeCase(t, "animated-2.gif", matchers.TypeGif),
		makeCase(t, "animated-2.webp", matchers.TypeWebp),
		makeCase(t, "animated-3.gif", matchers.TypeGif),
		makeCase(t, "static-1.avif", TypeAvif),
		makeCase(t, "static-1.png", matchers.TypePng),
		makeCase(t, "static-1.webp", matchers.TypeWebp),
		makeCase(t, "static-2.avif", TypeAvif),
		makeCase(t, "static-2.webp", matchers.TypeWebp),
		makeCase(t, "animated.mp4", matchers.TypeMp4),
		makeCase(t, "animated.flv", matchers.TypeFlv),
		makeCase(t, "animated.avi", matchers.TypeAvi),
		makeCase(t, "animated.mov", matchers.TypeMov),
		makeCase(t, "static-1.jpeg", matchers.TypeJpeg),
		makeCase(t, "animated-1.png", matchers.TypePng),
		makeCase(t, "static-1.tiff", matchers.TypeTiff),
		makeCase(t, "animated.webm", matchers.TypeWebm),
	}

	for _, c := range cases {
		c := c
		t.Run(c.Filename, func(t *testing.T) {
			t.Parallel()

			testutil.Assert(t, c.ExpectedType, Match(c.Data), fmt.Sprintf("image %s", c.Filename))
		})
	}
}
