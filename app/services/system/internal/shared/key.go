package shared

import "github.com/rs/xid"

func NewKey() string {
	return "f/" + xid.New().String()
}
