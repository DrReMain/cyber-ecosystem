package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
	genesisV1connect "cyber-ecosystem/apps/genesis/gen/go/v1/genesisV1connect"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/biz"
)

// region[rgba(236,64,122,0.15)] 🩷 Struct -------------------------------------------------------------------------------

type ArticleService struct {
	genesisV1.UnimplementedAdminArticleServiceServer
	log       *log.Helper
	articleUC *biz.ArticleUC
}

func NewArticleService(logger log.Logger, articleUC *biz.ArticleUC) *ArticleService {
	return &ArticleService{
		log:       log.NewHelper(log.With(logger, "module", "service/article")),
		articleUC: articleUC,
	}
}

func (s *ArticleService) RegisterGRPC(srv *grpc.Server) {
	genesisV1.RegisterAdminArticleServiceServer(srv, s)
}
func (s *ArticleService) RegisterHTTP(srv *http.Server) {
	genesisV1.RegisterAdminArticleServiceHTTPServer(srv, s)
}
func (s *ArticleService) RegisterConnect(srv *connect.Server) {
	srv.Register(genesisV1connect.NewAdminArticleServiceHandler(s, srv.HandlerOptions()...))
}

// region[rgba(255,167,38,0.15)] 🟠 Handler -----------------------------------------------------------------------------

func (s *ArticleService) CreateArticle(ctx context.Context, in *genesisV1.CreateArticleRequest) (*genesisV1.CreateArticleResponse, error) {
	a := &biz.Article{
		Title:   in.Title,
		Content: in.Content,
	}
	created, err := s.articleUC.Create(ctx, a)
	if err != nil {
		return nil, err
	}
	return &genesisV1.CreateArticleResponse{
		Id: utils.Wrap(created.ID, utils.StringW),
	}, nil
}

func (s *ArticleService) UpdateArticle(ctx context.Context, in *genesisV1.UpdateArticleRequest) (*genesisV1.UpdateArticleResponse, error) {
	a := &biz.Article{
		ID:      in.Id,
		Title:   in.Title,
		Content: in.Content,
	}
	updated, err := s.articleUC.Update(ctx, in.FieldsMask, a)
	if err != nil {
		return nil, err
	}
	return &genesisV1.UpdateArticleResponse{
		Id: utils.Wrap(updated.ID, utils.StringW),
	}, nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, in *genesisV1.DeleteArticleRequest) (*genesisV1.DeleteArticleResponse, error) {
	deletedID, err := s.articleUC.Delete(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return &genesisV1.DeleteArticleResponse{
		Id: utils.StringW(deletedID),
	}, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, in *genesisV1.GetArticleRequest) (*genesisV1.GetArticleResponse, error) {
	a, err := s.articleUC.Get(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return s.articleToProto(a), nil
}

func (s *ArticleService) QueryArticle(ctx context.Context, in *genesisV1.QueryArticleRequest) (*genesisV1.QueryArticleResponse, error) {
	out, err := s.articleUC.Query(ctx, &biz.ArticleQueryIn{
		PageRequest: utils.EnsurePageRequest(in.Page),
		OrderBy:     utils.ParseOrderBy(in.OrderBy),
		Title:       in.Title,
		Status:      in.Status,
	})
	if err != nil {
		return nil, err
	}
	return &genesisV1.QueryArticleResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.articleToProto),
	}, nil
}

func (s *ArticleService) SortArticle(ctx context.Context, in *genesisV1.SortArticleRequest) (*genesisV1.SortArticleResponse, error) {
	sorted, err := s.articleUC.Sort(ctx, *in.Id, in.PrevId, in.NextId)
	if err != nil {
		return nil, err
	}
	return &genesisV1.SortArticleResponse{
		Id: utils.Wrap(sorted.ID, utils.StringW),
	}, nil
}

func (s *ArticleService) UpdateArticleStatus(ctx context.Context, in *genesisV1.UpdateArticleStatusRequest) (*genesisV1.UpdateArticleStatusResponse, error) {
	_, err := s.articleUC.UpdateStatus(ctx, *in.Id, *in.Status)
	if err != nil {
		return nil, err
	}
	return &genesisV1.UpdateArticleStatusResponse{
		Id: utils.StringW(*in.Id),
	}, nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func (s *ArticleService) articleToProto(a *biz.Article) *genesisV1.GetArticleResponse {
	return &genesisV1.GetArticleResponse{
		Id:        utils.Wrap(a.ID, utils.StringW),
		CreatedAt: utils.ToTimestamp(a.CreatedAt),
		UpdatedAt: utils.ToTimestamp(a.UpdatedAt),
		Title:     utils.Wrap(a.Title, utils.StringW),
		Content:   utils.Wrap(a.Content, utils.StringW),
		Status:    utils.Wrap(a.Status, utils.StringW),
	}
}
