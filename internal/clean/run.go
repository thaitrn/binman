package clean

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thaitrn/binman/internal/human"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
)

// Report is the computed plan for one target.
type Report struct {
	Target    Target
	Paths     []string // trash candidates
	Size      int64    // sum for trash targets
	Command   string   // "name args" for command targets
	Available bool     // command target: binary on PATH
}

// Plan computes reports for the targets whose Group is in groups.
func Plan(home string, groups []string) []Report {
	want := make(map[string]bool, len(groups))
	for _, g := range groups {
		want[g] = true
	}
	var out []Report
	for _, t := range Targets() {
		if !want[t.Group] {
			continue
		}
		r := Report{Target: t}
		switch t.Kind {
		case KindTrash:
			r.Paths = t.Paths(home)
			for _, p := range r.Paths {
				if safety.IsForbidden(p) {
					continue
				}
				s, _ := scan.DirSize(p)
				r.Size += s
			}
		case KindCommand:
			name, args := t.Cmd(home)
			r.Command = strings.Join(append([]string{name}, args...), " ")
			r.Available = available(name)
		}
		out = append(out, r)
	}
	return out
}

// Apply acts on reports. With dryRun it prints what would happen; otherwise it
// trashes paths and runs commands. Output goes to stdout.
func Apply(reports []Report, dryRun bool) {
	for _, r := range reports {
		switch r.Target.Kind {
		case KindTrash:
			applyTrash(r, dryRun)
		case KindCommand:
			applyCommand(r, dryRun)
		}
	}
}

func applyTrash(r Report, dryRun bool) {
	if len(r.Paths) == 0 {
		fmt.Printf("  • %-28s nothing\n", r.Target.Name)
		return
	}
	if dryRun {
		fmt.Printf("  • %-28s %d item(s), ~%s\n", r.Target.Name, len(r.Paths), human.Bytes(r.Size))
		return
	}
	allowed := make([]string, 0, len(r.Paths))
	for _, p := range r.Paths {
		if safety.IsForbidden(p) {
			continue
		}
		allowed = append(allowed, p)
	}
	res := trash.Trash(allowed, false)
	failed := 0
	for _, x := range res {
		if x.Err != nil {
			failed++
		}
	}
	fmt.Printf("  ✓ %-28s moved %d item(s) (~%s) to Trash\n", r.Target.Name, len(allowed)-failed, human.Bytes(r.Size))
}

func applyCommand(r Report, dryRun bool) {
	name := strings.Fields(r.Command)[0]
	if !r.Available {
		fmt.Printf("  • %-28s skipped (%s not found)\n", r.Target.Name, name)
		return
	}
	if dryRun {
		fmt.Printf("  • %-28s would run `%s`\n", r.Target.Name, r.Command)
		return
	}
	c, a := r.Target.Cmd("")
	cmd := exec.Command(c, a...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("  ✗ %-28s %v\n", r.Target.Name, err)
		return
	}
	fmt.Printf("  ✓ %-28s ran\n", r.Target.Name)
}
