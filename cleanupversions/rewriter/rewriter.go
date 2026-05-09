// Package rewriter implements Phase 3 of cleanup-versions: applying
// the byte-level edits captured by the static phase to source files.
//
// Apply is atomic: it groups edits by file, applies them right-to-left
// (so earlier offsets don't shift), and writes the result via a
// temp-file + rename. If go/format fails on any output file, all files
// are restored from their pre-edit state.
package rewriter

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"

	"github.com/mattconzen/temporallint/cleanupversions"
)

// Apply applies the Edits from each Safe report to disk. Reports
// whose Decision != Safe are skipped silently; the caller is responsible
// for deciding whether to call Apply at all.
//
// Returns the list of file paths that were modified, plus any error.
// On error, all touched files are restored from their pre-edit content.
func Apply(reports []cleanupversions.SafetyReport) ([]string, error) {
	byFile := map[string][]cleanupversions.SafetyReport{}
	for _, r := range reports {
		if r.Decision != cleanupversions.DecisionSafe {
			continue
		}
		if len(r.Candidate.Edit.NewBytes) == 0 && r.Candidate.Edit.EndOffset == r.Candidate.Edit.StartOffset {
			continue
		}
		byFile[r.Candidate.File] = append(byFile[r.Candidate.File], r)
	}
	if len(byFile) == 0 {
		return nil, nil
	}

	originals := map[string][]byte{}
	var touched []string
	defer func() {
		// no-op when err is nil; caller handles restore via the error path
	}()
	for path, rs := range byFile {
		orig, err := os.ReadFile(path)
		if err != nil {
			return touched, fmt.Errorf("read %s: %w", path, err)
		}
		originals[path] = orig

		// Apply edits right-to-left (highest StartOffset first) so
		// earlier offsets stay valid.
		sort.Slice(rs, func(i, j int) bool {
			return rs[i].Candidate.Edit.StartOffset > rs[j].Candidate.Edit.StartOffset
		})
		modified := append([]byte(nil), orig...)
		for _, r := range rs {
			e := r.Candidate.Edit
			if e.StartOffset < 0 || e.EndOffset > len(modified) || e.StartOffset > e.EndOffset {
				return touched, fmt.Errorf("invalid edit range in %s: %d..%d (file size %d)", path, e.StartOffset, e.EndOffset, len(modified))
			}
			modified = append(modified[:e.StartOffset], append(e.NewBytes, modified[e.EndOffset:]...)...)
		}

		formatted, err := format.Source(modified)
		if err != nil {
			restoreAll(originals)
			return nil, fmt.Errorf("gofmt %s after rewrite: %w", path, err)
		}
		if err := writeFile(path, formatted); err != nil {
			restoreAll(originals)
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
		touched = append(touched, path)
	}
	sort.Strings(touched)
	return touched, nil
}

func writeFile(path string, content []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func restoreAll(originals map[string][]byte) {
	for path, orig := range originals {
		_ = os.WriteFile(path, orig, 0o644)
	}
}

// FilenameFor is exposed for tests so they can compute paths relative
// to a temp directory deterministically.
func FilenameFor(dir, name string) string { return filepath.Join(dir, name) }
