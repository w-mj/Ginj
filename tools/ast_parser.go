package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type visitor struct {
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		log.Printf("func name: %s", n.Name.String())
	}
	return v
}
func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s <dir>\n", os.Args[0])
		os.Exit(0)
	}

	fset := token.NewFileSet()
	pkgPath, _ := filepath.Abs(os.Args[1])
	pkgs, err := parser.ParseDir(fset, pkgPath, func(info fs.FileInfo) bool {
		fmt.Printf("Find go file %s\n", info.Name())
		return !strings.Contains(info.Name(), "_test.go")
	}, parser.ParseComments)

	if err != nil {
		fmt.Printf("Cannot parse dir %s: %s\n", os.Args[1], err.Error())
		os.Exit(1)
	}

	for pkgName, pkgAst := range pkgs {
		fmt.Printf("Parse package %s\n", pkgName)
		v := &visitor{}
		ast.Walk(v, pkgAst)
	}
}
