package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// region[rgba(66,165,245,0.15)] 🔵 Port --------------------------------------------------------------------------------

type MessageRP interface {
	Create(ctx context.Context, m *Message) (*Message, error)
	Update(ctx context.Context, fieldsMask []string, m *Message) (*Message, error)
	Delete(ctx context.Context, id string) (string, error)
	Get(ctx context.Context, id string) (*Message, error)
	Query(ctx context.Context, in *MessageQueryIn) (*MessageQueryOut, error)
	Sort(ctx context.Context, id string, prevID, nextID *string) (*Message, error)
}

// region[rgba(102,187,106,0.15)] 🟢 UC ----------------------------------------------------------------------------------

type MessageUC struct {
	UC
	messageRP MessageRP
}

func NewMessageUC(logger log.Logger, tm Transaction, messageRP MessageRP) *MessageUC {
	return &MessageUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/message_uc")),
			tm:  tm,
		},
		messageRP: messageRP,
	}
}

// region[rgba(186,104,200,0.15)] 🟣 Method ------------------------------------------------------------------------------

func (uc *MessageUC) Create(ctx context.Context, m *Message) (out *Message, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.messageRP.Create(ctx, m)
		return
	})
	return
}

func (uc *MessageUC) Update(ctx context.Context, fieldsMask []string, m *Message) (out *Message, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.messageRP.Update(ctx, fieldsMask, m)
		return
	})
	return
}

func (uc *MessageUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.messageRP.Delete(ctx, id)
		return
	})
	return
}

func (uc *MessageUC) Get(ctx context.Context, id string) (*Message, error) {
	return uc.messageRP.Get(ctx, id)
}

func (uc *MessageUC) Query(ctx context.Context, in *MessageQueryIn) (*MessageQueryOut, error) {
	return uc.messageRP.Query(ctx, in)
}

func (uc *MessageUC) UpdateStatus(ctx context.Context, id string, target string) (out *Message, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		m, e := uc.messageRP.Get(ctx, id)
		if e != nil {
			return e
		}
		if e = m.TransitionTo(ctx, target); e != nil {
			return e
		}
		out, e = uc.messageRP.Update(ctx, []string{"status"}, m)
		return
	})
	return
}

func (uc *MessageUC) Sort(ctx context.Context, id string, prevID, nextID *string) (out *Message, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.messageRP.Sort(ctx, id, prevID, nextID)
		return
	})
	return
}
