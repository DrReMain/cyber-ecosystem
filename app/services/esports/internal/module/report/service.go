package report

import (
	"context"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ReportService struct {
	esportspb.UnimplementedReportServiceServer

	uc *ReportUC
}

func NewService(uc *ReportUC) *ReportService {
	return &ReportService{uc: uc}
}

func (s *ReportService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterReportServiceServer(srv, s)
}

func (s *ReportService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterReportServiceHTTPServer(srv, s)
}

func (s *ReportService) RegisterConnect(srv *connect.Server) {
	esportspb.RegisterReportServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ReportService) GetReport(ctx context.Context, req *esportspb.GetReportRequest) (*esportspb.GetReportResponse, error) {
	r, err := s.uc.GetReport(ctx, req.GetRecordId())
	if err != nil {
		return nil, err
	}
	return &esportspb.GetReportResponse{
		RecordId:            r.RecordID,
		Unlocked:            r.Unlocked,
		Tier:                tierPB(r.Tier),
		CompositePercentile: r.Composite,
		Abilities:           abilitiesPB(r.Abilities),
		SubAbilities:        nil, // the sub-ability board waits for the engines' secondary-contribution outlet; absent, never zero-axed
		Comparison:          comparisonPB(r.Comparison),
		Improvements:        improvementsPB(r.Improvements),
		TestedDimensions:    utils.ConvNum[int32](r.Tested),
		TotalDimensions:     utils.ConvNum[int32](r.Total),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func abilitiesPB(lines []AbilityLine) []*esportspb.ReportAbility {
	out := make([]*esportspb.ReportAbility, 0, len(lines))
	for _, l := range lines {
		out = append(out, &esportspb.ReportAbility{
			Ability:     l.Ability,
			GameCode:    l.GameCode,
			Score:       l.Score,
			Percentile:  l.Percentile,
			Confidence:  l.Confidence,
			SampleCount: utils.ConvNum[int32](l.SampleCount),
			PeerAverage: l.PeerAverage,
		})
	}
	return out
}

func comparisonPB(deltas []AbilityDelta) *esportspb.ReportComparison {
	if deltas == nil {
		return nil
	}
	out := make([]*esportspb.ReportComparison_Delta, 0, len(deltas))
	for _, d := range deltas {
		out = append(out, &esportspb.ReportComparison_Delta{
			Ability:            d.Ability,
			PreviousPercentile: d.Previous,
			DeltaPercentile:    d.Delta,
		})
	}
	return &esportspb.ReportComparison{Deltas: out}
}

func improvementsPB(items []Improvement) []*esportspb.ReportImprovement {
	out := make([]*esportspb.ReportImprovement, 0, len(items))
	for _, it := range items {
		out = append(out, &esportspb.ReportImprovement{
			Ability:           it.Ability,
			CurrentPercentile: it.Current,
			TargetPercentile:  targetPercentile,
			GapPercent:        it.Gap,
		})
	}
	return out
}

func tierPB(tier string) esportspb.ReportTier {
	switch tier {
	case "S":
		return esportspb.ReportTier_REPORT_TIER_S
	case "A":
		return esportspb.ReportTier_REPORT_TIER_A
	case "B":
		return esportspb.ReportTier_REPORT_TIER_B
	case "C":
		return esportspb.ReportTier_REPORT_TIER_C
	}
	return esportspb.ReportTier_REPORT_TIER_UNSPECIFIED
}
