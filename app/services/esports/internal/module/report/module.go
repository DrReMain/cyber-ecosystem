package report

import (
	"github.com/google/wire"

	"cyber-ecosystem/app/services/esports/internal/module/profile"
)

var ProviderSet = wire.NewSet(
	NewReportUC,
	NewRecordBoardRP,
	NewSnapshotRP,
	NewReportService,
	NewMergeUC,
	wire.Bind(new(profile.MergePreviewer), new(*MergeUC)),
)
