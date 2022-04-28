package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/seventv/image-processor/go/internal/global"
	"github.com/seventv/image-processor/go/internal/instance"
)

type mon struct{}

func (m *mon) Register(r prometheus.Registerer) {
}

// func labelsFromKeyValue(kv []configure.KeyValue) prometheus.Labels {
// 	mp := prometheus.Labels{}

// 	for _, v := range kv {
// 		mp[v.Key] = v.Value
// 	}

// 	return mp
// }

func NewPrometheus(gCtx global.Context) instance.Monitoring {
	return &mon{}
}
