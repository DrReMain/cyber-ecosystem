package sample

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	samplepb "cyber-ecosystem/gen/go/cyber/sample/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type SampleService struct {
	samplepb.UnimplementedSampleServiceServer

	log      *slog.Logger
	sampleUC *SampleUC
}

func NewSampleService(logger *slog.Logger, sampleUC *SampleUC) *SampleService {
	return &SampleService{
		log:      logger.With("module", "module/sample_service"),
		sampleUC: sampleUC,
	}
}

func (s *SampleService) RegisterGRPC(srv *grpc.Server) {
	samplepb.RegisterSampleServiceServer(srv, s)
}

func (s *SampleService) RegisterHTTP(srv *http.Server) {
	samplepb.RegisterSampleServiceHTTPServer(srv, s)
}

func (s *SampleService) RegisterConnect(srv *connecttransport.Server) {
	samplepb.RegisterSampleServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *SampleService) CreateSample(ctx context.Context, in *samplepb.CreateSampleRequest) (*samplepb.CreateSampleResponse, error) {
	created, err := s.sampleUC.Create(ctx, &Item{
		Name: in.Name,
	})
	if err != nil {
		return nil, err
	}
	return &samplepb.CreateSampleResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *SampleService) ListSamples(ctx context.Context, in *samplepb.ListSamplesRequest) (*samplepb.ListSamplesResponse, error) {
	out, err := s.sampleUC.List(ctx, &ItemListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
	})
	if err != nil {
		return nil, err
	}
	return &samplepb.ListSamplesResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.itemToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *SampleService) itemToProto(a *Item) *samplepb.Sample {
	return &samplepb.Sample{
		Id:        utils.StringW(a.ID),
		CreatedAt: utils.ToTimestamp(&a.CreatedAt),
		Name:      utils.Wrap(a.Name, utils.StringW),
	}
}
