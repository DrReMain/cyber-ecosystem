package utils

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToTimestamp converts *time.Time to *timestamppb.Timestamp.
// Returns nil if t is nil or zero.
func ToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// FromTimestamp converts *timestamppb.Timestamp to *time.Time.
// Returns nil if ts is nil or invalid.
func FromTimestamp(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// ToTime converts *timestamppb.Timestamp to time.Time.
// Returns zero value if ts is nil or invalid.
func ToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil || !ts.IsValid() {
		return time.Time{}
	}
	return ts.AsTime()
}
