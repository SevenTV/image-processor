package global

import (
	"github.com/seventv/image-processor/go/internal/instance"
	messagequeue "github.com/seventv/message-queue/go"
)

type Instances struct {
	MessageQueue messagequeue.Instance
	S3           instance.S3
	Prometheus   instance.Prometheus
}
