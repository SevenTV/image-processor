package health

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/seventv/image-processor/go/internal/configure"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/testutil"
)

func TestHealth(t *testing.T) {
	config := &configure.Config{}
	config.Health.Enabled = true
	config.Health.Bind = "127.0.1.1:3000"

	gCtx, cancel := global.WithCancel(global.New(context.Background(), config))

	done := New(gCtx)

	time.Sleep(time.Millisecond * 50)

	resp, err := http.DefaultClient.Get("http://127.0.1.1:3000")
	testutil.IsNil(t, err, "No error")
	_ = resp.Body.Close()
	testutil.Assert(t, http.StatusOK, resp.StatusCode, "response code")

	cancel()

	<-done

	time.Sleep(time.Second)
}
