package biz

import (
	"context"

	"github.com/looplab/fsm"

	"cyber-ecosystem/shared-go/utils"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
)

// FSM States --------------------------------------------------------------------------------------------------------

const (
	StatusEnabled    = "enabled"
	StatusDisabled   = "disabled"
	StatusRestricted = "restricted"
)

// FSM --------------------------------------------------------------------------------------------------------

func newMobileUserFSM(current string, u *MobileUser) *fsm.FSM {
	return fsm.NewFSM(
		current,
		[]fsm.EventDesc{
			{Name: StatusDisabled, Src: []string{StatusEnabled}, Dst: StatusDisabled},
			{Name: StatusRestricted, Src: []string{StatusEnabled}, Dst: StatusRestricted},
			{Name: StatusEnabled, Src: []string{StatusDisabled, StatusRestricted}, Dst: StatusEnabled},
		},
		map[string]fsm.Callback{
			"after_" + StatusDisabled:   func(_ context.Context, _ *fsm.Event) { *u.Status = StatusDisabled },
			"after_" + StatusRestricted: func(_ context.Context, _ *fsm.Event) { *u.Status = StatusRestricted },
			"after_" + StatusEnabled:    func(_ context.Context, _ *fsm.Event) { *u.Status = StatusEnabled },
		},
	)
}

// Domain Method --------------------------------------------------------------------------------------------------------

func (u *MobileUser) TransitionTo(ctx context.Context, target string) error {
	u.Status = utils.Ptr(utils.Deref(u.Status, StatusEnabled))
	f := newMobileUserFSM(*u.Status, u)
	if err := f.Event(ctx, target); err != nil {
		return mobilepb.ErrorErrorReasonStatusInvalidTransition("").WithCause(err)
	}
	return nil
}
