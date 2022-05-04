package image_processor

import (
	"context"
	"encoding/json"
	"path"
	"runtime"
	"testing"

	kubemqgo "github.com/kubemq-io/kubemq-go"
	"github.com/seventv/image-processor/go/internal/configure"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/instance"
	"github.com/seventv/image-processor/go/internal/svc/kubemq"
	"github.com/seventv/image-processor/go/internal/svc/prometheus"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
)

func TestRun(t *testing.T) {
	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{}))
	defer cancel()

	gCtx.Inst().KubeMQ, err = kubemq.NewMock(gCtx)
	testutil.IsNil(t, err, "kubemq init successful")

	gCtx.Inst().Prometheus = prometheus.New(prometheus.Options{})

	_, cwd, _, _ := runtime.Caller(0)
	assetDir := path.Join(path.Dir(cwd), "..", "..", "..", "assets")

	gCtx.Inst().S3, err = s3.NewMock(gCtx, map[string]map[string][]byte{
		"input": {
			"static-1.png": testutil.ReadFile(t, path.Join(assetDir, "static-1.png")),
		},
		"output": {},
	})
	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	const TaskID = "batchest-test-123"

	task, err := json.Marshal(Task{
		ID: TaskID,
		Input: TaskInput{
			Bucket: "input",
			Key:    "static-1.png",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "",
		},
	})
	testutil.IsNil(t, err, "task marshals")

	_, err = gCtx.Inst().KubeMQ.Send(gCtx, kubemqgo.NewQueueMessage().
		SetChannel("image-processor-jobs").
		SetBody(task),
	)
	testutil.IsNil(t, err, "We send a queue message")

	done := make(chan struct{})
	err = gCtx.Inst().KubeMQ.Subscribe(
		gCtx,
		"image-processor-results",
		func(response instance.QueueTransactionMessageResponse, err error) {
			defer close(done)

			result := Result{}
			testutil.IsNil(t, json.Unmarshal(response.Msg().Body, &result), "The response is a result")

			testutil.Assert(t, TaskID, result.ID, "The result is for the task we sent")

			testutil.Assert(t, "", result.Message, "No message was returned")

			testutil.Assert(t, ResultStateSuccess, result.State, "The job processed successfully")
		},
	)
	testutil.IsNil(t, err, "We can subscribe to a channel")

	<-done
}
