package health

import (
	"context"
	"time"

	"github.com/seventv/image-processor/go/internal/global"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func New(gCtx global.Context) <-chan struct{} {
	done := make(chan struct{})

	srv := fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			defer func() {
				if err := recover(); err != nil {
					zap.S().Errorw("panic in health",
						"panic", err,
					)
				}
			}()

			s3Down := false

			lCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			mqDown := gCtx.Inst().MessageQueue != nil && !gCtx.Inst().MessageQueue.Connected(lCtx)
			cancel()
			if mqDown {
				zap.S().Warnw("mq is not connected")
			}

			if gCtx.Inst().S3 != nil {
				lCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				if _, err := gCtx.Inst().S3.ListBuckets(lCtx); err != nil {
					s3Down = true
					zap.S().Warnw("s3 is not responding",
						"error", err,
					)
				}
				cancel()
			}

			if mqDown || s3Down {
				ctx.SetStatusCode(500)
			}
		},
	}

	go func() {
		defer close(done)
		zap.S().Infow("Health enabled",
			"bind", gCtx.Config().Health.Bind,
		)

		if err := srv.ListenAndServe(gCtx.Config().Health.Bind); err != nil {
			zap.S().Fatalw("failed to bind health",
				"error", err,
			)
		}
	}()

	go func() {
		<-gCtx.Done()

		_ = srv.Shutdown()
	}()

	return done
}
