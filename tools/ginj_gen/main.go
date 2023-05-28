package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s <dir>\n", os.Args[0])
		os.Exit(0)
	}

	fset := token.NewFileSet()
	pkgPath, _ := filepath.Abs(os.Args[1])
	filepath.Walk(pkgPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			pkgs, err := parser.ParseDir(fset, path, func(info fs.FileInfo) bool {
				return !strings.Contains(info.Name(), "_test.go") && !strings.Contains(info.Name(), "_generate.go")
			}, parser.ParseComments)

			if err != nil {
				fmt.Printf("Cannot parse dir %s: %v\n", path, err)
				return err
			}

			for pkgName, pkgAst := range pkgs {
				parsePackage(path, pkgName, pkgAst)
			}
		}
		return nil
	})
}

type handler struct {
	Annotation string
	Function   string
}

type visitor struct {
	List []handler
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		anno := strings.TrimSpace(n.Doc.Text())
		if strings.HasPrefix(anno, "@Ginj:") {
			anno = strings.TrimPrefix(anno, "@Ginj:")
			anno = strings.TrimSpace(anno)
			v.List = append(v.List, handler{
				Annotation: anno,
				Function:   n.Name.String(),
			})
			fmt.Printf("Find annotated route `%v` \"%v\"\n", n.Name.String(), anno)
		}
	}
	return v
}

func parsePackage(path, pkgName string, pkgAst ast.Node) {
	fmt.Printf("Parse package %s at %s\n", pkgName, path)
	v := &visitor{}
	ast.Walk(v, pkgAst)
	if len(v.List) > 0 {
		generateCode(path, pkgName, v)
	}
}

func generateCode(pkgPath, pkgName string, vis *visitor) {
	generateFilePath := path.Join(pkgPath, fmt.Sprintf("%s_ginj_generate.go", pkgName))
	file, err := os.OpenFile(generateFilePath, os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Cannot open file %s: %v\n", generateFilePath, err)
		return
	}
	write := bufio.NewWriter(file)
	write.WriteString(fmt.Sprintf("package %s\n", pkgName))
	write.WriteString("\nimport ginj \"github.com/w-mj/ginj/lib\"\n")
	write.WriteString("\nfunc init() {\n")
	for _, f := range vis.List {
		write.WriteString(fmt.Sprintf("\t ginj.AddAnnotatedRoute(\"%s\", %s)\n", f.Annotation, f.Function))
	}
	write.WriteString("}\n")
	write.Flush()
	err = file.Close()
	if err != nil {
		log.Printf("Cannot close file %s: %v\n", generateFilePath, err)
		return
	}
}
