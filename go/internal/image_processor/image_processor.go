package image_processor

import (
	"encoding/json"
	"runtime"
	"time"

	"github.com/kubemq-io/kubemq-go"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/instance"
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

	if err := gCtx.Inst().KubeMQ.Subscribe(gCtx, "image-processor-jobs", func(response instance.QueueTransactionMessageResponse, err error) {
		if err != nil {
			zap.S().Warnw("failed to get message",
				"error", err,
			)
			return
		}
		msg := response.Msg()
		zap.S().Infow("new message",
			"id", msg.MessageID,
		)

		t := Task{}
		if err := json.Unmarshal(msg.Body, &t); err != nil {
			zap.S().Warnw("bad task payload",
				"error", multierr.Append(err, response.Ack()),
			)
			return
		}

		worker := <-workers

		ctx, cancel := global.WithCancel(gCtx)
		go func() {
			tick := time.NewTicker(time.Second * 15)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-tick.C:
				}
				if err := response.ExtendVisibilitySeconds(60); err != nil {
					zap.S().Errorw("failed to extend task",
						"error", err,
					)
					cancel()
				}
			}
		}()
		go func() {
			defer func() {
				cancel()
				blockers <- struct{}{}
			}()
			result := Result{
				ID:    t.ID,
				State: ResultStateFailed,
			}

			err := worker.Work(ctx, t, &result)
			if err != nil {
				err = multierr.Append(err, response.Reject())
				zap.S().Errorw("task processing failed",
					"error", err,
				)
			} else if err = response.Ack(); err != nil {
				zap.S().Errorw("failed to ack task",
					"error", err,
				)
			} else {
				result.State = ResultStateSuccess
			}

			if err != nil {
				result.Message = err.Error()
			}

			resultData, err := json.Marshal(result)
			if err != nil {
				zap.S().Errorw("failed to marshal result",
					"error", err,
				)
			} else {
				if _, err := gCtx.Inst().KubeMQ.Send(ctx, kubemq.NewQueueMessage().
					SetChannel("image-processor-results").
					SetBody(resultData),
				); err != nil {
					zap.S().Errorw("failed to publish result",
						"error", err,
					)
				}
			}
		}()

		<-blockers
	}); err != nil {
		zap.S().Fatalw("failed to start image processor",
			"error", err,
		)
	}

	zap.S().Infof("Starting job worker with %d jobs", jobCount)
}
