package image_processor

import (
	"context"
	"encoding/json"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/seventv/image-processor/go/internal/configure"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/instance"
	"github.com/seventv/image-processor/go/internal/svc/prometheus"
	"github.com/seventv/image-processor/go/internal/svc/rmq"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
	"github.com/streadway/amqp"
)

func TestRun(t *testing.T) {
	t.Parallel()

	var err error
	gCtx, cancel := global.WithCancel(global.New(context.Background(), &configure.Config{
		RMQ: struct {
			URI       string "mapstructure:\"uri\" json:\"uri\""
			JobsQueue string "mapstructure:\"jobs_queue\" json:\"jobs_queue\""
		}{
			JobsQueue: "image-processor-jobs",
		},
	}))
	defer cancel()

	gCtx.Inst().RMQ, err = rmq.NewMock()
	testutil.IsNil(t, err, "kubemq init successful")

	gCtx.Inst().Prometheus = prometheus.New(prometheus.Options{})

	_, cwd, _, _ := runtime.Caller(0)
	assetDir := path.Join(path.Dir(cwd), "..", "..", "..", "assets")

	gCtx.Inst().S3, err = s3.NewMock(gCtx, map[string]map[string][]byte{
		"input": {
			"animated-2.gif": testutil.ReadFile(t, path.Join(assetDir, "animated-2.gif")),
		},
		"output": {},
	})
	testutil.IsNil(t, err, "s3 init successful")

	Run(gCtx)

	const TaskID = "batchest-test-123"
	const CallbackEvent = "image-processor-results"

	task, err := json.Marshal(Task{
		ID:    TaskID,
		Flags: TaskFlagALL,
		Input: TaskInput{
			Bucket: "input",
			Key:    "animated-2.gif",
		},
		Output: TaskOutput{
			Bucket: "output",
			Prefix: "output",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	})
	testutil.IsNil(t, err, "task marshals")

	err = gCtx.Inst().RMQ.Publish(instance.RmqPublishRequest{
		RoutingKey: "image-processor-jobs",
		Immediate:  true,
		Mandatory:  true,
		Publishing: amqp.Publishing{
			ContentType:  "application/json",
			ReplyTo:      CallbackEvent,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         task,
		},
	})
	testutil.IsNil(t, err, "We send a queue message")

	ch, err := gCtx.Inst().RMQ.Subscribe(gCtx, instance.RmqSubscribeRequest{
		Queue: CallbackEvent,
	})
	testutil.IsNil(t, err, "We can subscribe to a channel")

	msg := <-ch

	result := Result{}
	testutil.IsNil(t, json.Unmarshal(msg.Body, &result), "The response is a result")

	testutil.Assert(t, TaskID, result.ID, "The result is for the task we sent")

	testutil.Assert(t, "", result.Message, "No message was returned")

	testutil.Assert(t, ResultStateSuccess, result.State, "The job processed successfully")
}
