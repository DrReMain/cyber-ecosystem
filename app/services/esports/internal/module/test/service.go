package test

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type TestService struct {
	esportspb.UnimplementedTestServiceServer

	log    *slog.Logger
	testUC *TestUC
}

func NewTestService(logger *slog.Logger, testUC *TestUC) *TestService {
	return &TestService{
		log:    logger.With("module", "module/test_service"),
		testUC: testUC,
	}
}

func (s *TestService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterTestServiceServer(srv, s)
}

func (s *TestService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterTestServiceHTTPServer(srv, s)
}

func (s *TestService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterTestServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *TestService) ListTests(ctx context.Context, _ *esportspb.ListTestsRequest) (*esportspb.ListTestsResponse, error) {
	tests, err := s.testUC.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*esportspb.Test, 0, len(tests))
	for _, t := range tests {
		out = append(out, mapTestPB(t))
	}
	return &esportspb.ListTestsResponse{Tests: out}, nil
}

func (s *TestService) GetTest(ctx context.Context, in *esportspb.GetTestRequest) (*esportspb.GetTestResponse, error) {
	t, err := s.testUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetTestResponse{Test: mapTestPB(t), Introduction: utils.Wrap(t.Introduction, utils.StringW)}, nil
}

func (s *TestService) GetTestConfig(ctx context.Context, in *esportspb.GetTestConfigRequest) (*esportspb.GetTestConfigResponse, error) {
	cfg, err := s.testUC.GetConfig(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetTestConfigResponse{Config: &esportspb.TestConfig{
		TestId:                   utils.StringW(cfg.TestID),
		DurationSeconds:          utils.ConvNum[int32](cfg.DurationSec),
		BackgroundInvalidSeconds: utils.ConvNum[int32](cfg.BackgroundInvalidSec),
		EngineParams:             cfg.EngineParams,
	}}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapTestPB(t *Test) *esportspb.Test {
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
