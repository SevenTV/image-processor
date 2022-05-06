package health

import (
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func New(gCtx global.Context) <-chan struct{} {
	done := make(chan struct{})

	srv := fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(200)
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
