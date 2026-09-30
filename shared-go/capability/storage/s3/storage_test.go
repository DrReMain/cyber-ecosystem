package s3

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// newTestBackend builds a backend around a zero-value client. Tests using it
// only exercise pre-network paths (argument guards, local assembly); anything
// reaching the network would fail on the missing endpoint.
func newTestBackend() *backend {
	client := s3.New(s3.Options{})
	return &backend{client: client, presigner: s3.NewPresignClient(client), cfg: Config{Bucket: "t"}}
}

func TestNewResolvesLimits(t *testing.T) {
	client := s3.New(s3.Options{})

	zero := New(client, nil, Config{Bucket: "t"}).Limits()
	if zero.MultipartThreshold != defaultMultipartThreshold ||
		zero.PartSize != defaultPartSize ||
		zero.PresignTTL != defaultPresignTTL {
		t.Errorf("zero config: limits not defaulted: %+v", zero)
	}
	if zero.MaxFileSize != 0 {
		t.Errorf("zero config: MaxFileSize must stay 0 (no limit), got %d", zero.MaxFileSize)
	}

	explicit := New(client, nil, Config{
		Bucket:             "t",
		MaxFileSize:        1 << 20,
		MultipartThreshold: 16 << 20,
		PartSize:           8 << 20,
		PresignTTL:         time.Minute,
	}).Limits()
	if explicit.MaxFileSize != 1<<20 ||
		explicit.MultipartThreshold != 16<<20 ||
		explicit.PartSize != 8<<20 ||
		explicit.PresignTTL != time.Minute {
		t.Errorf("explicit config: limits mismatch: %+v", explicit)
	}
}

func TestForReturnsBucketView(t *testing.T) {
	s := New(s3.New(s3.Options{}), nil, Config{Bucket: "t"})
	if v := s.For("tenant-1"); v == nil || v.Object == nil || v.Presign == nil || v.Multipart == nil || v.List == nil {
		t.Fatal("For returned a nil or incomplete view")
	}
}
