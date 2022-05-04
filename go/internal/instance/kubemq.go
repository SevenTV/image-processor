package instance

import (
	"context"

	"github.com/kubemq-io/kubemq-go"
)

type KubeMQ interface {
	Send(ctx context.Context, msg *kubemq.QueueMessage) (*kubemq.SendQueueMessageResult, error)
	SendBatch(ctx context.Context, msgs []*kubemq.QueueMessage) ([]*kubemq.SendQueueMessageResult, error)
	Subscribe(ctx context.Context, channel string, cb func(response QueueTransactionMessageResponse, err error)) error
}

type QueueTransactionMessageResponse interface {
	Msg() *kubemq.QueueMessage
	Ack() error
	Reject() error
	ExtendVisibilitySeconds(value int) error
	Resend(channel string) error
	ResendNewMessage(msg *kubemq.QueueMessage) error
}
