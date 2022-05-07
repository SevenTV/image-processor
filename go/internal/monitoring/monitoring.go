package monitoring

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/zap"
)

func New(gCtx global.Context) <-chan struct{} {
	registry := gCtx.Inst().Prometheus.Registry()

	server := fasthttp.Server{
		Handler: fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			Registry:          registry,
			EnableOpenMetrics: true,
		})),
		GetOnly:          true,
		DisableKeepalive: true,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		zap.S().Infow("Monitoring enabled",
			"bind", gCtx.Config().Monitoring.Bind,
		)
		if err := server.ListenAndServe(gCtx.Config().Monitoring.Bind); err != nil {
			zap.S().Fatalw("failed to start monitoring bind",
				"error", err,
			)
		}
	}()

	go func() {
		<-gCtx.Done()
		_ = server.Shutdown()
	}()
	return done
}
