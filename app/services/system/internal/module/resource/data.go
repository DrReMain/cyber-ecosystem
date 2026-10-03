package resource

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type resourceRP struct {
	shared.RP

	fsys     fs.FS
	dirLabel string

	mu       sync.Mutex
	stamp    string
	services []*ServiceMeta
}

func NewResourceRP(logger *slog.Logger, p *platform.Platform, c *conf.Catalog) (ResourceRP, error) {
	if c == nil || c.GetDir() == "" {
		return nil, fmt.Errorf("catalog.dir is required")
	}
	rp := &resourceRP{
		RP:       shared.NewRP(logger.With("module", "module/resource_rp"), p),
		fsys:     os.DirFS(c.GetDir()),
		dirLabel: c.GetDir(),
	}
	if _, err := rp.load(); err != nil {
		return nil, err
	}
	return rp, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *resourceRP) ListResource(ctx context.Context) ([]*ServiceMeta, error) {
	stamp, err := rp.statStamp()
	if err != nil {
		return nil, err
	}
	rp.mu.Lock()
	defer rp.mu.Unlock()
	if stamp != rp.stamp {
		rp.stamp = stamp
		if _, err := rp.reload(); err != nil {
			return nil, err
		}
	}
	return rp.services, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *resourceRP) load() ([]*ServiceMeta, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.stamp, _ = rp.statStamp()
	return rp.reload()
}

func (rp *resourceRP) reload() ([]*ServiceMeta, error) {
	entries, err := fs.ReadDir(rp.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("catalog: read dir %s: %w", rp.dirLabel, err)
	}
	var services []*ServiceMeta
	seen := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := fs.ReadFile(rp.fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("catalog: read %s: %w", e.Name(), err)
		}
		batch, err := utils.Unmarshal[[]*ServiceMeta](raw)
		if err != nil {
			return nil, fmt.Errorf("catalog: parse %s: %w", e.Name(), err)
		}
		for _, s := range batch {
			if prev, dup := seen[s.FullName]; dup {
				return nil, fmt.Errorf("catalog: service %s declared in both %s and %s", s.FullName, prev, e.Name())
			}
			seen[s.FullName] = e.Name()
			services = append(services, s)
		}
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("catalog: no manifests under %s", rp.dirLabel)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].FullName < services[j].FullName })
	rp.services = services
	return services, nil
}

func (rp *resourceRP) statStamp() (string, error) {
	entries, err := fs.ReadDir(rp.fsys, ".")
	if err != nil {
		return "", fmt.Errorf("catalog: read dir %s: %w", rp.dirLabel, err)
	}
	var b strings.Builder
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return "", fmt.Errorf("catalog: stat %s: %w", e.Name(), err)
		}
		fmt.Fprintf(&b, "%s:%d:%d;", e.Name(), info.ModTime().UnixNano(), info.Size())
	}
	return b.String(), nil
}
