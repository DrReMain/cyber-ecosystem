package report

import (
	"github.com/google/wire"

	"cyber-ecosystem/app/services/esports/internal/module/profile"
)

var ProviderSet = wire.NewSet(
	NewMergeUC,
	NewReportUC,
	NewRecordBoardRP,
	NewSnapshotRP,
	NewReportService,
	wire.Bind(new(profile.MergePreviewer), new(*MergeUC)),
)
