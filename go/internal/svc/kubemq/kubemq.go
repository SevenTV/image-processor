package kubemq

import (
	"context"

	"github.com/kubemq-io/kubemq-go"
	"github.com/seventv/image-processor/go/internal/instance"
	"go.uber.org/multierr"
)

type Instance struct {
	sender  *kubemq.QueuesClient
	options Options
}

type QueueTransactionMessageResponse struct {
	*kubemq.QueueTransactionMessageResponse
}

func (q QueueTransactionMessageResponse) Msg() *kubemq.QueueMessage {
	return q.QueueTransactionMessageResponse.Message
}

func new_cl(ctx context.Context, o Options) (*kubemq.QueuesClient, error) {
	return kubemq.NewQueuesStreamClient(ctx,
		kubemq.WithAddress(o.Host, o.Port),
		kubemq.WithClientId(o.ClientId),
		kubemq.WithTransportType(kubemq.TransportTypeGRPC),
		kubemq.WithAuthToken(o.AuthToken),
		kubemq.WithAutoReconnect(true),
	)
}

func New(ctx context.Context, o Options) (instance.KubeMQ, error) {
	sender, err := new_cl(ctx, o)
	if err != nil {
		return nil, err
	}

	return &Instance{
		sender:  sender,
		options: o,
	}, nil
}

func (i *Instance) Send(ctx context.Context, msg *kubemq.QueueMessage) (*kubemq.SendQueueMessageResult, error) {
	return i.sender.Send(ctx, msg)
}

func (i *Instance) SendBatch(ctx context.Context, msgs []*kubemq.QueueMessage) ([]*kubemq.SendQueueMessageResult, error) {
	return i.sender.Batch(ctx, msgs)
}

func (i *Instance) Subscribe(ctx context.Context, channel string, cb func(response instance.QueueTransactionMessageResponse, err error)) error {
	recv, err := new_cl(ctx, i.options)
	if err != nil {
		return err
	}

	done, err := recv.TransactionStream(
		ctx,
		kubemq.NewQueueTransactionMessageRequest().
			SetChannel(channel).
			SetClientId(i.options.ClientId).
			SetWaitTimeSeconds(10).
			SetVisibilitySeconds(60),
		func(response *kubemq.QueueTransactionMessageResponse, err error) {
			if err != nil && err.Error() == "Error 138: no new message in queue, wait time expired" {
				return
			}

			cb(QueueTransactionMessageResponse{response}, err)
		},
	)
	if err != nil {
		return multierr.Append(err, recv.Close())
	}

	go func() {
		<-ctx.Done()
		done <- struct{}{}

		_ = recv.Close()
	}()

	return nil
}
