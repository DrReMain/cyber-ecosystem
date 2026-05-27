package biz

import (
	"context"

	"github.com/looplab/fsm"

	"cyber-ecosystem/shared-go/utils"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
)

// region[rgba(255,238,88,0.12)] 🟡 FSM States ---------------------------------------------------------------------------

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// region[rgba(255,167,38,0.15)] 🟠 FSM ----------------------------------------------------------------------------------

func newArticleFSM(current string, a *Article) *fsm.FSM {
	return fsm.NewFSM(
		current,
		[]fsm.EventDesc{
			{Name: StatusPublished, Src: []string{StatusDraft}, Dst: StatusPublished},
			{Name: StatusArchived, Src: []string{StatusDraft, StatusPublished}, Dst: StatusArchived},
			{Name: StatusDraft, Src: []string{StatusArchived}, Dst: StatusDraft},
		},
		map[string]fsm.Callback{
			"after_" + StatusPublished: func(_ context.Context, _ *fsm.Event) { *a.Status = StatusPublished },
			"after_" + StatusArchived:  func(_ context.Context, _ *fsm.Event) { *a.Status = StatusArchived },
			"after_" + StatusDraft:     func(_ context.Context, _ *fsm.Event) { *a.Status = StatusDraft },
		},
	)
}

// region[rgba(186,104,200,0.15)] 🟣 Domain Method -----------------------------------------------------------------------

func (a *Article) TransitionTo(ctx context.Context, target string) error {
	a.Status = utils.Ptr(utils.Deref(a.Status, StatusDraft))
	f := newArticleFSM(*a.Status, a)
	if err := f.Event(ctx, target); err != nil {
		return genesisV1.ErrorErrorReasonArticleInvalidStatusTransition("").WithCause(err)
	}
	return nil
}
