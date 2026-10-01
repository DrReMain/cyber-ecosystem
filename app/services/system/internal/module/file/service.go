package file

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type FileService struct {
	systempb.UnimplementedFileServiceServer

	log    *slog.Logger
	fileUC *FileUC
}

func NewFileService(logger *slog.Logger, fileUC *FileUC) *FileService {
	return &FileService{
		log:    logger.With("module", "module/file_service"),
		fileUC: fileUC,
	}
}

func (s *FileService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterFileServiceServer(srv, s)
}

func (s *FileService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterFileServiceHTTPServer(srv, s)
}

func (s *FileService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterFileServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *FileService) RenameFile(ctx context.Context, in *systempb.RenameFileRequest) (*systempb.RenameFileResponse, error) {
	if _, err := s.fileUC.Update(ctx, in.Id, in.GetName()); err != nil {
		return nil, err
	}
	return &systempb.RenameFileResponse{}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, in *systempb.DeleteFileRequest) (*systempb.DeleteFileResponse, error) {
	if err := s.fileUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &systempb.DeleteFileResponse{}, nil
}

func (s *FileService) ListFiles(ctx context.Context, in *systempb.ListFilesRequest) (*systempb.ListFilesResponse, error) {
	listIn := &FileListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Name:        in.Name,
		ContentType: in.ContentType,
	}
	if in.Status != nil {
		listIn.Status = utils.Ptr(statusToString(*in.Status))
	}
	if in.Source != nil {
		listIn.Source = utils.Ptr(sourceToString(*in.Source))
	}
	out, err := s.fileUC.List(ctx, listIn)
	if err != nil {
		return nil, err
	}
	return &systempb.ListFilesResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, ToProtoFile),
	}, nil
}

func (s *FileService) GetFile(ctx context.Context, in *systempb.GetFileRequest) (*systempb.GetFileResponse, error) {
	f, err := s.fileUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetFileResponse{File: ToProtoFile(f)}, nil
}

func (s *FileService) GetFileUrls(ctx context.Context, in *systempb.GetFileUrlsRequest) (*systempb.GetFileUrlsResponse, error) {
	mints, err := s.fileUC.MintUrls(ctx, in.Ids)
	if err != nil {
		return nil, err
	}
	urls := make(map[string]*systempb.FileUrl, len(mints))
	for _, mt := range mints {
		urls[mt.ID] = &systempb.FileUrl{
			Url:       mt.URL,
			ExpiresAt: utils.ToTimestamp(&mt.ExpiresAt),
		}
	}
	return &systempb.GetFileUrlsResponse{Urls: urls}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func ToProtoFile(f *File) *systempb.File {
	return &systempb.File{
		Id:          utils.StringW(f.ID),
		CreatedAt:   utils.ToTimestamp(&f.CreatedAt),
		UpdatedAt:   utils.ToTimestamp(&f.UpdatedAt),
		Name:        utils.StringW(f.Name),
		ContentType: utils.StringWEmptyNil(f.ContentType),
		Size:        utils.Int64W(f.Size),
		Source:      toProtoSource(f.Source),
		Status:      toProtoStatus(f.Status),
		OwnerId:     utils.StringW(f.OwnerID),
	}
}

func toProtoStatus(s string) systempb.FileStatus {
	switch s {
	case StatusConfirmed:
		return systempb.FileStatus_FILE_STATUS_CONFIRMED
	case StatusProcessing:
		return systempb.FileStatus_FILE_STATUS_PROCESSING
	case StatusFailed:
		return systempb.FileStatus_FILE_STATUS_FAILED
	}
	return systempb.FileStatus_FILE_STATUS_UNSPECIFIED
}

func statusToString(s systempb.FileStatus) string {
	switch s {
	case systempb.FileStatus_FILE_STATUS_CONFIRMED:
		return StatusConfirmed
	case systempb.FileStatus_FILE_STATUS_PROCESSING:
		return StatusProcessing
	case systempb.FileStatus_FILE_STATUS_FAILED:
		return StatusFailed
	}
	return ""
}

func toProtoSource(s string) systempb.FileSource {
	if s == SourceServerGenerated {
		return systempb.FileSource_FILE_SOURCE_SERVER_GENERATED
	}
	return systempb.FileSource_FILE_SOURCE_CLIENT_UPLOAD
}

func sourceToString(s systempb.FileSource) string {
	if s == systempb.FileSource_FILE_SOURCE_SERVER_GENERATED {
		return SourceServerGenerated
	}
	return SourceClientUpload
}
