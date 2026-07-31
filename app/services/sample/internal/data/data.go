package data

import (
	"log/slog"
)

// RP is the shared base for data repos. sample is currently a pure ACL facade with
// no own persistence; when it gains its own data, add a repo in its own file, declare
// a ProviderSet here, then register it in wireApp.
type RP struct {
	log *slog.Logger
}
