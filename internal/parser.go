package internal

import (
	"bytes"
	"strings"

	"github.com/inkless/go-grip/pkg/alert"
	"github.com/inkless/go-grip/pkg/details"
	"github.com/inkless/go-grip/pkg/footnote"
	"github.com/inkless/go-grip/pkg/ghissue"
	"github.com/inkless/go-grip/pkg/highlighting"
	"github.com/inkless/go-grip/pkg/mathjax"
	"github.com/inkless/go-grip/pkg/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/mermaid"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (m Parser) MdToHTML(input []byte) ([]byte, string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Table,
			extension.Strikethrough,
			footnote.Footnote,
			tasklist.TaskList,
			emoji.Emoji,
			&hashtag.Extender{},
			alert.New(),
			highlighting.Highlighting,
			&mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true},
			mathjax.MathJax,
			ghissue.New(),
			details.New(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	doc := md.Parser().Parse(text.NewReader(input))
	title := firstH1(doc, input)

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, input, doc); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), title, nil
}

func firstH1(doc ast.Node, source []byte) string {
	var title string
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
			title = strings.TrimSpace(string(h.Text(source)))
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return title
}
