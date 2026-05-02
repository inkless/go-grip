# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`go-grip` is a Go CLI that serves Markdown files locally as GitHub-styled HTML. It is a pure-Go reimplementation of [grip](https://github.com/joeyespo/grip) that does not call GitHub's API — all rendering happens in-process via [goldmark](https://github.com/yuin/goldmark) extensions.

Module path: `github.com/chrishrb/go-grip`. Go 1.25.

## Commands

Tasks live in the `Makefile`. CI uses the same targets via `actions/setup-go` and `golangci-lint-action`. Required local tools: `go` (see `go.mod` for version) and `golangci-lint`.

- `make build` — build to `bin/grip`
- `make install` — build and install to `$GOPATH/bin/grip` so the global `grip` reflects this checkout. Run this after editing source.
- `make run ARGS="<args>"` — `go run main.go <args>`
- `make test` — `go test ./...`
- `make lint` — `golangci-lint run`
- `make format` — `go fmt ./...`
- `make compile` — cross-compile for darwin/linux/windows × amd64/arm64 into `bin/`
- `make all` — format + lint + test + build (the default target)
- `make clean` — remove `bin/`

Run a single test: `go test ./pkg/alert -run TestAlert` (package path + `-run <regex>`).

Optional pre-commit hooks (`.pre-commit-config.yaml`) run `golangci-lint run --fix` and `gofmt -w` on commit if you install them via `pre-commit install`. CI enforces `gofmt -d` produces no diff regardless.

## Architecture

The entry point is trivial: `main.go` → `cmd.Execute()` (cobra). The `cmd/root.go` `RunE` wires together a `Parser` and `Server` from `internal/`, then calls `server.Serve(file)`.

### Request flow

1. **`internal/server.go`** starts an `http.Server`. The mux has two routes:
   - `/static/*` is served from the embedded FS in `defaults/` (CSS, JS, emojis, images).
   - `/*` either renders `.md` files through the parser into `defaults/templates/layout.html`, or falls through to a regular `http.FileServer` for everything else (so directory listings, images, etc. work).
2. When `--no-reload` is not set, the handler is wrapped in [`aarol/reload`](https://github.com/aarol/reload), which watches the served directory and pushes a websocket refresh to the browser on file change. CORS is opened up for that websocket.
3. Markdown files bypass HTTP caching (`setNoCacheHeaders` + `stripCacheValidators` for directories) so reloads always show fresh content.

### Markdown pipeline

`internal/parser.go` is the single place where the goldmark instance is configured. Each feature is a goldmark extension, most of which live under `pkg/` and follow the same shape:

- `pkg/<feature>/<feature>.go` — exposes `New()`/`Extender` and `Extend(goldmark.Markdown)`.
- `transformer.go` — AST transformer that recognizes the syntax and rewrites nodes.
- `renderer.go` — `NodeRenderer` that emits HTML.
- `ast.go` — custom AST node types (when needed).

Extensions wired in:

- `pkg/alert` — GitHub-style `> [!NOTE]` / `[!TIP]` / `[!IMPORTANT]` / `[!WARNING]` / `[!CAUTION]` blockquotes.
- `pkg/details` — collapsible `<details>` blocks.
- `pkg/footnote` — footnote support.
- `pkg/ghissue` — `#123` and `owner/repo#123` references; auto-detects the repo via `git.go` (`git remote`) when no `WithRepository` option is passed. Inline parser priority is intentionally **higher** than the goldmark `hashtag` extension so `#123` is consumed first.
- `pkg/highlighting` — code block syntax highlighting via [chroma](https://github.com/alecthomas/chroma); the server pre-renders both `github` and `github-dark` chroma stylesheets and ships both in the page (theme switch happens in JS).
- `pkg/mathjax` — `$inline$`, `$$block$$`, and ```` ```math ```` blocks (the code-block transformer rewrites fenced math into MathJax nodes).
- `pkg/tasklist` — `- [ ]` / `- [x]` checkboxes.

Plus stock extensions: `Linkify`, `Table`, `Strikethrough`, `goldmark-emoji`, `go.abhg.dev/goldmark/hashtag`, and `go.abhg.dev/goldmark/mermaid` in client-render mode (`NoScript: true` — the page already loads mermaid.js from `static/`).

The renderer runs with `html.WithUnsafe()` so raw HTML in markdown passes through.

### Embedded assets

`defaults/embed.go` uses `//go:embed templates` and `//go:embed static` to bundle `defaults/templates/layout.html` and everything under `defaults/static/` (css/, js/, emojis/, images/) into the binary. The layout template receives `Content`, `BoundingBox`, `CssCodeLight`, and `CssCodeDark`.

When changing the rendering pipeline, remember: extension order/priority in `parser.go` matters (e.g. `ghissue` must precede `hashtag`), and the chroma class names emitted by `highlighting` must match the embedded CSS.
