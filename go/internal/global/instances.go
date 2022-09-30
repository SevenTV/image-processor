package global

import (
	"github.com/seventv/common/svc/s3"
	"github.com/seventv/image-processor/go/internal/instance"
	messagequeue "github.com/seventv/message-queue/go"
)

type Instances struct {
	MessageQueue messagequeue.Instance
	S3           s3.Instance
	Prometheus   instance.Prometheus
}
