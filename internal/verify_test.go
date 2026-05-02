package internal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestTitleVerification is a manual verification harness. Run with:
//
//	go test ./internal -run TestTitleVerification -v
func TestTitleVerification(t *testing.T) {
	cases := []struct {
		label, file, body string
	}{
		{"first H1 wins", "doc.md", "intro paragraph\n\n# My Document\n\n## Sub\n"},
		{"no H1 -> filename fallback", "notes.md", "no heading at all\n"},
		{"setext H1", "setext.md", "Setext Title\n============\n\ncontent\n"},
		{"inline formatting in H1", "fancy.md", "# Hello **world** `code`\n"},
		{"H1 inside fenced code is ignored", "codeonly.md", "```\n# not a real heading\n```\n"},
	}

	tmp := t.TempDir()
	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmp))
	titleRe := regexp.MustCompile(`<title>([^<]*)</title>`)

	for _, tc := range cases {
		path := filepath.Join(tmp, tc.file)
		if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
			t.Fatalf("write %s: %v", tc.file, err)
		}
		req := httptest.NewRequest(http.MethodGet, "/"+tc.file, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		match := titleRe.FindStringSubmatch(rec.Body.String())
		if len(match) < 2 {
			t.Errorf("[%s] no <title> tag in response", tc.label)
			continue
		}
		t.Logf("[%-35s] %s -> <title>%s</title>", tc.label, tc.file, match[1])
	}
}
