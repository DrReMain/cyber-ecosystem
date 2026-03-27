package utils

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cyber-ecosystem/contracts/go/common"
)

func Ptr[T any](v T) *T {
	return &v
}

func ToPtrWrapper[T any, W any](v *T, constructor func(T) *W) *W {
	if v == nil {
		return nil
	}
	return constructor(*v)
}

func GetOrBuildPage(request *common.PageRequest) *common.PageRequest {
	if request == nil {
		return &common.PageRequest{
			PageNo:     nil,
			PageSize:   nil,
			All:        nil,
			CreatedAtA: nil,
			CreatedAtZ: nil,
			UpdatedAtA: nil,
			UpdatedAtZ: nil,
		}
	}
	return request
}

func GetPTimeFromPPbTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	return Ptr(ts.AsTime())
}

func GetPPbTimeFromPTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func GetStringFromWrapper(v *wrapperspb.StringValue) *string {
	if v == nil {
		return nil
	}
	return Ptr(v.GetValue())
}
