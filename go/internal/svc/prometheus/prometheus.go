package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/seventv/image-processor/go/internal/instance"
)

type Options struct {
	Labels prometheus.Labels
}

func New(o Options) instance.Prometheus {
	return &Instance{}
}

type Instance struct {
}

func (m *Instance) Register(r prometheus.Registerer) {
}
