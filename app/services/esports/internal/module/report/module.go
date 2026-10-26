package report

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewReportUC,
	NewRecordBoardRP,
	NewSnapshotRP,
	NewReportService,
)
