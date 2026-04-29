package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/apps/template/services/base/internal/biz"
	"cyber-ecosystem/apps/template/services/base/internal/conf"
)

// region[rgba(236,64,122,0.15)] 🩷 Struct -------------------------------------------------------------------------------

type FileService struct {
	log         *log.Helper
	fileUC      *biz.FileUC
	maxFileSize int64
}

func NewFileService(logger log.Logger, cd *conf.Data, fileUC *biz.FileUC) *FileService {
	return &FileService{
		log:    log.NewHelper(log.With(logger, "module", "service/file")),
		fileUC: fileUC,
		maxFileSize: func() int64 {
			if cd != nil && cd.Storage != nil && cd.Storage.MaxFileSize > 0 {
				return cd.Storage.MaxFileSize
			}
			return 50 * 1024 * 1024
		}(),
	}
}
func (s *FileService) RegisterHTTP(srv *krahttp.Server) {
	r := srv.Route("/api/v1")
	r.Handle("POST", "/files", s.handleUpload)
	r.Handle("GET", "/files/{id}", s.handleDownload)
}
func (s *FileService) RegisterGRPC(_ *grpc.Server)       {}
func (s *FileService) RegisterConnect(_ *connect.Server) {}

// region[rgba(255,167,38,0.15)] 🟠 Handler -----------------------------------------------------------------------------

func (s *FileService) handleUpload(ctx krahttp.Context) error {
	r := ctx.Request()
	r.Body = http.MaxBytesReader(ctx.Response(), r.Body, s.maxFileSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(err)
	}
	defer file.Close()

	if s.maxFileSize > 0 && header.Size > s.maxFileSize {
		return errorspb.ErrorInfraErrorStorageSizeExceed("").WithCause(fmt.Errorf("max size: %d bytes", s.maxFileSize))
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	h := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
		return s.fileUC.Upload(ctx, header.Filename, file, header.Size, contentType)
	})
	result, err := h(r.Context(), nil)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, result)
}

func (s *FileService) handleDownload(ctx krahttp.Context) error {
	type downloadResult struct {
		Body io.ReadCloser
		Info *biz.File
	}

	id := ctx.Vars().Get("id")
	if id == "" {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("missing file id"))
	}

	h := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
		body, info, err := s.fileUC.Download(ctx, id)
		if err != nil {
			return nil, err
		}
		return &downloadResult{Body: body, Info: info}, nil
	})
	result, err := h(ctx.Request().Context(), nil)
	if err != nil {
		return err
	}

	dr := result.(*downloadResult)
	defer dr.Body.Close()

	w := ctx.Response()
	w.Header().Set("Content-Type", dr.Info.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, escapeFilename(dr.Info.Name)))
	if dr.Info.Size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", dr.Info.Size))
	}
	io.Copy(w, dr.Body)
	return nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func escapeFilename(name string) string {
	r := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= ' ' && c <= '~' && c != '"' && c != '\\':
			r = append(r, c)
		default:
			r = append(r, '_')
		}
	}
	return string(r)
}
