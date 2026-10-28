package testadmin

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/protobuf/types/known/structpb"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/test"
)

// Struct --------------------------------------------------------------------------------------------------------------

type TestAdminService struct {
	esportspb.UnimplementedTestAdminServiceServer

	log    *slog.Logger
	testUC *test.TestUC
}

func NewTestAdminService(logger *slog.Logger, testUC *test.TestUC) *TestAdminService {
	return &TestAdminService{
		log:    logger.With("module", "module/testadmin_service"),
		testUC: testUC,
	}
}

// Handler ------------------------------------------------------------------------------------------------------------

func (s *TestAdminService) ListTests(ctx context.Context, _ *esportspb.TestAdminServiceListTestsRequest) (*esportspb.TestAdminServiceListTestsResponse, error) {
	tests, err := s.testUC.List(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.TestAdminServiceListTestsResponse{
		Tests: utils.SliceMap(tests, mapTestPB),
	}, nil
}

func (s *TestAdminService) GetTestConfig(ctx context.Context, req *esportspb.TestAdminServiceGetTestConfigRequest) (*esportspb.TestAdminServiceGetTestConfigResponse, error) {
	cfg, err := s.testUC.GetConfig(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &esportspb.TestAdminServiceGetTestConfigResponse{Config: mapConfigPB(cfg)}, nil
}

func (s *TestAdminService) UpdateTestConfig(ctx context.Context, req *esportspb.UpdateTestConfigRequest) (*esportspb.UpdateTestConfigResponse, error) {
	in := &test.UpdateConfigIn{TestID: req.GetTestId()}
	if req.BackgroundInvalidSeconds != nil {
		i := int(req.GetBackgroundInvalidSeconds())
		in.BackgroundInvalidSec = &i
	}
	in.EngineParams = req.GetEngineParams()
	cfg, err := s.testUC.UpdateConfig(ctx, in)
	if err != nil {
		return nil, err
	}
	return &esportspb.UpdateTestConfigResponse{Config: mapConfigPB(cfg)}, nil
}

func (s *TestAdminService) ListQuestionBank(ctx context.Context, req *esportspb.ListQuestionBankRequest) (*esportspb.ListQuestionBankResponse, error) {
	out, err := s.testUC.ListQuestionBank(ctx, req.GetTestCode(), req.GetPage())
	if err != nil {
		return nil, err
	}
	rows := make([]*esportspb.QuestionBankRow, 0, len(out.Rows))
	for _, r := range out.Rows {
		row, rerr := questionBankRowPB(r)
		if rerr != nil {
			return nil, esportspb.ErrorEsportsUnspecified("").WithCause(fmt.Errorf("bank row %s: %w", r.ScenarioCode, rerr))
		}
		rows = append(rows, row)
	}
	return &esportspb.ListQuestionBankResponse{Page: out.PageResponse, Questions: rows}, nil
}

func (s *TestAdminService) ListTestAttempts(ctx context.Context, req *esportspb.ListTestAttemptsRequest) (*esportspb.ListTestAttemptsResponse, error) {
	out, err := s.testUC.ListAttempts(ctx, &test.AttemptsQueryIn{
		PageRequest: req.GetPage(),
		TestCode:    req.TestCode,
		UserID:      req.UserId,
		SessionID:   req.MatchSessionId,
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.ListTestAttemptsResponse{
		Page:     out.PageResponse,
		Attempts: utils.SliceMap(out.Rows, attemptRowPB),
	}, nil
}

func (s *TestAdminService) GetTestAttempt(ctx context.Context, req *esportspb.GetTestAttemptRequest) (*esportspb.GetTestAttemptResponse, error) {
	rec, err := s.testUC.GetAttempt(ctx, req.GetRecordId())
	if err != nil {
		return nil, err
	}
	if rec == nil {
		// The same stored row the report verdicts cut from — one not-found
		// word for one entity.
		return nil, esportspb.ErrorEsportsReportRecordNotFound("")
	}
	detail, derr := structpb.NewStruct(rec.RawDetail)
	if derr != nil {
		return nil, esportspb.ErrorEsportsUnspecified("").WithCause(fmt.Errorf("raw detail to struct: %w", derr))
	}
	return &esportspb.GetTestAttemptResponse{Attempt: attemptRowPB(rec), RawDetail: detail}, nil
}

func (s *TestAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterTestAdminServiceServer(srv, s)
}

func (s *TestAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterTestAdminServiceHTTPServer(srv, s)
}

func (s *TestAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterTestAdminServiceConnectServer(srv, s)
}

// Private -------------------------------------------------------------------------------------------------------------

func mapTestPB(t *test.Test) *esportspb.Test {
	return &esportspb.Test{
		Id:              utils.StringW(t.ID),
		Code:            utils.StringW(t.Code),
		Name:            utils.StringW(t.Name),
		Category:        utils.Wrap(t.Category, utils.StringW),
		Ability:         utils.Wrap(t.Ability, utils.StringW),
		Description:     utils.Wrap(t.Description, utils.StringW),
		UnitScoreLabel:  utils.Wrap(t.UnitScoreLabel, utils.StringW),
		DurationSeconds: utils.ConvNum[int32](t.DurationSec),
		CreatedAt:       utils.ToTimestamp(&t.CreatedAt),
		UpdatedAt:       utils.ToTimestamp(&t.UpdatedAt),
	}
}

func mapConfigPB(c *test.TestConfig) *esportspb.TestConfig {
	return &esportspb.TestConfig{
		TestId:                   utils.StringW(c.TestID),
		DurationSeconds:          utils.ConvNum[int32](c.DurationSec),
		BackgroundInvalidSeconds: utils.ConvNum[int32](c.BackgroundInvalidSec),
		EngineParams:             c.EngineParams,
	}
}

func questionBankRowPB(r *test.QuestionBankRow) (*esportspb.QuestionBankRow, error) {
	payload, err := structpb.NewStruct(r.Payload)
	if err != nil {
		return nil, fmt.Errorf("payload to struct: %w", err)
	}
	answer, err := structpb.NewStruct(r.Answer)
	if err != nil {
		return nil, fmt.Errorf("answer to struct: %w", err)
	}
	return &esportspb.QuestionBankRow{
		ScenarioCode: r.ScenarioCode,
		Payload:      payload,
		Answer:       answer,
		CreatedAt:    utils.ToTimestamp(&r.CreatedAt),
	}, nil
}

func attemptRowPB(r *test.AttemptRecord) *esportspb.TestAttemptRow {
	return &esportspb.TestAttemptRow{
		RecordId:           r.ID,
		TestCode:           r.TestCode,
		UserId:             r.UserID,
		MatchSessionId:     utils.Wrap(r.SessionID, utils.StringW),
		AttemptNo:          utils.ConvNum[int32](r.AttemptNo),
		StartedAt:          utils.ToTimestamp(&r.StartedAt),
		EndedAt:            utils.ToTimestamp(&r.EndedAt),
		RawScore:           utils.Wrap(r.RawScore, utils.DoubleW),
		AuthoritativeScore: utils.Wrap(r.Authoritative, utils.DoubleW),
		Grade:              utils.Wrap(r.Grade, utils.StringW),
		Valid:              r.Valid,
		InvalidReason:      utils.Wrap(r.InvalidReason, utils.StringW),
	}
}
