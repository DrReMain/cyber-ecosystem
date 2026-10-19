package utils

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToTimestamp(t *testing.T) {
	if got := ToTimestamp(nil); got != nil {
		t.Errorf("ToTimestamp(nil) = %v, want nil", got)
	}
	zero := time.Time{}
	if got := ToTimestamp(&zero); got != nil {
		t.Errorf("ToTimestamp(&zero) = %v, want nil", got)
	}
	now := time.Now()
	got := ToTimestamp(&now)
	if got == nil || !got.AsTime().Equal(now) {
		t.Errorf("ToTimestamp(&now) = %v, want %v", got, now)
	}
}

func TestFromTimestamp(t *testing.T) {
	if got := FromTimestamp(nil); got != nil {
		t.Errorf("FromTimestamp(nil) = %v, want nil", got)
	}
	now := time.Now()
	ts := timestamppb.New(now)
	got := FromTimestamp(ts)
	if got == nil || !got.Equal(now) {
		t.Errorf("FromTimestamp(valid) = %v, want %v", got, now)
	}
}

func TestToTime(t *testing.T) {
	if got := ToTime(nil); !got.IsZero() {
		t.Errorf("ToTime(nil) = %v, want zero", got)
	}
	now := time.Now()
	ts := timestamppb.New(now)
	got := ToTime(ts)
	if !got.Equal(now) {
		t.Errorf("ToTime(valid) = %v, want %v", got, now)
	}
}
