package report

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewRecordBoardRP,
	NewReportUC,
	NewService,
)
