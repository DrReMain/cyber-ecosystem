package report

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ReportService struct {
	esportspb.UnimplementedReportServiceServer

	log *slog.Logger
	uc  *ReportUC
}

func NewReportService(logger *slog.Logger, uc *ReportUC) *ReportService {
	return &ReportService{log: logger.With("module", "module/report_service"), uc: uc}
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
		Comparison:          comparisonPB(r.Comparison),
		Improvements:        improvementsPB(r.Improvements),
		TestedDimensions:    utils.ConvNum[int32](r.Tested),
		TotalDimensions:     utils.ConvNum[int32](r.Total),
	}, nil
}

func (s *ReportService) GetSessionReport(ctx context.Context, req *esportspb.GetSessionReportRequest) (*esportspb.GetSessionReportResponse, error) {
	r, err := s.uc.GetSessionReport(ctx, req.GetSessionId())
	if err != nil {
		return nil, err
	}
	return &esportspb.GetSessionReportResponse{
		SessionId:           r.SessionID,
		Genre:               r.Genre,
		AnchorRecordId:      utils.Wrap(r.Anchor, utils.StringW),
		Tier:                tierPB(r.Tier),
		CompositePercentile: utils.Wrap(r.Composite, utils.DoubleW),
		Abilities:           sessionAbilitiesPB(r.Abilities),
		TestedDimensions:    utils.ConvNum[int32](r.Tested),
		TotalDimensions:     utils.ConvNum[int32](r.Total),
		StartedAt:           utils.ToTimestamp(&r.StartedAt),
	}, nil
}

func (s *ReportService) GetAbilityDigest(ctx context.Context, _ *esportspb.GetAbilityDigestRequest) (*esportspb.GetAbilityDigestResponse, error) {
	d, err := s.uc.GetAbilityDigest(ctx)
	if err != nil {
		return nil, err
	}
	abilities := make([]*esportspb.DigestAbilityLine, 0, len(d.Abilities))
	for _, a := range d.Abilities {
		abilities = append(abilities, &esportspb.DigestAbilityLine{
			Ability:        a.Ability,
			BestScore:      a.Best,
			BestGrade:      utils.Wrap(a.BestGrade, utils.StringW),
			BestPercentile: utils.Wrap(a.Percentile, utils.DoubleW),
			SessionsTested: utils.ConvNum[int32](a.Sessions),
			LastAt:         utils.ToTimestamp(&a.LastAt),
		})
	}
	sessions := make([]*esportspb.DigestSessionLine, 0, len(d.Sessions))
	for _, line := range d.Sessions {
		sessions = append(sessions, &esportspb.DigestSessionLine{
			SessionId:      line.SessionID,
			Genre:          line.Genre,
			Finished:       line.Finished,
			DoneCount:      utils.ConvNum[int32](line.DoneCount),
			StepCount:      utils.ConvNum[int32](line.StepCount),
			CreatedAt:      utils.ToTimestamp(&line.CreatedAt),
			AnchorRecordId: utils.Wrap(line.Anchor, utils.StringW),
		})
	}
	return &esportspb.GetAbilityDigestResponse{
		SessionsTotal: utils.ConvNum[int32](d.SessionsTotal),
		SessionsFps:   utils.ConvNum[int32](d.SessionsFPS),
		SessionsMoba:  utils.ConvNum[int32](d.SessionsMOBA),
		Abilities:     abilities,
		Sessions:      sessions,
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func sessionAbilitiesPB(rows []SessionAbilityRow) []*esportspb.SessionAbilityLine {
	out := make([]*esportspb.SessionAbilityLine, 0, len(rows))
	for _, r := range rows {
		out = append(out, &esportspb.SessionAbilityLine{
			Ability:          r.Ability,
			GameCode:         r.GameCode,
			BestScore:        r.Best,
			AttemptScores:    r.Attempts,
			BestGrade:        utils.Wrap(r.BestGrade, utils.StringW),
			Percentile:       utils.Wrap(r.Percentile, utils.DoubleW),
			PrevSessionScore: utils.Wrap(r.PrevSession, utils.DoubleW),
			HistoryBestScore: utils.Wrap(r.HistoryBest, utils.DoubleW),
		})
	}
	return out
}

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

func improvementsPB(items []ReportImprovement) []*esportspb.ReportImprovement {
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
	case shared.TierS:
		return esportspb.ReportTier_REPORT_TIER_S
	case shared.TierA:
		return esportspb.ReportTier_REPORT_TIER_A
	case shared.TierB:
		return esportspb.ReportTier_REPORT_TIER_B
	case shared.TierC:
		return esportspb.ReportTier_REPORT_TIER_C
	}
	return esportspb.ReportTier_REPORT_TIER_UNSPECIFIED
}
