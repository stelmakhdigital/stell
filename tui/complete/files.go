package complete

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var skipDir = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, "bin": {},
	".idea": {}, ".vscode": {}, "dist": {}, "coverage": {},
}

const popupLimit = 20

// FileIndex is a cached workspace path list.
type FileIndex struct {
	mu      sync.Mutex
	root    string
	paths   []string
	loaded  time.Time
	ttl     time.Duration
	limit   int
	ready   bool
	loading bool
}

// NewFileIndex indexes root. Walk is bounded (limit) and skips heavy dirs.
func NewFileIndex(root string) *FileIndex {
	if root == "" {
		root, _ = os.Getwd()
	}
	return &FileIndex{root: root, ttl: 3 * time.Second, limit: 2000, paths: []string{}}
}

// Fresh reports a usable cache within TTL.
func (f *FileIndex) Fresh() bool {
	if f == nil {
		return true
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ready && time.Since(f.loaded) < f.ttl
}

// StartRebuild claims the walk. Caller must Rebuild if this returns true.
func (f *FileIndex) StartRebuild() bool {
	if f == nil {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loading {
		return false
	}
	if f.ready && time.Since(f.loaded) < f.ttl {
		return false
	}
	f.loading = true
	return true
}

// Rebuild walks the workspace without holding the lock during WalkDir.
func (f *FileIndex) Rebuild() {
	if f == nil {
		return
	}
	f.mu.Lock()
	root, limit := f.root, f.limit
	f.mu.Unlock()

	var paths []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if _, skip := skipDir[name]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		paths = append(paths, filepath.ToSlash(rel))
		if len(paths) >= limit {
			return fs.SkipAll
		}
		return nil
	})
	if paths == nil {
		paths = []string{}
	}
	f.mu.Lock()
	f.paths = paths
	f.loaded = time.Now()
	f.ready = true
	f.loading = false
	f.mu.Unlock()
}

// Match filters the current snapshot. Does not walk.
func (f *FileIndex) Match(query string) []Item {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	paths := append([]string(nil), f.paths...)
	f.mu.Unlock()
	return matchFiles(paths, query)
}

// Files completes a path/query (after optional @). Sync walk for tests.
func (f *FileIndex) Files(query string) []Item {
	q := strings.TrimPrefix(query, "@")
	q = strings.TrimPrefix(q, "./")
	if f == nil {
		return nil
	}
	if !f.Fresh() {
		f.Rebuild()
	}
	return f.Match(q)
}

func matchFiles(paths []string, query string) []Item {
	q := strings.TrimPrefix(query, "@")
	q = strings.TrimPrefix(q, "./")
	ql := strings.ToLower(q)
	type hit struct {
		p    string
		rank int
	}
	var hits []hit
	for _, p := range paths {
		pl := strings.ToLower(p)
		base := strings.ToLower(filepath.Base(p))
		var rank int
		switch {
		case ql == "":
			rank = 2
		case strings.HasPrefix(base, ql) || strings.HasPrefix(pl, ql):
			rank = 0
		case strings.Contains(pl, ql):
			rank = 1
		default:
			continue
		}
		hits = append(hits, hit{p: p, rank: rank})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].rank != hits[j].rank {
			return hits[i].rank < hits[j].rank
		}
		return hits[i].p < hits[j].p
	})
	if len(hits) > popupLimit {
		hits = hits[:popupLimit]
	}
	out := make([]Item, len(hits))
	for i, h := range hits {
		out[i] = Item{Value: h.p, Label: h.p, Kind: "file"}
	}
	return out
}
