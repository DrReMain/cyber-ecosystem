package sku

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

type SkuService struct {
	esportspb.UnimplementedSkuServiceServer

	log   *slog.Logger
	skuUC *SkuUC
}

func NewSkuService(logger *slog.Logger, skuUC *SkuUC) *SkuService {
	return &SkuService{
		log:   logger.With("module", "module/sku_service"),
		skuUC: skuUC,
	}
}

func (s *SkuService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterSkuServiceServer(srv, s)
}

func (s *SkuService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterSkuServiceHTTPServer(srv, s)
}

func (s *SkuService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterSkuServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *SkuService) ListSkus(ctx context.Context, in *esportspb.ListSkusRequest) (*esportspb.ListSkusResponse, error) {
	skus, err := s.skuUC.List(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListSkusResponse{Skus: utils.SliceMap(skus, mapSkuPB)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapSkuPB(s *Sku) *esportspb.Sku {
	return &esportspb.Sku{
		SkuCode:      utils.StringW(s.Code),
		Name:         utils.StringW(s.Name),
		Description:  utils.Wrap(s.Description, utils.StringW),
		Kind:         mapSkuKindPB(s.Kind),
		PriceFen:     utils.Int64W(s.PriceFen),
		ListPriceFen: utils.Int64W(s.ListPriceFen),
		PeriodDays:   utils.Wrap(utils.ConvPtr[int32](s.PeriodDays), utils.Int32W),
		CreatedAt:    utils.ToTimestamp(&s.CreatedAt),
		UpdatedAt:    utils.ToTimestamp(&s.UpdatedAt),
	}
}

func mapSkuKindPB(kind string) esportspb.SkuKind {
	switch kind {
	case KindTrial:
		return esportspb.SkuKind_SKU_KIND_TRIAL
	case KindMembership:
		return esportspb.SkuKind_SKU_KIND_MEMBERSHIP
	default:
		return esportspb.SkuKind_SKU_KIND_UNSPECIFIED
	}
}
