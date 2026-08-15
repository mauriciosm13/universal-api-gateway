package covercheck

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// Package describes an internal Go package for coverage accounting.
type Package struct {
	ImportPath string
	Dir        string
	GoFiles    []string
	HasTests   bool
}

func fileHasFuncDecl(src []byte) bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.SkipObjectResolution)
	if err != nil {
		return false
	}
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.FuncDecl); ok {
			return true
		}
	}
	return false
}

func packageHasFuncs(pkg Package) bool {
	for _, name := range pkg.GoFiles {
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(pkg.Dir, name)
		}
		src, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if fileHasFuncDecl(src) {
			return true
		}
	}
	return false
}
