package skuadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/sku"
)

// Struct --------------------------------------------------------------------------------------------------------------

type SkuAdminService struct {
	esportspb.UnimplementedSkuAdminServiceServer

	log   *slog.Logger
	skuUC *sku.SkuUC
}

func NewSkuAdminService(logger *slog.Logger, skuUC *sku.SkuUC) *SkuAdminService {
	return &SkuAdminService{
		log:   logger.With("module", "module/skuadmin_service"),
		skuUC: skuUC,
	}
}

func (s *SkuAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterSkuAdminServiceServer(srv, s)
}

func (s *SkuAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterSkuAdminServiceHTTPServer(srv, s)
}

func (s *SkuAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterSkuAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *SkuAdminService) ListCatalogSkus(ctx context.Context, _ *esportspb.ListCatalogSkusRequest) (*esportspb.ListCatalogSkusResponse, error) {
	skus, err := s.skuUC.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListCatalogSkusResponse{Skus: utils.SliceMap(skus, adminSkuPB)}, nil
}

func (s *SkuAdminService) CreateSku(ctx context.Context, req *esportspb.CreateSkuRequest) (*esportspb.CreateSkuResponse, error) {
	in := &sku.SaveIn{
		Name:     req.GetName(),
		Kind:     kindFromPB(req.GetKind()),
		PriceFen: req.GetPriceFen(),
	}
	// Presence forwarding, not zero-sniffing: the optional fields ride as
	// the wire carried them, and the defaulting answers in the UC.
	in.ListPriceFen = req.ListPriceFen
	in.PeriodDays = utils.ConvPtr[int](req.PeriodDays)
	in.Description = req.Description
	saved, err := s.skuUC.Save(ctx, in)
	if err != nil {
		return nil, err
	}
	return &esportspb.CreateSkuResponse{Sku: adminSkuPB(saved)}, nil
}

func (s *SkuAdminService) UpdateSkuPrice(ctx context.Context, req *esportspb.UpdateSkuPriceRequest) (*esportspb.UpdateSkuPriceResponse, error) {
	updated, err := s.skuUC.UpdatePrice(ctx, req.GetSkuCode(), req.GetPriceFen(), req.ListPriceFen)
	if err != nil {
		return nil, err
	}
	return &esportspb.UpdateSkuPriceResponse{Sku: adminSkuPB(updated)}, nil
}

func (s *SkuAdminService) SetSkuShelf(ctx context.Context, req *esportspb.SetSkuShelfRequest) (*esportspb.SetSkuShelfResponse, error) {
	updated, err := s.skuUC.SetShelf(ctx, req.GetSkuCode(), req.GetOnShelf())
	if err != nil {
		return nil, err
	}
	return &esportspb.SetSkuShelfResponse{Sku: adminSkuPB(updated)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func adminSkuPB(s *sku.Sku) *esportspb.AdminSku {
	return &esportspb.AdminSku{
		Sku: &esportspb.Sku{
			SkuCode:      utils.StringW(s.Code),
			Name:         utils.StringW(s.Name),
			Description:  utils.Wrap(s.Description, utils.StringW),
			Kind:         skuKindPB(s.Kind),
			PriceFen:     utils.Int64W(s.PriceFen),
			ListPriceFen: utils.Int64W(s.ListPriceFen),
			PeriodDays:   utils.Wrap(utils.ConvPtr[int32](s.PeriodDays), utils.Int32W),
			CreatedAt:    utils.ToTimestamp(&s.CreatedAt),
			UpdatedAt:    utils.ToTimestamp(&s.UpdatedAt),
		},
		OnShelf:   s.OnShelf,
		UpdatedBy: utils.Wrap(s.UpdatedBy, utils.StringW),
	}
}

func skuKindPB(kind string) esportspb.SkuKind {
	switch kind {
	case sku.KindTrial:
		return esportspb.SkuKind_SKU_KIND_TRIAL
	case sku.KindMembership:
		return esportspb.SkuKind_SKU_KIND_MEMBERSHIP
	default:
		return esportspb.SkuKind_SKU_KIND_UNSPECIFIED
	}
}

func kindFromPB(kind esportspb.SkuKind) string {
	switch kind {
	case esportspb.SkuKind_SKU_KIND_TRIAL:
		return sku.KindTrial
	case esportspb.SkuKind_SKU_KIND_MEMBERSHIP:
		return sku.KindMembership
	default:
		return ""
	}
}
