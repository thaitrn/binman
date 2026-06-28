package trash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChunk(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e"}
	got := chunk(in, 2)
	want := [][]string{{"a", "b"}, {"c", "d"}, {"e"}}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			t.Errorf("chunk %d: %v, want %v", i, got[i], want[i])
		}
	}
}

func TestTrash_DryRunKeepsFiles(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "keepme.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := Trash([]string{f}, true)
	if len(res) != 1 || res[0].Path != f || res[0].Err != nil {
		t.Fatalf("dry-run result unexpected: %+v", res)
	}
	if _, err := os.Stat(f); err != nil {
		t.Errorf("file removed during dry-run: %v", err)
	}
}

func TestTrashWithOSA_ForbiddenSkipped(t *testing.T) {
	// Forbidden paths must be refused and never sent to Finder. Call the OSA
	// path directly so the test does not depend on whether `trash` is present.
	res := trashWithOSA([]string{"/System/Library/foo", "/usr/local/x"})
	for _, r := range res {
		if r.Err != ErrForbidden {
			t.Errorf("%s: err = %v, want ErrForbidden", r.Path, r.Err)
		}
	}
}
