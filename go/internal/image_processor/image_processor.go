package image_processor

import (
	"encoding/json"
	"runtime"
	"time"

	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/instance"
	"github.com/streadway/amqp"
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

	ch, err := gCtx.Inst().RMQ.Subscribe(gCtx, instance.RmqSubscribeRequest{
		Queue: gCtx.Config().RMQ.JobsQueue,
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

func process(gCtx global.Context, msg amqp.Delivery, workers chan Worker, blockers chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorw("panic in process",
				"panic", err,
			)
		}
	}()
	t := Task{}
	if err := json.Unmarshal(msg.Body, &t); err != nil {
		zap.S().Warnw("bad task payload",
			"error", multierr.Append(err, msg.Ack(false)),
		)
		return
	}

	zap.S().Infow("new message",
		"id", t.ID,
		"msg_id", msg.MessageId,
	)

	worker := <-workers

	ctx, cancel := global.WithCancel(gCtx)
	result := Result{
		ID:    t.ID,
		State: ResultStateFailed,
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
				"msg_id", msg.MessageId,
				"run_duration", result.FinishedAt.Sub(result.StartedAt),
				"state", result.State,
				"message", result.Message,
			)
		}()

		err := worker.Work(ctx, t, &result)
		if err != nil {
			err = multierr.Append(err, msg.Reject(false))
			zap.S().Errorw("task processing failed",
				"error", err,
			)
		} else if err = msg.Ack(false); err != nil {
			zap.S().Errorw("failed to ack task",
				"error", err,
			)
		} else {
			result.State = ResultStateSuccess
		}

		if err != nil {
			result.Message = err.Error()
		}

		if msg.ReplyTo != "" {
			resultData, err := json.Marshal(result)
			if err != nil {
				zap.S().Errorw("failed to marshal result",
					"error", err,
				)
			} else {
				if err := gCtx.Inst().RMQ.Publish(instance.RmqPublishRequest{
					RoutingKey: msg.ReplyTo,
					Publishing: amqp.Publishing{
						ContentType:  "application/json",
						DeliveryMode: amqp.Persistent,
						Body:         resultData,
						Timestamp:    time.Now(),
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
