package nats

import (
	"context"
	"strconv"

	natsclient "github.com/nats-io/nats.go"

	"cyber-ecosystem/shared-go/mq"
)

type publisher struct{ h *handle }

func newPublisher(h *handle) mq.Publisher { return &publisher{h: h} }

func (p *publisher) Publish(ctx context.Context, topic string, msg *mq.Message) (string, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return "", err
	}
	if err := p.h.ensureStream(ctx, topic); err != nil {
		return "", mapError(err, "ensure stream")
	}
	nmsg := &natsclient.Msg{Subject: topic, Data: msg.Payload}
	if len(msg.Headers) > 0 {
		hdr := natsclient.Header{}
		for k, v := range msg.Headers {
			hdr.Set(k, v)
		}
		nmsg.Header = hdr
	}
	ack, err := p.h.js.PublishMsg(ctx, nmsg)
	if err != nil {
		return "", mapError(err, "publish")
	}
	return strconv.FormatUint(ack.Sequence, 10), nil
}
