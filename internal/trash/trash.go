// Package trash moves paths to the macOS Trash (Finder Bin) so deletions are
// undoable via "Put Back". It never permanently removes user data and refuses
// SIP-protected paths.
package trash

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thaitrn/binman/internal/safety"
)

// ErrForbidden is returned for a path that is SIP-protected.
var ErrForbidden = errors.New("SIP-protected path; refusing to delete")

// Result is the outcome of trashing one path.
type Result struct {
	Path string
	Err  error
}

// Trash moves paths to the Trash. With dryRun, nothing is deleted and every
// Result has a nil Err. Forbidden paths are always refused.
func Trash(paths []string, dryRun bool) []Result {
	if dryRun {
		out := make([]Result, 0, len(paths))
		for _, p := range paths {
			out = append(out, Result{Path: p})
		}
		return out
	}
	// Prefer the `trash` CLI (fast, per-path errors) when available; otherwise
	// batch Finder deletes via osascript.
	if exe, err := exec.LookPath("trash"); err == nil {
		return trashWithCLI(exe, paths)
	}
	return trashWithOSA(paths)
}

// trashWithCLI deletes each path with the trash binary, refusing forbidden ones.
func trashWithCLI(exe string, paths []string) []Result {
	out := make([]Result, 0, len(paths))
	for _, p := range paths {
		if safety.IsForbidden(p) {
			out = append(out, Result{Path: p, Err: ErrForbidden})
			continue
		}
		if err := exec.Command(exe, p).Run(); err != nil {
			out = append(out, Result{Path: p, Err: fmt.Errorf("trash: %w", err)})
			continue
		}
		out = append(out, Result{Path: p})
	}
	return out
}

// trashWithOSA deletes paths via Finder in chunks (Finder chokes on very large
// single calls). Forbidden paths are skipped with ErrForbidden and never sent
// to Finder.
func trashWithOSA(paths []string) []Result {
	var out []Result
	for _, c := range chunk(paths, 50) {
		var allowed []string
		for _, p := range c {
			if safety.IsForbidden(p) {
				out = append(out, Result{Path: p, Err: ErrForbidden})
				continue
			}
			allowed = append(allowed, p)
		}
		if len(allowed) == 0 {
			continue
		}
		err := runFinderDelete(allowed)
		for _, p := range allowed {
			r := Result{Path: p}
			if err != nil {
				r.Err = fmt.Errorf("osascript: %w", err)
			}
			out = append(out, r)
		}
	}
	return out
}

// runFinderDelete tells Finder to delete the given POSIX paths (-> Trash).
func runFinderDelete(paths []string) error {
	items := make([]string, 0, len(paths))
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		items = append(items, fmt.Sprintf(`(POSIX file %q) as alias`, abs))
	}
	script := "tell application \"Finder\" to delete {" + strings.Join(items, ", ") + "}"
	return exec.Command("osascript", "-e", script).Run()
}

// chunk splits s into runs of at most n elements.
func chunk(s []string, n int) [][]string {
	if n <= 0 {
		return [][]string{s}
	}
	var out [][]string
	for i := 0; i < len(s); i += n {
		j := i + n
		if j > len(s) {
			j = len(s)
		}
		out = append(out, s[i:j])
	}
	return out
}
