package biz

import (
	"context"

	"github.com/looplab/fsm"

	"cyber-ecosystem/shared-go/utils"

	templateV1 "cyber-ecosystem/apps/template/gen/go/v1"
)

// region[rgba(255,238,88,0.12)] 🟡 FSM States ---------------------------------------------------------------------------

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// region[rgba(255,167,38,0.15)] 🟠 FSM ---------------------------------------------------------------------------------

func newMessageFSM(current string, m *Message) *fsm.FSM {
	return fsm.NewFSM(
		current,
		[]fsm.EventDesc{
			{Name: StatusPublished, Src: []string{StatusDraft}, Dst: StatusPublished},
			{Name: StatusArchived, Src: []string{StatusDraft, StatusPublished}, Dst: StatusArchived},
			{Name: StatusDraft, Src: []string{StatusArchived}, Dst: StatusDraft},
		},
		map[string]fsm.Callback{
			"after_" + StatusPublished: func(_ context.Context, _ *fsm.Event) { *m.Status = StatusPublished },
			"after_" + StatusArchived:  func(_ context.Context, _ *fsm.Event) { *m.Status = StatusArchived },
			"after_" + StatusDraft:     func(_ context.Context, _ *fsm.Event) { *m.Status = StatusDraft },
		},
	)
}

// region[rgba(186,104,200,0.15)] 🟣 Domain Method -----------------------------------------------------------------------

func (m *Message) TransitionTo(ctx context.Context, target string) error {
	m.Status = utils.Ptr(utils.Deref(m.Status, StatusDraft))
	f := newMessageFSM(*m.Status, m)
	if err := f.Event(ctx, target); err != nil {
		return templateV1.ErrorErrorReasonMessageInvalidStatusTransition("").WithCause(err)
	}
	return nil
}
