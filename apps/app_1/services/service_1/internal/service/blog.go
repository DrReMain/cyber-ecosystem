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
	app1ConnectV1 "cyber-ecosystem/apps/app_1/gen/go/v1/app1V1connect"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
)

type BlogService struct {
	app1V1.UnimplementedBlogServiceServer

	log    *log.Helper
	blogUC *biz.BlogUC
}

func NewBlogService(logger log.Logger, blogUC *biz.BlogUC) *BlogService {
	return &BlogService{
		log:    log.NewHelper(log.With(logger, "module", "service/blog")),
		blogUC: blogUC,
	}
}
func (s *BlogService) RegisterGRPC(srv *grpc.Server) {
	app1V1.RegisterBlogServiceServer(srv, s)
}
func (s *BlogService) RegisterHTTP(srv *http.Server) {
	app1V1.RegisterBlogServiceHTTPServer(srv, s)
}
func (s *BlogService) RegisterConnect(srv *connect.Server) {
	srv.Register(app1ConnectV1.NewBlogServiceHandler(s, srv.HandlerOptions()...))
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *BlogService) CreateBlog(ctx context.Context, in *app1V1.CreateBlogRequest) (*app1V1.CreateBlogResponse, error) {
	entity := &biz.BlogEntity{
		Title:       in.Title,
		Content:     in.Content,
		PublishedAt: utils.GetPTimeFromPPbTime(in.PublishedAt),
	}
	if err := s.blogUC.CreateBlog(ctx, entity); err != nil {
		return nil, err
	}
	return &app1V1.CreateBlogResponse{}, nil
}

func (s *BlogService) UpdateBlog(ctx context.Context, in *app1V1.UpdateBlogRequest) (*app1V1.UpdateBlogResponse, error) {
	entity := &biz.BlogEntity{
		ID:          in.Id,
		Title:       in.Title,
		Content:     in.Content,
		PublishedAt: utils.GetPTimeFromPPbTime(in.PublishedAt),
	}
	if err := s.blogUC.UpdateBlog(ctx, in.FieldsMask, entity); err != nil {
		return nil, err
	}
	return &app1V1.UpdateBlogResponse{}, nil
}

func (s *BlogService) DeleteBlog(ctx context.Context, in *app1V1.DeleteBlogRequest) (*app1V1.DeleteBlogResponse, error) {
	if err := s.blogUC.DeleteBlog(ctx, in.Id); err != nil {
		return nil, err
	}
	return &app1V1.DeleteBlogResponse{}, nil
}

func (s *BlogService) DeleteBatchBlog(ctx context.Context, in *app1V1.DeleteBatchBlogRequest) (*app1V1.DeleteBatchBlogResponse, error) {
	count, err := s.blogUC.DeleteBatchBlog(ctx, in.Ids)
	if err != nil {
		return nil, err
	}
	return &app1V1.DeleteBatchBlogResponse{Count: int32(count)}, nil
}

func (s *BlogService) GetBlog(ctx context.Context, in *app1V1.GetBlogRequest) (*app1V1.GetBlogResponse, error) {
	entity, err := s.blogUC.GetBlog(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &app1V1.GetBlogResponse{
		Id:          entity.ID,
		CreatedAt:   utils.GetPPbTimeFromPTime(entity.CreatedAt),
		UpdatedAt:   utils.GetPPbTimeFromPTime(entity.UpdatedAt),
		Title:       utils.ToPtrWrapper(entity.Title, wrapperspb.String),
		Content:     utils.ToPtrWrapper(entity.Content, wrapperspb.String),
		PublishedAt: utils.GetPPbTimeFromPTime(entity.PublishedAt),
	}, nil
}

func (s *BlogService) QueryBlog(ctx context.Context, in *app1V1.QueryBlogRequest) (*app1V1.QueryBlogResponse, error) {
	out, err := s.blogUC.QueryBlog(ctx, &biz.BlogQueryIn{
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
	return &app1V1.QueryBlogResponse{
		Page: out.PageResponse,
		List: func() []*app1V1.GetBlogResponse {
			result := make([]*app1V1.GetBlogResponse, len(out.List))
			for i, entity := range out.List {
				result[i] = &app1V1.GetBlogResponse{
					Id:          entity.ID,
					CreatedAt:   utils.GetPPbTimeFromPTime(entity.CreatedAt),
					UpdatedAt:   utils.GetPPbTimeFromPTime(entity.UpdatedAt),
					Title:       utils.ToPtrWrapper(entity.Title, wrapperspb.String),
					Content:     utils.ToPtrWrapper(entity.Content, wrapperspb.String),
					PublishedAt: utils.GetPPbTimeFromPTime(entity.PublishedAt),
				}
			}
			return result
		}(),
	}, nil
}
