package service

import (
	"context"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	"cyber-ecosystem/apps/app_1/gen/go/v1/app1V1connect"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/biz"
)

type ReadingService struct {
	app1V1.UnimplementedReadingServiceServer

	log       *log.Helper
	readingUC *biz.ReadingUC
}

func NewReadingService(logger log.Logger, readingUC *biz.ReadingUC) *ReadingService {
	return &ReadingService{
		log:       log.NewHelper(log.With(logger, "module", "service/reading")),
		readingUC: readingUC,
	}
}
func (s *ReadingService) RegisterGRPC(srv *grpc.Server) {
	app1V1.RegisterReadingServiceServer(srv, s)
}
func (s *ReadingService) RegisterHTTP(srv *http.Server) {
	app1V1.RegisterReadingServiceHTTPServer(srv, s)
}
func (s *ReadingService) RegisterConnect(srv *connect.Server) {
	srv.Register(app1V1connect.NewReadingServiceHandler(s, srv.HandlerOptions()...))
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *ReadingService) QueryBlogReading(ctx context.Context, in *app1V1.QueryBlogReadingRequest) (*app1V1.QueryBlogReadingResponse, error) {
	out, err := s.readingUC.QueryBlog(ctx, &biz.ReadingQueryIn{
		PageRequest:  utils.GetOrBuildPage(in.Page),
		OrderBy:      utils.ParseOrderBy(in.OrderBy),
		ID:           in.Id,
		Title:        in.Title,
		PublishedAtA: utils.GetPTimeFromPPbTime(in.PublishedAtA),
		PublishedAtZ: utils.GetPTimeFromPPbTime(in.PublishedAtZ),
	})
	if err != nil {
		return nil, err
	}
	return &app1V1.QueryBlogReadingResponse{
		Page: out.PageResponse,
		List: func() []*app1V1.BlogWithReading {
			result := make([]*app1V1.BlogWithReading, len(out.List))
			for i, entity := range out.List {
				result[i] = &app1V1.BlogWithReading{
					Id:           entity.ID,
					CreatedAt:    utils.GetPPbTimeFromPTime(entity.CreatedAt),
					UpdatedAt:    utils.GetPPbTimeFromPTime(entity.UpdatedAt),
					Title:        utils.ToPtrWrapper(entity.Title, wrapperspb.String),
					Content:      utils.ToPtrWrapper(entity.Content, wrapperspb.String),
					PublishedAt:  utils.GetPPbTimeFromPTime(entity.PublishedAt),
					ReadingCount: entity.ReadingCount,
				}
			}
			return result
		}(),
	}, nil
}

func (s *ReadingService) GetBlogReading(ctx context.Context, in *app1V1.GetBlogReadingRequest) (*app1V1.GetBlogReadingResponse, error) {
	entity, err := s.readingUC.GetBlog(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &app1V1.GetBlogReadingResponse{
		Id:           entity.ID,
		CreatedAt:    utils.GetPPbTimeFromPTime(entity.CreatedAt),
		UpdatedAt:    utils.GetPPbTimeFromPTime(entity.UpdatedAt),
		Title:        utils.ToPtrWrapper(entity.Title, wrapperspb.String),
		Content:      utils.ToPtrWrapper(entity.Content, wrapperspb.String),
		PublishedAt:  utils.GetPPbTimeFromPTime(entity.PublishedAt),
		ReadingCount: entity.ReadingCount,
	}, nil
}

func (s *ReadingService) RecordReading(ctx context.Context, in *app1V1.RecordReadingRequest) (*app1V1.RecordReadingResponse, error) {
	readingCount, err := s.readingUC.RecordReading(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &app1V1.RecordReadingResponse{
		ReadingCount: readingCount,
	}, nil
}
