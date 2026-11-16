package audit

import (
	"github.com/google/wire"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
)

var ProviderSet = wire.NewSet(
	NewAuditUC,
	NewAuditRP,
	wire.Bind(new(kaudit.Sink), new(*AuditUC)),
)
