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
	"github.com/seventv/image-processor/go/internal/svc/prometheus"
	"github.com/seventv/image-processor/go/internal/svc/s3"
	"github.com/seventv/image-processor/go/internal/testutil"
	"github.com/seventv/image-processor/go/task"
	messagequeue "github.com/seventv/message-queue/go"
)

func TestRun(t *testing.T) {
	t.Parallel()

	const jobQueue = "image-processor-jobs"

	var err error

	config := configure.Config{}
	config.MessageQueue.JobsQueue = jobQueue

	gCtx, cancel := global.WithCancel(global.New(context.Background(), &config))
	defer cancel()

	gCtx.Inst().MessageQueue, err = messagequeue.New(gCtx, messagequeue.ConfigMock{})
	testutil.IsNil(t, err, "mq init successful")

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

	const (
		TaskID        = "batchest-test-123"
		CallbackEvent = "image-processor-results"
	)

	tsk, err := json.Marshal(task.Task{
		ID:    TaskID,
		Flags: task.TaskFlagALL,
		Input: task.TaskInput{
			Bucket: "input",
			Key:    "animated-2.gif",
		},
		Output: task.TaskOutput{
			Bucket: "output",
			Prefix: "output",
		},
		SmallestMaxWidth:  96,
		SmallestMaxHeight: 32,
		Scales:            []int{1, 2, 3, 4},
	})
	testutil.IsNil(t, err, "task marshals")

	headers := messagequeue.MessageHeaders{}
	headers.SetContentType("application/json")
	headers.SetReplyTo(CallbackEvent)
	headers.SetTimestamp(time.Now())

	err = gCtx.Inst().MessageQueue.Publish(gCtx, messagequeue.OutgoingMessage{
		Queue:   jobQueue,
		Flags:   messagequeue.MessageFlags{},
		Body:    tsk,
		Headers: headers,
	})
	testutil.IsNil(t, err, "We send a queue message")

	ch, err := gCtx.Inst().MessageQueue.Subscribe(gCtx, messagequeue.Subscription{
		Queue: CallbackEvent,
	})
	testutil.IsNil(t, err, "We can subscribe to a channel")

	msg := <-ch

	result := task.Result{}
	testutil.IsNil(t, json.Unmarshal(msg.Body(), &result), "The response is a result")

	testutil.Assert(t, TaskID, result.ID, "The result is for the task we sent")

	testutil.Assert(t, "", result.Message, "No message was returned")

	testutil.Assert(t, task.ResultStateSuccess, result.State, "The job processed successfully")

	mock, _ := gCtx.Inst().MessageQueue.(*messagequeue.InstanceMock)
	mock.SetConnected(false)

	time.Sleep(time.Second)

	mock.SetConnected(true)

	err = gCtx.Inst().MessageQueue.Publish(gCtx, messagequeue.OutgoingMessage{
		Queue:   jobQueue,
		Flags:   messagequeue.MessageFlags{},
		Body:    tsk,
		Headers: headers,
	})
	testutil.IsNil(t, err, "We send a queue message")

	ch, err = gCtx.Inst().MessageQueue.Subscribe(gCtx, messagequeue.Subscription{
		Queue: CallbackEvent,
	})
	testutil.IsNil(t, err, "We can subscribe to a channel")

	msg = <-ch

	result = task.Result{}
	testutil.IsNil(t, json.Unmarshal(msg.Body(), &result), "The response is a result")

	testutil.Assert(t, TaskID, result.ID, "The result is for the task we sent")

	testutil.Assert(t, "", result.Message, "No message was returned")

	testutil.Assert(t, task.ResultStateSuccess, result.State, "The job processed successfully")
}
