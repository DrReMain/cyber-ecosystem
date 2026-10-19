package fileproxy

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
)

// Struct --------------------------------------------------------------------------------------------------------------

type FileProxyService struct {
	systempb.UnimplementedFileProxyServiceServer

	log         *slog.Logger
	fileProxyUC *FileProxyUC
}

func NewFileProxyService(logger *slog.Logger, fileProxyUC *FileProxyUC) *FileProxyService {
	return &FileProxyService{
		log:         logger.With("module", "module/fileproxy_service"),
		fileProxyUC: fileProxyUC,
	}
}

func (s *FileProxyService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterFileProxyServiceServer(srv, s)
}

func (s *FileProxyService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterFileProxyServiceHTTPServer(srv, s)
}

func (s *FileProxyService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterFileProxyServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *FileProxyService) UploadFile(ctx context.Context, in *systempb.UploadFileRequest) (*systempb.UploadFileResponse, error) {
	f, err := s.fileProxyUC.Create(ctx, &CreateIn{
		Name:        in.GetName(),
		ContentType: in.GetContentType(),
		Data:        in.GetData(),
	})
	if err != nil {
		return nil, err
	}
	return &systempb.UploadFileResponse{File: file.ToProtoFile(f)}, nil
}
