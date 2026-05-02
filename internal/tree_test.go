package internal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildTreeFiltersAndSorts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "README.md"), "# r\n")
	mustWrite(t, filepath.Join(tmp, "zzz.md"), "z\n")
	mustWrite(t, filepath.Join(tmp, "ignored.txt"), "skip me")
	mustMkdir(t, filepath.Join(tmp, "docs"))
	mustWrite(t, filepath.Join(tmp, "docs", "guide.md"), "# g\n")
	mustMkdir(t, filepath.Join(tmp, "empty"))
	mustWrite(t, filepath.Join(tmp, "empty", "no-md-here.txt"), "skip me")
	mustMkdir(t, filepath.Join(tmp, ".hidden"))
	mustWrite(t, filepath.Join(tmp, ".hidden", "secret.md"), "no\n")

	tree, err := BuildTree(tmp)
	if err != nil {
		t.Fatalf("BuildTree: %v", err)
	}

	// Expect: docs/ (dir, contains guide.md), README.md, zzz.md
	// NOT: ignored.txt, empty/, .hidden/
	got := flatten(tree)
	want := []string{
		"docs/ [dir]",
		"docs/guide.md",
		"/README.md",
		"/zzz.md",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d: want %q, got %q", i, want[i], got[i])
		}
	}
}

func TestRenderTreeHighlightsCurrent(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "README.md"), "# r\n")
	mustMkdir(t, filepath.Join(tmp, "docs"))
	mustWrite(t, filepath.Join(tmp, "docs", "guide.md"), "# g\n")

	tree, err := BuildTree(tmp)
	if err != nil {
		t.Fatal(err)
	}

	html := RenderTree(tree, "/docs/guide.md")

	if !strings.Contains(html, `<details open><summary>docs</summary>`) {
		t.Errorf("expected docs/ to render with open details, got: %s", html)
	}
	if !strings.Contains(html, `class="file current"`) || !strings.Contains(html, `href="/docs/guide.md"`) {
		t.Errorf("expected current file to be highlighted, got: %s", html)
	}
	// README is not the current file → not marked current
	if strings.Contains(html, `class="file current"><a href="/README.md"`) {
		t.Errorf("README should not be marked current, got: %s", html)
	}
}

func TestServerRendersFileExplorer(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "README.md"), "# r\n")
	mustMkdir(t, filepath.Join(tmp, "docs"))
	mustWrite(t, filepath.Join(tmp, "docs", "guide.md"), "# g\n")

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmp))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs/guide.md", nil))
	body := rec.Body.String()

	for _, want := range []string{
		`id="file-explorer"`,
		`href="/README.md"`,
		`<summary>docs</summary>`,
		`href="/docs/guide.md"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}

// helpers

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func flatten(n *Node) []string {
	var out []string
	var walk func(n *Node, depth int)
	walk = func(n *Node, depth int) {
		for _, c := range n.Children {
			if c.IsDir {
				out = append(out, c.Name+"/ [dir]")
				for _, gc := range c.Children {
					if gc.IsDir {
						walk(c, depth+1)
					} else {
						out = append(out, c.Name+"/"+gc.Name)
					}
				}
			} else {
				out = append(out, c.URLPath)
			}
		}
	}
	walk(n, 0)
	return out
}
