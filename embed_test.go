package rt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pkg.safego.dev/rt"
)

// The embedded tree is only worth anything if it is the tree on disk. A source file added to
// an intrinsic package without a matching embed pattern would ship a binary whose idea of its
// own contract is a file short, so this walks the directory and demands the two agree.
func TestEmbeddedSourceMatchesTheTreeOnDisk(t *testing.T) {
	files, err := rt.Collect()
	if err != nil {
		t.Fatalf("collecting: %v", err)
	}

	embedded := map[string]string{}
	for _, f := range files {
		embedded[f.Name] = string(f.Content)
	}

	for _, dir := range []string{"errors", "sync", "time", "volatile"} {
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatalf("reading %s: %v", dir, readErr)
		}

		for _, entry := range entries {
			name := filepath.ToSlash(filepath.Join(dir, entry.Name()))
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			want, fileErr := os.ReadFile(name)
			if fileErr != nil {
				t.Fatalf("reading %s: %v", name, fileErr)
			}

			got, isEmbedded := embedded[name]
			if !isEmbedded {
				t.Errorf("%s is on disk but not embedded", name)

				continue
			}

			if got != string(want) {
				t.Errorf("%s differs between the embedded copy and the file on disk", name)
			}

			delete(embedded, name)
		}
	}

	delete(embedded, "go.mod")

	for name := range embedded {
		t.Errorf("%s is embedded but not on disk", name)
	}
}

// A test file is embedded (the published module's hash covers it) but must not be handed out
// as intrinsic source.
func TestCollectSkipsTestFiles(t *testing.T) {
	files, err := rt.Collect()
	if err != nil {
		t.Fatalf("collecting: %v", err)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name, "_test.go") {
			t.Errorf("Collect returned the test file %s", f.Name)
		}
	}
}
