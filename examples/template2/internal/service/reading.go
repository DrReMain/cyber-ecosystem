package service

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/biz"
	template2V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template2/v1"

	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/util"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ReadingService struct {
	template2V1.UnimplementedReadingServiceServer

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
	template2V1.RegisterReadingServiceServer(srv, s)
}
func (s *ReadingService) RegisterHTTP(srv *http.Server) {
	template2V1.RegisterReadingServiceHTTPServer(srv, s)
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *ReadingService) QueryBlog(ctx context.Context, in *template2V1.QueryBlogRequest) (*template2V1.QueryBlogResponse, error) {
	out, err := s.readingUC.QueryBlog(ctx, &biz.ReadingQueryIn{
		PageRequest:  util.GetOrBuildPage(in.Page),
		OrderBy:      order_by.ParseOrderBy(in.OrderBy),
		ID:           in.Id,
		Title:        in.Title,
		PublishedAtA: util.GetPTimeFromPPbTime(in.PublishedAtA),
		PublishedAtZ: util.GetPTimeFromPPbTime(in.PublishedAtZ),
	})
	if err != nil {
		return nil, err
	}
	return &template2V1.QueryBlogResponse{
		Page: out.PageResponse,
		List: func() []*template2V1.BlogWithReading {
			result := make([]*template2V1.BlogWithReading, len(out.List))
			for i, entity := range out.List {
				result[i] = &template2V1.BlogWithReading{
					Id:           entity.ID,
					Title:        util.ToPtrWrapper(entity.Title, wrapperspb.String),
					Content:      util.ToPtrWrapper(entity.Content, wrapperspb.String),
					PublishedAt:  util.GetPPbTimeFromPTime(entity.PublishedAt),
					CreatedAt:    util.GetPPbTimeFromPTime(entity.CreatedAt),
					UpdatedAt:    util.GetPPbTimeFromPTime(entity.UpdatedAt),
					ReadingCount: entity.ReadingCount,
				}
			}
			return result
		}(),
	}, nil
}

func (s *ReadingService) GetBlog(ctx context.Context, in *template2V1.GetBlogRequest) (*template2V1.GetBlogResponse, error) {
	entity, err := s.readingUC.GetBlog(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &template2V1.GetBlogResponse{
		Id:           entity.ID,
		Title:        util.ToPtrWrapper(entity.Title, wrapperspb.String),
		Content:      util.ToPtrWrapper(entity.Content, wrapperspb.String),
		PublishedAt:  util.GetPPbTimeFromPTime(entity.PublishedAt),
		CreatedAt:    util.GetPPbTimeFromPTime(entity.CreatedAt),
		UpdatedAt:    util.GetPPbTimeFromPTime(entity.UpdatedAt),
		ReadingCount: entity.ReadingCount,
	}, nil
}

func (s *ReadingService) RecordReading(ctx context.Context, in *template2V1.RecordReadingRequest) (*template2V1.RecordReadingResponse, error) {
	readingCount, err := s.readingUC.RecordReading(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &template2V1.RecordReadingResponse{
		ReadingCount: readingCount,
	}, nil
}
