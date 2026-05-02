package internal

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Node struct {
	Name     string
	URLPath  string
	IsDir    bool
	Children []*Node
}

func BuildTree(root string) (*Node, error) {
	return buildNode(root, "")
}

func buildNode(absPath, relURLPath string) (*Node, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Name:    filepath.Base(absPath),
		URLPath: "/" + strings.TrimPrefix(relURLPath, "/"),
		IsDir:   info.IsDir(),
	}
	if relURLPath == "" {
		node.URLPath = "/"
	}

	if !node.IsDir {
		return node, nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var dirs, files []*Node
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		childAbs := filepath.Join(absPath, name)
		childRel := path.Join(relURLPath, name)

		if e.IsDir() {
			child, err := buildNode(childAbs, childRel)
			if err != nil {
				continue
			}
			if hasMarkdown(child) {
				dirs = append(dirs, child)
			}
			continue
		}

		if isMarkdown(name) {
			files = append(files, &Node{
				Name:    name,
				URLPath: "/" + strings.TrimPrefix(childRel, "/"),
				IsDir:   false,
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name) })
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name) })
	node.Children = append(dirs, files...)
	return node, nil
}

func hasMarkdown(n *Node) bool {
	if !n.IsDir {
		return isMarkdown(n.Name)
	}
	for _, c := range n.Children {
		if hasMarkdown(c) {
			return true
		}
	}
	return false
}

func isMarkdown(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".md" || ext == ".markdown"
}
