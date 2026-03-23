package service

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"
	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	template1V1connect "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1/template1V1connect"

	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/util"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type BlogService struct {
	template1V1.UnimplementedBlogServiceServer

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
	template1V1.RegisterBlogServiceServer(srv, s)
}
func (s *BlogService) RegisterHTTP(srv *http.Server) {
	template1V1.RegisterBlogServiceHTTPServer(srv, s)
}
func (s *BlogService) RegisterConnect(srv *connect.Server) {
	srv.Register(template1V1connect.NewBlogServiceHandler(s, srv.HandlerOptions()...))
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *BlogService) CreateBlog(ctx context.Context, in *template1V1.CreateBlogRequest) (*template1V1.CreateBlogResponse, error) {
	entity := &biz.BlogEntity{
		Title:       in.Title,
		Content:     in.Content,
		PublishedAt: util.GetPTimeFromPPbTime(in.PublishedAt),
	}
	if err := s.blogUC.CreateBlog(ctx, entity); err != nil {
		return nil, err
	}
	return &template1V1.CreateBlogResponse{}, nil
}

func (s *BlogService) UpdateBlog(ctx context.Context, in *template1V1.UpdateBlogRequest) (*template1V1.UpdateBlogResponse, error) {
	entity := &biz.BlogEntity{
		ID:          in.Id,
		Title:       in.Title,
		Content:     in.Content,
		PublishedAt: util.GetPTimeFromPPbTime(in.PublishedAt),
	}
	if err := s.blogUC.UpdateBlog(ctx, in.FieldsMask, entity); err != nil {
		return nil, err
	}
	return &template1V1.UpdateBlogResponse{}, nil
}

func (s *BlogService) DeleteBlog(ctx context.Context, in *template1V1.DeleteBlogRequest) (*template1V1.DeleteBlogResponse, error) {
	if err := s.blogUC.DeleteBlog(ctx, in.Id); err != nil {
		return nil, err
	}
	return &template1V1.DeleteBlogResponse{}, nil
}

func (s *BlogService) DeleteBatchBlog(ctx context.Context, in *template1V1.DeleteBatchBlogRequest) (*template1V1.DeleteBatchBlogResponse, error) {
	count, err := s.blogUC.DeleteBatchBlog(ctx, in.Ids)
	if err != nil {
		return nil, err
	}
	return &template1V1.DeleteBatchBlogResponse{Count: int32(count)}, nil
}

func (s *BlogService) GetBlog(ctx context.Context, in *template1V1.GetBlogRequest) (*template1V1.GetBlogResponse, error) {
	entity, err := s.blogUC.GetBlog(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &template1V1.GetBlogResponse{
		Id:          entity.ID,
		Title:       util.ToPtrWrapper(entity.Title, wrapperspb.String),
		Content:     util.ToPtrWrapper(entity.Content, wrapperspb.String),
		PublishedAt: util.GetPPbTimeFromPTime(entity.PublishedAt),
		CreatedAt:   util.GetPPbTimeFromPTime(entity.CreatedAt),
		UpdatedAt:   util.GetPPbTimeFromPTime(entity.UpdatedAt),
	}, nil
}

func (s *BlogService) QueryBlog(ctx context.Context, in *template1V1.QueryBlogRequest) (*template1V1.QueryBlogResponse, error) {
	out, err := s.blogUC.QueryBlog(ctx, &biz.BlogQueryIn{
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
	return &template1V1.QueryBlogResponse{
		Page: out.PageResponse,
		List: func() []*template1V1.GetBlogResponse {
			result := make([]*template1V1.GetBlogResponse, len(out.List))
			for i, entity := range out.List {
				result[i] = &template1V1.GetBlogResponse{
					Id:          entity.ID,
					Title:       util.ToPtrWrapper(entity.Title, wrapperspb.String),
					Content:     util.ToPtrWrapper(entity.Content, wrapperspb.String),
					PublishedAt: util.GetPPbTimeFromPTime(entity.PublishedAt),
					CreatedAt:   util.GetPPbTimeFromPTime(entity.CreatedAt),
					UpdatedAt:   util.GetPPbTimeFromPTime(entity.UpdatedAt),
				}
			}
			return result
		}(),
	}, nil
}
