package internal

import (
	"html"
	"strings"
)

func RenderTree(root *Node, currentPath string) string {
	if root == nil || len(root.Children) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<ul class="file-tree-list">`)
	for _, c := range root.Children {
		renderNode(&b, c, currentPath)
	}
	b.WriteString(`</ul>`)
	return b.String()
}

func renderNode(b *strings.Builder, n *Node, currentPath string) {
	if n.IsDir {
		open := ""
		if containsCurrent(n, currentPath) {
			open = " open"
		}
		b.WriteString(`<li class="dir"><details`)
		b.WriteString(open)
		b.WriteString(`><summary>`)
		b.WriteString(html.EscapeString(n.Name))
		b.WriteString(`</summary><ul>`)
		for _, c := range n.Children {
			renderNode(b, c, currentPath)
		}
		b.WriteString(`</ul></details></li>`)
		return
	}

	cls := "file"
	if n.URLPath == currentPath {
		cls += " current"
	}
	b.WriteString(`<li class="` + cls + `"><a href="`)
	b.WriteString(html.EscapeString(n.URLPath))
	b.WriteString(`">`)
	b.WriteString(html.EscapeString(n.Name))
	b.WriteString(`</a></li>`)
}

func containsCurrent(n *Node, currentPath string) bool {
	for _, c := range n.Children {
		if !c.IsDir && c.URLPath == currentPath {
			return true
		}
		if c.IsDir && containsCurrent(c, currentPath) {
			return true
		}
	}
	return false
}
