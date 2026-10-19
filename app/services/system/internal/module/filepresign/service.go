package filepresign

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/protobuf/types/known/timestamppb"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
)

// Struct --------------------------------------------------------------------------------------------------------------

type FilePresignService struct {
	systempb.UnimplementedFilePresignServiceServer

	log           *slog.Logger
	filePresignUC *FilePresignUC
}

func NewFilePresignService(logger *slog.Logger, filePresignUC *FilePresignUC) *FilePresignService {
	return &FilePresignService{
		log:           logger.With("module", "module/filepresign_service"),
		filePresignUC: filePresignUC,
	}
}

func (s *FilePresignService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterFilePresignServiceServer(srv, s)
}

func (s *FilePresignService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterFilePresignServiceHTTPServer(srv, s)
}

func (s *FilePresignService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterFilePresignServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *FilePresignService) CreateUpload(ctx context.Context, in *systempb.CreateUploadRequest) (*systempb.CreateUploadResponse, error) {
	out, err := s.filePresignUC.Create(ctx, &CreateIn{
		Name:        in.GetName(),
		ContentType: in.GetContentType(),
		Size:        in.GetSize(),
	})
	if err != nil {
		return nil, err
	}
	resp := &systempb.CreateUploadResponse{File: file.ToProtoFile(out.File)}
	if out.Single != nil {
		resp.Channel = &systempb.CreateUploadResponse_Single{
			Single: &systempb.PresignedPut{Url: out.Single.URL, ExpiresAt: timestamppb.New(out.Single.ExpiresAt)},
		}
	} else {
		partURLs := make([]*systempb.UploadPartUrl, 0, len(out.Multipart.PartURLs))
		for _, p := range out.Multipart.PartURLs {
			partURLs = append(partURLs, &systempb.UploadPartUrl{PartNumber: p.PartNumber, Url: p.URL})
		}
		resp.Channel = &systempb.CreateUploadResponse_Multipart{
			Multipart: &systempb.UploadSession{
				UploadId:  out.Multipart.UploadID,
				PartSize:  out.Multipart.PartSize,
				PartCount: out.Multipart.PartCount,
				PartUrls:  partURLs,
				ExpiresAt: timestamppb.New(out.Multipart.ExpiresAt),
			},
		}
	}
	return resp, nil
}

func (s *FilePresignService) ConfirmUpload(ctx context.Context, in *systempb.ConfirmUploadRequest) (*systempb.ConfirmUploadResponse, error) {
	parts := make([]PartETag, 0, len(in.GetParts()))
	for _, p := range in.GetParts() {
		parts = append(parts, PartETag{PartNumber: p.GetPartNumber(), ETag: p.GetEtag()})
	}
	f, err := s.filePresignUC.Confirm(ctx, in.GetId(), parts)
	if err != nil {
		return nil, err
	}
	return &systempb.ConfirmUploadResponse{File: file.ToProtoFile(f)}, nil
}

func (s *FilePresignService) AbortUpload(ctx context.Context, in *systempb.AbortUploadRequest) (*systempb.AbortUploadResponse, error) {
	if err := s.filePresignUC.Abort(ctx, in.GetId()); err != nil {
		return nil, err
	}
	return &systempb.AbortUploadResponse{}, nil
}

func (s *FilePresignService) ListUploadedParts(ctx context.Context, in *systempb.ListUploadedPartsRequest) (*systempb.ListUploadedPartsResponse, error) {
	out, err := s.filePresignUC.ListParts(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	uploaded := make([]*systempb.UploadedPart, 0, len(out.Uploaded))
	for _, p := range out.Uploaded {
		uploaded = append(uploaded, &systempb.UploadedPart{PartNumber: p.PartNumber, Etag: p.ETag})
	}
	missing := make([]*systempb.UploadPartUrl, 0, len(out.Missing))
	for _, p := range out.Missing {
		missing = append(missing, &systempb.UploadPartUrl{PartNumber: p.PartNumber, Url: p.URL})
	}
	return &systempb.ListUploadedPartsResponse{
		File:      file.ToProtoFile(out.File),
		PartSize:  out.PartSize,
		Uploaded:  uploaded,
		Missing:   missing,
		ExpiresAt: timestamppb.New(out.ExpiresAt),
	}, nil
}
