package biz

import (
	"context"
	"io"

	"github.com/go-kratos/kratos/v2/log"
)

// region[rgba(66,165,245,0.15)] 🔵 Port --------------------------------------------------------------------------------

type FileRP interface {
	Upload(ctx context.Context, name string, reader io.Reader, size int64, contentType string) (*File, error)
	Download(ctx context.Context, id string) (io.ReadCloser, *File, error)
	Delete(ctx context.Context, id string) error
}

// region[rgba(102,187,106,0.15)] 🟢 UC ----------------------------------------------------------------------------------

type FileUC struct {
	UC
	fileRP FileRP
}

func NewFileUC(logger log.Logger, tm Transaction, fileRP FileRP) *FileUC {
	return &FileUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/file_uc")),
			tm:  tm,
		},
		fileRP: fileRP,
	}
}

// region[rgba(186,104,200,0.15)] 🟣 Method ------------------------------------------------------------------------------

func (uc *FileUC) Upload(ctx context.Context, name string, reader io.Reader, size int64, contentType string) (*File, error) {
	return uc.fileRP.Upload(ctx, name, reader, size, contentType)
}

func (uc *FileUC) Download(ctx context.Context, id string) (io.ReadCloser, *File, error) {
	return uc.fileRP.Download(ctx, id)
}

func (uc *FileUC) Delete(ctx context.Context, id string) error {
	return uc.fileRP.Delete(ctx, id)
}
