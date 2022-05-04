package kubemq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SevenTV/Common/sync_map"
	"github.com/kubemq-io/kubemq-go"
	"github.com/seventv/image-processor/go/internal/instance"
)

type mtxArr struct {
	mtx sync.Mutex
	arr []*kubemq.QueueMessage
}

type MockInstance struct {
	msgs *sync_map.Map[string, *mtxArr]
}

type MockQueueTransactionMessageResponse struct {
	inst *MockInstance
	arr  *mtxArr
	msg  *kubemq.QueueMessage
	once sync.Once
}

func NewMock(ctx context.Context) (instance.KubeMQ, error) {
	return &MockInstance{
		msgs: &sync_map.Map[string, *mtxArr]{},
	}, nil
}

func (i *MockInstance) Send(ctx context.Context, msg *kubemq.QueueMessage) (*kubemq.SendQueueMessageResult, error) {
	val, _ := i.msgs.LoadOrStore(msg.Channel, &mtxArr{})
	val.mtx.Lock()
	val.arr = append(val.arr, msg)
	val.mtx.Unlock()
	return &kubemq.SendQueueMessageResult{}, nil
}

func (i *MockInstance) SendBatch(ctx context.Context, msgs []*kubemq.QueueMessage) ([]*kubemq.SendQueueMessageResult, error) {
	for _, msg := range msgs {
		val, _ := i.msgs.LoadOrStore(msg.Channel, &mtxArr{})
		val.mtx.Lock()
		val.arr = append(val.arr, msg)
		val.mtx.Unlock()
	}

	return make([]*kubemq.SendQueueMessageResult, len(msgs)), nil
}

func (i *MockInstance) Subscribe(ctx context.Context, channel string, cb func(response instance.QueueTransactionMessageResponse, err error)) error {
	ch, _ := i.msgs.LoadOrStore(channel, &mtxArr{})
	cbMtx := sync.Mutex{}

	go func() {
		tick := time.NewTicker(time.Millisecond * 100)
		defer tick.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
			}

			ch.mtx.Lock()
			if len(ch.arr) > 0 {
				last := ch.arr[len(ch.arr)-1]
				ch.arr = ch.arr[:len(ch.arr)-1]
				go func(last *kubemq.QueueMessage) {
					cbMtx.Lock()
					cb(&MockQueueTransactionMessageResponse{
						inst: i,
						arr:  ch,
						msg:  last,
					}, nil)
					cbMtx.Unlock()
				}(last)
			}
			ch.mtx.Unlock()
		}
	}()

	return nil
}

func (m *MockQueueTransactionMessageResponse) Ack() error {
	err := fmt.Errorf("already responded")
	m.once.Do(func() {
		err = nil
	})
	return err
}

func (m *MockQueueTransactionMessageResponse) Msg() *kubemq.QueueMessage {
	return m.msg
}

func (m *MockQueueTransactionMessageResponse) Reject() error {
	err := fmt.Errorf("already responded")
	m.once.Do(func() {
		err = nil
	})
	return err
}

func (m *MockQueueTransactionMessageResponse) ExtendVisibilitySeconds(value int) error {
	return nil
}

func (m *MockQueueTransactionMessageResponse) Resend(channel string) error {
	err := fmt.Errorf("already responded")
	m.once.Do(func() {
		_, _ = m.inst.Send(context.Background(), m.msg)
		err = nil
	})
	return err
}

func (m *MockQueueTransactionMessageResponse) ResendNewMessage(msg *kubemq.QueueMessage) error {
	err := fmt.Errorf("already responded")
	m.once.Do(func() {
		_, _ = m.inst.Send(context.Background(), msg)
		err = nil
	})

	return err
}
