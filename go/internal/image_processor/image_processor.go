package image_processor

import (
	"context"
	"encoding/json"
	"runtime"
	"time"

	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/task"
	messagequeue "github.com/seventv/message-queue/go"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func Run(gCtx global.Context) {
	jobCount := gCtx.Config().Worker.Jobs
	if jobCount <= 0 {
		jobCount = runtime.GOMAXPROCS(0)
	}

	workers := make(chan Worker, jobCount)
	blockers := make(chan struct{}, jobCount-1)

	for i := 0; i < jobCount; i++ {
		workers <- Worker{}

		if i != 0 {
			blockers <- struct{}{}
		}
	}

	go func() {
		first := true
		for gCtx.Err() == nil {
			if !first {
				time.Sleep(time.Second * 5)
			} else {
				first = false
			}

			retryProcess(gCtx, workers, blockers)
		}
	}()

	zap.S().Infof("Starting job worker with %d jobs", jobCount)
}

func retryProcess(gCtx global.Context, workers chan Worker, blockers chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorw("panic in process",
				"panic", err,
			)
		}
	}()

	ch, err := gCtx.Inst().MessageQueue.Subscribe(gCtx, messagequeue.Subscription{
		Queue: gCtx.Config().MessageQueue.JobsQueue,
		RMQ:   messagequeue.SubscriptionRMQ{},
		SQS: messagequeue.SubscriptionSQS{
			WaitTimeSeconds: 10,
		},
	})
	if err != nil {
		zap.S().Fatalw("failed to start image processor",
			"error", err,
		)

		return
	}

	for {
		select {
		case <-gCtx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			process(gCtx, msg, workers, blockers)
		}
	}
}

func process(gCtx global.Context, msg *messagequeue.IncomingMessage, workers chan Worker, blockers chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorw("panic in process",
				"panic", err,
			)
		}
	}()

	t := task.Task{}
	headers := msg.Headers()

	if headers.ContentType() == "application/json" {
		if err := json.Unmarshal(msg.Body(), &t); err != nil {
			zap.S().Warnw("bad task payload",
				"error", multierr.Append(err, msg.Ack(gCtx)),
			)

			return
		}
	} else {
		zap.S().Warnw("bad task content-type",
			"error", msg.Ack(gCtx),
			"content-type", headers.ContentType(),
		)

		return
	}

	zap.S().Infow("new message",
		"id", t.ID,
		"msg_id", msg.ID(),
	)

	worker := <-workers

	var (
		ctx    global.Context
		cancel context.CancelFunc
	)

	if t.Limits.MaxProcessingTime != 0 {
		ctx, cancel = global.WithTimeout(gCtx, t.Limits.MaxProcessingTime)
	} else {
		ctx, cancel = global.WithCancel(gCtx)
	}

	result := task.Result{
		ID:    t.ID,
		State: task.ResultStateFailed,
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorw("panic in process",
					"panic", err,
				)
			}

			cancel()
			blockers <- struct{}{}
			workers <- worker

			zap.S().Infow("finished",
				"id", t.ID,
				"msg_id", msg.ID(),
				"run_duration", result.FinishedAt.Sub(result.StartedAt),
				"state", result.State,
				"message", result.Message,
			)
		}()

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 15):
					if err := msg.Extend(ctx, time.Second*30); err != nil && err != messagequeue.ErrUnimplemented {
						zap.S().Errorw("task failed to extend lease",
							"error", err,
						)
						cancel()
					}
				}
			}
		}()

		err := worker.Work(ctx, t, &result)
		if err != nil {
			err = multierr.Append(err, msg.Nack(gCtx))
			zap.S().Errorw("task processing failed",
				"error", err,
			)
		} else if err = msg.Ack(gCtx); err != nil {
			zap.S().Errorw("failed to ack task",
				"error", err,
			)
		} else {
			result.State = task.ResultStateSuccess
		}

		if err != nil {
			result.Message = err.Error()
		}

		if headers.ReplyTo() != "" {
			resultData, err := json.Marshal(result)
			if err != nil {
				zap.S().Errorw("failed to marshal result",
					"error", err,
				)
			} else {
				if err := gCtx.Inst().MessageQueue.Publish(ctx, messagequeue.OutgoingMessage{
					Queue: headers.ReplyTo(),
					Body:  resultData,
					Flags: messagequeue.MessageFlags{
						ContentType: "application/json",
						Timestamp:   time.Now(),
						RMQ: messagequeue.MessageFlagsRMQ{
							DeliveryMode: messagequeue.RMQDeliveryModePersistent,
						},
						SQS: messagequeue.MessageFlagsSQS{},
					},
				}); err != nil {
					zap.S().Errorw("failed to publish result",
						"error", err,
					)
				}
			}
		}
	}()

	<-blockers
}
