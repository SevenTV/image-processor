package health

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/seventv/image-processor/go/internal/configure"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
	messagequeue "github.com/seventv/message-queue/go"
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
}

func TestHealthS3RMQ(t *testing.T) {
	config := &configure.Config{}
	config.Health.Enabled = true
	config.Health.Bind = "127.0.1.1:3000"

	gCtx, cancel := global.WithCancel(global.New(context.Background(), config))

	var err error
	gCtx.Inst().S3, err = s3.NewMock(gCtx, map[string]map[string][]byte{})
	testutil.IsNil(t, err, "s3 init successful")

	gCtx.Inst().MessageQueue, err = messagequeue.New(gCtx, messagequeue.ConfigMock{})
	testutil.IsNil(t, err, "rmq init successful")

	done := New(gCtx)

	time.Sleep(time.Millisecond * 50)

	resp, err := http.DefaultClient.Get("http://127.0.1.1:3000")
	testutil.IsNil(t, err, "No error")
	_ = resp.Body.Close()
	testutil.Assert(t, http.StatusOK, resp.StatusCode, "response code all up")

	gCtx.Inst().MessageQueue.(*messagequeue.InstanceMock).SetConnected(false)

	resp, err = http.DefaultClient.Get("http://127.0.1.1:3000")
	testutil.IsNil(t, err, "No error")
	_ = resp.Body.Close()
	testutil.Assert(t, http.StatusInternalServerError, resp.StatusCode, "response code rmq down")

	gCtx.Inst().MessageQueue.(*messagequeue.InstanceMock).SetConnected(true)
	gCtx.Inst().S3.(*s3.MockInstance).SetConnected(false)

	resp, err = http.DefaultClient.Get("http://127.0.1.1:3000")
	testutil.IsNil(t, err, "No error")
	_ = resp.Body.Close()
	testutil.Assert(t, http.StatusInternalServerError, resp.StatusCode, "response code s3 down")

	cancel()

	<-done
}
