package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/kratos/security"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

const (
	StatusUploading  = "uploading"
	StatusConfirmed  = "confirmed"
	StatusProcessing = "processing"
	StatusFailed     = "failed"

	SourceClientUpload    = "client_upload"
	SourceServerGenerated = "server_generated"
)

const (
	generationTimeout   = 30 * time.Minute
	generationDrainWait = 10 * time.Second
	cleanupTimeout      = 30 * time.Second

	sweepInterval   = 7 * 24 * time.Hour
	sweepBatch      = 100
	sweepTTLFactor  = 4
	sweepKeyPrefix  = "f/"
	retentionPeriod = 7 * 24 * time.Hour

	// Beyond any live writer's max PROCESSING lifetime (generationTimeout +
	// cleanupTimeout), with slack for scheduler slop and clock skew — older
	// rows belong to a dead writer.
	staleGenerationAfter = generationTimeout + cleanupTimeout + 5*time.Minute
	recoveryInterval     = 5 * time.Minute

	// Reviewed non-empty factory message: the ambient-tx guard fires only on
	// developer misuse, where the message is the whole diagnosis.
	errGenerateAmbientTx = "Generate must run outside a transaction; put domain rows in the record callback"
)

type countingReader struct {
	r io.Reader
	n int64
}

// DO ------------------------------------------------------------------------------------------------------------------

type File struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TenantID    string
	Key         string
	Name        string
	ContentType string
	Size        int64
	Source      string
	Status      string
	UploadID    string
	ETag        string
	OwnerID     string
}

type FileListIn struct {
	*commonpb.PageRequest
	OrderBy     []string
	Name        *string
	ContentType *string
	Status      *string
	Source      *string
}

type FileListOut struct {
	*commonpb.PageResponse
	List []*File
}

type UrlMint struct {
	ID        string
	URL       string
	ExpiresAt time.Time
}

type GenerationSpec struct {
	Name        string
	ContentType string
	Produce     func(ctx context.Context, w io.Writer) error
}

// Port ----------------------------------------------------------------------------------------------------------------

type FileRP interface {
	UpdateName(ctx context.Context, id, name string) (*File, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, in *FileListIn) (*FileListOut, error)
	FindByID(ctx context.Context, id string) (*File, error)
	FindRef(ctx context.Context, id string) (*File, error)
	Limits() storage.Limits
	PresignDownload(ctx context.Context, key string, ttl time.Duration, opts storage.DownloadOptions) (string, error)
	CreateGeneration(ctx context.Context, f *File) (*File, error)
	UploadObject(ctx context.Context, key, contentType string, r io.Reader) (*storage.ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
	MarkGenerated(ctx context.Context, id string, size int64, etag string) error
	MarkFailed(ctx context.Context, id string) error
	FailStaleProcessing(ctx context.Context, createdBefore time.Time) (int, error)
	ListStaleUploading(ctx context.Context, createdBefore time.Time, afterID string, limit int) ([]*File, error)
	ClaimUploading(ctx context.Context, id string, createdBefore time.Time) (bool, error)
	FindByKey(ctx context.Context, key string) (*File, error)
	ListMultipartUploads(ctx context.Context, prefix string, visit func(storage.PendingUpload) error) error
	AbortSession(ctx context.Context, key, uploadID string) error
	ListObjects(ctx context.Context, prefix, cursor string, maxKeys int32) (*storage.ListResult, error)
	FindRefByKey(ctx context.Context, key string) (*File, error)
	ListSoftDeleted(ctx context.Context, deletedBefore time.Time, afterID string, limit int) ([]*File, error)
	ReclaimDeleted(ctx context.Context, id string, deletedBefore time.Time) (bool, error)
	InAmbientTx(ctx context.Context) bool
}

// UC ------------------------------------------------------------------------------------------------------------------

type FileUC struct {
	shared.UC
	fileRP FileRP

	genCtx    context.Context
	genCancel context.CancelFunc
	genWG     sync.WaitGroup

	sweepCtx    context.Context
	sweepCancel context.CancelFunc
	sweepWG     sync.WaitGroup
}

func NewFileUC(logger *slog.Logger, tm shared.Transaction, fileRP FileRP, lc shared.HookRegistry) *FileUC {
	genCtx, genCancel := context.WithCancel(context.Background())
	sweepCtx, sweepCancel := context.WithCancel(context.Background())
	uc := &FileUC{
		UC:          shared.NewUC(logger.With("module", "module/file"), tm),
		fileRP:      fileRP,
		genCtx:      genCtx,
		genCancel:   genCancel,
		sweepCtx:    sweepCtx,
		sweepCancel: sweepCancel,
	}
	lc.OnStart(uc.StartSweep)
	lc.OnStop(uc.ShutdownGenerations)
	lc.OnStop(uc.StopSweep)
	return uc
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *FileUC) Update(ctx context.Context, id, name string) (*File, error) {
	if _, err := uc.Get(ctx, id); err != nil {
		return nil, err
	}
	return uc.fileRP.UpdateName(ctx, id, name)
}

func (uc *FileUC) Delete(ctx context.Context, id string) error {
	f, err := uc.Get(ctx, id)
	if err != nil {
		return err
	}
	// Only terminal rows retire here. UPLOADING's exit is abort (upload
	// channel lifecycle), PROCESSING's is the generation task itself; a FAILED
	// row is deletable as a record whose session leak the sweeper owns.
	if f.Status != StatusConfirmed && f.Status != StatusFailed {
		return systempb.ErrorSystemFileInvalidState("")
	}
	return uc.fileRP.Delete(ctx, id)
}

func (uc *FileUC) List(ctx context.Context, in *FileListIn) (*FileListOut, error) {
	return uc.fileRP.List(ctx, in)
}

func (uc *FileUC) Get(ctx context.Context, id string) (*File, error) {
	f, err := uc.fileRP.FindByID(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemFileNotFound("")
		}
		return nil, err
	}
	return f, nil
}

func (uc *FileUC) MintUrls(ctx context.Context, ids []string) ([]*UrlMint, error) {
	ttl := uc.fileRP.Limits().PresignTTL
	mints := make([]*UrlMint, 0, len(ids))
	for _, id := range ids {
		f, err := uc.fileRP.FindRef(ctx, id)
		if err != nil {
			if errorspb.IsInfraErrorDbNotFound(err) {
				continue
			}
			return nil, err
		}
		if f.Status != StatusConfirmed {
			continue
		}
		url, err := uc.fileRP.PresignDownload(ctx, f.Key, ttl, storage.DownloadOptions{
			ContentType:        f.ContentType,
			ContentDisposition: fmt.Sprintf("inline; filename=%q", f.Name),
		})
		if err != nil {
			return nil, err
		}
		mints = append(mints, &UrlMint{ID: id, URL: url, ExpiresAt: time.Now().Add(ttl)})
	}
	return mints, nil
}

func (uc *FileUC) Generate(ctx context.Context, spec *GenerationSpec, record func(context.Context, *File) error) (*File, error) {
	// Fail loud, not late: InTx joins an ambient tx, so a caller-side tx
	// would have the task launch before that tx commits — writing against a
	// row that may never exist. Domain rows ride the record callback.
	if uc.fileRP.InAmbientTx(ctx) {
		return nil, errorspb.ErrorGeneralErrorInternal(errGenerateAmbientTx)
	}
	ownerID, err := subjectUserID(ctx)
	if err != nil {
		return nil, err
	}
	var f *File
	if err := uc.Tm.InTx(ctx, func(tctx context.Context) error {
		created, err := uc.fileRP.CreateGeneration(tctx, &File{
			Key:         newKey(),
			Name:        spec.Name,
			ContentType: spec.ContentType,
			Source:      SourceServerGenerated,
			Status:      StatusProcessing,
			OwnerID:     ownerID,
		})
		if err != nil {
			return err
		}
		if record != nil {
			if err := record(tctx, created); err != nil {
				return err
			}
		}
		f = created
		return nil
	}); err != nil {
		return nil, err
	}
	// Post-commit launch: the task must not write against a rolled-back row.
	// The entry guard rejects ambient transactions, so the row is committed
	// by the time InTx returns here.
	uc.runGeneration(f, spec)
	return f, nil
}

func (uc *FileUC) ShutdownGenerations(context.Context) error {
	// Cancel first so in-flight uploads abort cleanly instead of leaking
	// half objects, then wait a bounded time for the drain.
	uc.genCancel()
	done := make(chan struct{})
	go func() {
		uc.genWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(generationDrainWait):
		uc.Log.Warn("generation tasks did not drain in time")
	}
	return nil
}

func (uc *FileUC) StartSweep(ctx context.Context) error {
	uc.sweepWG.Add(2)
	go uc.sweepLoop()
	go uc.recoveryLoop()
	return nil
}

func (uc *FileUC) StopSweep(context.Context) error {
	// Bounded by each item's ctx check; an abort is idempotent, so losing a
	// partially finished pass to shutdown costs one item at most.
	uc.sweepCancel()
	uc.sweepWG.Wait()
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *FileUC) runGeneration(f *File, spec *GenerationSpec) {
	uc.genWG.Go(func() {
		// genCtx is cancelled at shutdown; the timeout caps a runaway job.
		ctx, cancel := context.WithTimeout(uc.genCtx, generationTimeout)
		defer cancel()
		if err := uc.generate(ctx, f, spec); err != nil {
			uc.Log.Warn("file generation failed", "id", f.ID, "key", f.Key, "error", err)
		}
	})
}

func (uc *FileUC) generate(ctx context.Context, f *File, spec *GenerationSpec) error {
	pr, pw := io.Pipe()
	go func() {
		// nil → EOF for the reader; an error propagates through Upload.
		pw.CloseWithError(spec.Produce(ctx, pw))
	}()
	// Count bytes here: Upload backfills Size only when the backend enforces
	// MaxFileSize; row truth must not depend on conf.
	body := &countingReader{r: pr}
	info, err := uc.fileRP.UploadObject(ctx, f.Key, spec.ContentType, body)
	if err == nil {
		err = uc.fileRP.MarkGenerated(ctx, f.ID, body.n, info.ETag)
	}
	if err == nil {
		return nil
	}
	// Cleanup outlives the failure cause — a dead ctx must not block it.
	cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
	defer cancel()
	if dErr := uc.fileRP.DeleteObject(cctx, f.Key); dErr != nil {
		uc.Log.Warn("generation cleanup failed", "id", f.ID, "key", f.Key, "error", dErr)
	}
	if mErr := uc.fileRP.MarkFailed(cctx, f.ID); mErr != nil {
		return errors.Join(mErr, err)
	}
	return err
}

func newKey() string {
	return "f/" + xid.New().String()
}

func (uc *FileUC) sweepLoop() {
	defer uc.sweepWG.Done()
	// First pass runs immediately: interval semantics catch up after downtime,
	// unlike a fixed clock time that would silently miss its window.
	uc.runSweep(uc.sweepCtx)
	t := time.NewTicker(sweepInterval)
	defer t.Stop()
	for {
		select {
		case <-uc.sweepCtx.Done():
			return
		case <-t.C:
			uc.runSweep(uc.sweepCtx)
		}
	}
}

func (uc *FileUC) recoveryLoop() {
	defer uc.sweepWG.Done()
	// Same catch-up semantics as the sweep loop, but bounded by age — never by
	// boot assumptions: a live replica may be mid-generation elsewhere.
	uc.runRecovery(uc.sweepCtx)
	t := time.NewTicker(recoveryInterval)
	defer t.Stop()
	for {
		select {
		case <-uc.sweepCtx.Done():
			return
		case <-t.C:
			uc.runRecovery(uc.sweepCtx)
		}
	}
}

func (uc *FileUC) runRecovery(ctx context.Context) {
	cut := time.Now().Add(-staleGenerationAfter)
	n, err := uc.fileRP.FailStaleProcessing(ctx, cut)
	if err != nil {
		uc.Log.Warn("generation recovery: fail stale processing", "error", err)
		return
	}
	if n > 0 {
		uc.Log.Info("recovered stale generations", "count", n)
	}
}

func (uc *FileUC) runSweep(ctx context.Context) {
	cut := time.Now().Add(-time.Duration(sweepTTLFactor) * uc.fileRP.Limits().PresignTTL)
	rows := uc.sweepStaleRows(ctx, cut)
	sessions := uc.sweepOrphanSessions(ctx)
	objects := uc.sweepOrphanObjects(ctx, cut)
	reclaimed := uc.sweepExpiredDeleted(ctx, time.Now().Add(-retentionPeriod))
	if rows+sessions+objects+reclaimed > 0 {
		uc.Log.Info("file sweep", "rows", rows, "sessions", sessions, "objects", objects, "reclaimed", reclaimed)
	}
}

func (uc *FileUC) sweepStaleRows(ctx context.Context, cut time.Time) int {
	claimed := 0
	afterID := ""
	for {
		if ctx.Err() != nil {
			return claimed
		}
		rows, err := uc.fileRP.ListStaleUploading(ctx, cut, afterID, sweepBatch)
		if err != nil {
			uc.Log.Warn("file sweep: list stale rows", "error", err)
			return claimed
		}
		for _, f := range rows {
			if ctx.Err() != nil {
				return claimed
			}
			// Claim-first: the conditional hard delete is the fence. Losing it
			// means confirm/abort just won the row, and the S3 side must not be
			// touched — a confirmed object would otherwise be destroyed.
			ok, err := uc.fileRP.ClaimUploading(ctx, f.ID, cut)
			if err != nil {
				uc.Log.Warn("file sweep: claim", "id", f.ID, "error", err)
				continue
			}
			if !ok {
				continue
			}
			claimed++
			if f.UploadID != "" {
				if err := uc.fileRP.AbortSession(ctx, f.Key, f.UploadID); err != nil {
					uc.Log.Warn("file sweep: abort session", "key", f.Key, "error", err)
				}
			} else if err := uc.fileRP.DeleteObject(ctx, f.Key); err != nil {
				uc.Log.Warn("file sweep: delete object", "key", f.Key, "error", err)
			}
		}
		if len(rows) < sweepBatch {
			return claimed
		}
		afterID = rows[len(rows)-1].ID
	}
}

func (uc *FileUC) sweepOrphanSessions(ctx context.Context) int {
	aborted := 0
	err := uc.fileRP.ListMultipartUploads(ctx, sweepKeyPrefix, func(u storage.PendingUpload) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		f, err := uc.fileRP.FindByKey(ctx, u.Key)
		if err != nil {
			return err
		}
		// Row-join, not age: this backend omits Initiated. A live row owns the
		// session (pass 1 for uploading, its owner for processing/confirmed);
		// no row — or a terminal failed row — means leaked.
		if f != nil && (f.Status == StatusUploading || f.Status == StatusProcessing || f.Status == StatusConfirmed) {
			return nil
		}
		if err := uc.fileRP.AbortSession(ctx, u.Key, u.UploadID); err != nil {
			uc.Log.Warn("file sweep: abort orphan session", "key", u.Key, "error", err)
			return nil
		}
		aborted++
		return nil
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		uc.Log.Warn("file sweep: list sessions", "error", err)
	}
	return aborted
}

func (uc *FileUC) sweepOrphanObjects(ctx context.Context, cut time.Time) int {
	deleted := 0
	cursor := ""
	for {
		if ctx.Err() != nil {
			return deleted
		}
		res, err := uc.fileRP.ListObjects(ctx, sweepKeyPrefix, cursor, sweepBatch)
		if err != nil {
			uc.Log.Warn("file sweep: list objects", "error", err)
			return deleted
		}
		for _, obj := range res.Objects {
			if ctx.Err() != nil {
				return deleted
			}
			// Unlike sessions, objects carry a real timestamp. The grace
			// window only needs to cover the proxy path where the object
			// lands before its row exists; direct uploads mint the row first.
			if !obj.LastModified.Before(cut) {
				continue
			}
			// Unscoped lookup: a soft-deleted row still owns its object until
			// the retention pass reclaims it — rowless-object deletion must
			// not bypass the retention window.
			f, err := uc.fileRP.FindRefByKey(ctx, obj.Key)
			if err != nil {
				uc.Log.Warn("file sweep: find by key", "key", obj.Key, "error", err)
				return deleted
			}
			if f != nil {
				continue
			}
			if err := uc.fileRP.DeleteObject(ctx, obj.Key); err != nil {
				uc.Log.Warn("file sweep: delete orphan object", "key", obj.Key, "error", err)
				continue
			}
			deleted++
		}
		// Key-ordered cursor: deleting already-paged keys never shifts the
		// remaining sequence.
		if res.NextCursor == "" {
			return deleted
		}
		cursor = res.NextCursor
	}
}

func (uc *FileUC) sweepExpiredDeleted(ctx context.Context, cut time.Time) int {
	reclaimed := 0
	afterID := ""
	for {
		if ctx.Err() != nil {
			return reclaimed
		}
		rows, err := uc.fileRP.ListSoftDeleted(ctx, cut, afterID, sweepBatch)
		if err != nil {
			uc.Log.Warn("file sweep: list deleted rows", "error", err)
			return reclaimed
		}
		for _, f := range rows {
			if ctx.Err() != nil {
				return reclaimed
			}
			// Claim-first again: the conditional hard delete owns the row
			// first, so a failed object delete afterwards only leaves a
			// rowless object — exactly what the orphan-object pass reclaims
			// next round.
			ok, err := uc.fileRP.ReclaimDeleted(ctx, f.ID, cut)
			if err != nil {
				uc.Log.Warn("file sweep: reclaim", "id", f.ID, "error", err)
				continue
			}
			if !ok {
				continue
			}
			reclaimed++
			if err := uc.fileRP.DeleteObject(ctx, f.Key); err != nil {
				uc.Log.Warn("file sweep: delete reclaimed object", "key", f.Key, "error", err)
			}
		}
		if len(rows) < sweepBatch {
			return reclaimed
		}
		afterID = rows[len(rows)-1].ID
	}
}

func subjectUserID(ctx context.Context) (string, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	return subject.UserID, nil
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}
