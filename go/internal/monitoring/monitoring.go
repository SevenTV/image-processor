package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/zap"
)

func New(gCtx global.Context) <-chan struct{} {
	gCtx.Inst().Prometheus.Register(prometheus.DefaultRegisterer)

	server := fasthttp.Server{
		Handler: fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			Registry:          prometheus.DefaultRegisterer,
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
