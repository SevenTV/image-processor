package global

import "github.com/seventv/image-processor/go/internal/instance"

type Instances struct {
	KubeMQ     instance.KubeMQ
	S3         instance.S3
	Prometheus instance.Prometheus
}
