package filepresign

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const maxParts int64 = 10000

// DO ------------------------------------------------------------------------------------------------------------------

type CreateIn struct {
	Name        string
	ContentType string
	Size        int64
}

type PartETag struct {
	PartNumber int32
	ETag       string
}

type PartURL struct {
	PartNumber int32
	URL        string
}

type SingleMint struct {
	URL       string
	ExpiresAt time.Time
}

type SessionMint struct {
	UploadID  string
	PartSize  int64
	PartCount int32
	PartURLs  []PartURL
	ExpiresAt time.Time
}

type CreateOut struct {
	File      *file.File
	Single    *SingleMint
	Multipart *SessionMint
}

type ResumeOut struct {
	File      *file.File
	PartSize  int64
	Uploaded  []storage.CompletedPart
	Missing   []PartURL
	ExpiresAt time.Time
}

// Port ----------------------------------------------------------------------------------------------------------------

type FilePresignRP interface {
	Create(ctx context.Context, f *file.File) (*file.File, error)
	MarkConfirmed(ctx context.Context, id string, size int64, contentType, etag string) (*file.File, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*file.File, error)
	Limits() storage.Limits
	PresignUpload(ctx context.Context, key string, size int64, ttl time.Duration) (string, error)
	PresignUploadPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error)
	CreateMultipart(ctx context.Context, key, contentType string) (string, error)
	ListParts(ctx context.Context, key, uploadID string) ([]storage.CompletedPart, error)
	CompleteMultipart(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error
	AbortMultipart(ctx context.Context, key, uploadID string) error
	Stat(ctx context.Context, key string) (*storage.ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type FilePresignUC struct {
	shared.UC
	fileRP FilePresignRP
}

func NewFilePresignUC(logger *slog.Logger, tm shared.Transaction, fileRP FilePresignRP) *FilePresignUC {
	return &FilePresignUC{
		UC:     shared.NewUC(logger.With("module", "module/filepresign"), tm),
		fileRP: fileRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *FilePresignUC) Create(ctx context.Context, in *CreateIn) (*CreateOut, error) {
	subjectID, err := subjectUserID(ctx)
	if err != nil {
		return nil, err
	}
	limits := uc.fileRP.Limits()
	if limits.MaxFileSize > 0 && in.Size > limits.MaxFileSize {
		return nil, systempb.ErrorSystemFileTooLarge("")
	}
	f := &file.File{
		Key:         newKey(),
		Name:        in.Name,
		ContentType: in.ContentType,
		Size:        in.Size,
		Source:      file.SourceClientUpload,
		Status:      file.StatusUploading,
		OwnerID:     subjectID,
	}
	ttl := limits.PresignTTL
	expiresAt := time.Now().Add(ttl)
	if in.Size < limits.MultipartThreshold {
		url, err := uc.fileRP.PresignUpload(ctx, f.Key, in.Size, ttl)
		if err != nil {
			return nil, err
		}
		created, err := uc.fileRP.Create(ctx, f)
		if err != nil {
			return nil, err
		}
		return &CreateOut{File: created, Single: &SingleMint{URL: url, ExpiresAt: expiresAt}}, nil
	}
	uploadID, err := uc.fileRP.CreateMultipart(ctx, f.Key, f.ContentType)
	if err != nil {
		return nil, err
	}
	f.UploadID = uploadID
	created, err := uc.fileRP.Create(ctx, f)
	if err != nil {
		if aErr := uc.fileRP.AbortMultipart(ctx, f.Key, uploadID); aErr != nil {
			uc.Log.Warn("presign create orphan cleanup failed", "key", f.Key, "error", aErr)
		}
		return nil, err
	}
	partSize, partCount := planOf(limits, in.Size)
	partURLs := make([]PartURL, 0, partCount)
	for n := int32(1); n <= partCount; n++ {
		url, err := uc.fileRP.PresignUploadPart(ctx, f.Key, uploadID, n, ttl)
		if err != nil {
			return nil, err
		}
		partURLs = append(partURLs, PartURL{PartNumber: n, URL: url})
	}
	return &CreateOut{
		File:      created,
		Multipart: &SessionMint{UploadID: uploadID, PartSize: partSize, PartCount: partCount, PartURLs: partURLs, ExpiresAt: expiresAt},
	}, nil
}

func (uc *FilePresignUC) Confirm(ctx context.Context, id string, parts []PartETag) (*file.File, error) {
	subjectID, err := subjectUserID(ctx)
	if err != nil {
		return nil, err
	}
	f, err := uc.getOwned(ctx, subjectID, id)
	if err != nil {
		return nil, err
	}
	if f.Status == file.StatusConfirmed {
		return f, nil
	}
	obj, err := uc.fileRP.Stat(ctx, f.Key)
	if err != nil {
		if !errorspb.IsInfraErrorStorageNotFound(err) {
			return nil, err
		}
		if f.UploadID == "" || len(parts) == 0 {
			return nil, systempb.ErrorSystemFileInvalidState("")
		}
		if obj, err = uc.completeSession(ctx, f, parts); err != nil {
			return nil, err
		}
	}
	if obj.Size != f.Size {
		return nil, systempb.ErrorSystemFileSizeMismatch("")
	}
	return uc.fileRP.MarkConfirmed(ctx, id, obj.Size, obj.ContentType, obj.ETag)
}

func (uc *FilePresignUC) Abort(ctx context.Context, id string) error {
	subjectID, err := subjectUserID(ctx)
	if err != nil {
		return err
	}
	f, err := uc.getOwned(ctx, subjectID, id)
	if err != nil {
		return err
	}
	if f.Status != file.StatusUploading {
		return systempb.ErrorSystemFileInvalidState("")
	}
	if f.UploadID != "" {
		if err := uc.fileRP.AbortMultipart(ctx, f.Key, f.UploadID); err != nil {
			return err
		}
	} else if err := uc.fileRP.DeleteObject(ctx, f.Key); err != nil {
		return err
	}
	return uc.fileRP.Delete(ctx, id)
}

func (uc *FilePresignUC) ListParts(ctx context.Context, id string) (*ResumeOut, error) {
	subjectID, err := subjectUserID(ctx)
	if err != nil {
		return nil, err
	}
	f, err := uc.getOwned(ctx, subjectID, id)
	if err != nil {
		return nil, err
	}
	if f.Status != file.StatusUploading {
		return nil, systempb.ErrorSystemFileInvalidState("")
	}
	limits := uc.fileRP.Limits()
	expiresAt := time.Now().Add(limits.PresignTTL)
	// A single-PUT session is the degenerate one-part plan (partSize = whole
	// size): there is no part state to list — resume either confirms in place
	// when the object already landed, or retries the whole object on a
	// re-minted URL.
	if f.UploadID == "" {
		obj, err := uc.fileRP.Stat(ctx, f.Key)
		if err != nil {
			if !errorspb.IsInfraErrorStorageNotFound(err) {
				return nil, err
			}
			url, uErr := uc.fileRP.PresignUpload(ctx, f.Key, f.Size, limits.PresignTTL)
			if uErr != nil {
				return nil, uErr
			}
			return &ResumeOut{
				File:      f,
				PartSize:  f.Size,
				Missing:   []PartURL{{PartNumber: 1, URL: url}},
				ExpiresAt: expiresAt,
			}, nil
		}
		return &ResumeOut{
			File:      f,
			PartSize:  f.Size,
			Uploaded:  []storage.CompletedPart{{PartNumber: 1, ETag: obj.ETag, Size: obj.Size}},
			ExpiresAt: expiresAt,
		}, nil
	}
	uploaded, err := uc.fileRP.ListParts(ctx, f.Key, f.UploadID)
	if err != nil {
		return nil, err
	}
	partSize, partCount := planOf(limits, f.Size)
	uploadedSet := make(map[int32]struct{}, len(uploaded))
	for _, p := range uploaded {
		if misalignedPart(p, partSize, partCount) {
			return nil, systempb.ErrorSystemFileInvalidState("")
		}
		uploadedSet[p.PartNumber] = struct{}{}
	}
	missing := make([]PartURL, 0, partCount-utils.ConvNum[int32](len(uploadedSet)))
	for n := int32(1); n <= partCount; n++ {
		if _, ok := uploadedSet[n]; ok {
			continue
		}
		url, err := uc.fileRP.PresignUploadPart(ctx, f.Key, f.UploadID, n, limits.PresignTTL)
		if err != nil {
			return nil, err
		}
		missing = append(missing, PartURL{PartNumber: n, URL: url})
	}
	return &ResumeOut{
		File:      f,
		PartSize:  partSize,
		Uploaded:  uploaded,
		Missing:   missing,
		ExpiresAt: expiresAt,
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *FilePresignUC) completeSession(ctx context.Context, f *file.File, parts []PartETag) (*storage.ObjectInfo, error) {
	uploaded, err := uc.fileRP.ListParts(ctx, f.Key, f.UploadID)
	if err != nil {
		return nil, err
	}
	partSize, partCount := planOf(uc.fileRP.Limits(), f.Size)
	truth := make(map[int32]storage.CompletedPart, len(uploaded))
	for _, p := range uploaded {
		if misalignedPart(p, partSize, partCount) {
			return nil, systempb.ErrorSystemFileInvalidState("")
		}
		truth[p.PartNumber] = p
	}
	claimed := make(map[int32]string, len(parts))
	for _, p := range parts {
		claimed[p.PartNumber] = p.ETag
	}
	if utils.ConvNum[int32](len(truth)) != partCount || len(parts) != int(partCount) ||
		utils.ConvNum[int32](len(claimed)) != partCount {
		return nil, systempb.ErrorSystemFileInvalidState("")
	}
	completed := make([]storage.CompletedPart, 0, partCount)
	for n := int32(1); n <= partCount; n++ {
		t, ok := truth[n]
		if !ok || claimed[n] != t.ETag {
			return nil, systempb.ErrorSystemFileInvalidState("")
		}
		completed = append(completed, t)
	}
	if err := uc.fileRP.CompleteMultipart(ctx, f.Key, f.UploadID, completed); err != nil {
		// A concurrent confirm or sweep already closed the session — the
		// object on the backend is the winner's, so re-Stat and let the size
		// check judge it.
		if !errorspb.IsInfraErrorStorageNotFound(err) {
			return nil, err
		}
		obj, statErr := uc.fileRP.Stat(ctx, f.Key)
		if statErr != nil {
			if errorspb.IsInfraErrorStorageNotFound(statErr) {
				return nil, systempb.ErrorSystemFileInvalidState("")
			}
			return nil, statErr
		}
		return obj, nil
	}
	return uc.fileRP.Stat(ctx, f.Key)
}

func (uc *FilePresignUC) getOwned(ctx context.Context, userID, id string) (*file.File, error) {
	f, err := uc.fileRP.FindByID(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemFileNotFound("")
		}
		return nil, err
	}
	// NOT_FOUND rather than FORBIDDEN on a foreign row: ids are unguessable
	// xids and the answer leaks no existence either way.
	if f.OwnerID != userID {
		return nil, systempb.ErrorSystemFileNotFound("")
	}
	return f, nil
}

func planOf(limits storage.Limits, size int64) (int64, int32) {
	partSize := effectivePartSize(limits, size)
	return partSize, utils.ConvNum[int32](ceilDiv(size, partSize))
}

func misalignedPart(p storage.CompletedPart, partSize int64, partCount int32) bool {
	return p.PartNumber < 1 || p.PartNumber > partCount ||
		(p.Size != partSize && p.PartNumber != partCount)
}

func newKey() string {
	return "f/" + xid.New().String()
}

func effectivePartSize(limits storage.Limits, size int64) int64 {
	// maxParts mirrors the S3 multipart cap — raise the floor so a plan
	// never exceeds it.
	if minSize := ceilDiv(size, maxParts); minSize > limits.PartSize {
		return minSize
	}
	return limits.PartSize
}

func ceilDiv(a, b int64) int64 {
	return (a + b - 1) / b
}

func subjectUserID(ctx context.Context) (string, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	return subject.UserID, nil
}
