package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func scanPackages(rootDir string) ([]PackageInfo, error) {
	var packages []PackageInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		if path != rootDir && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		pkgInfo, err := scanPackage(path)
		if err != nil {
			return nil
		}

		if pkgInfo != nil && len(pkgInfo.Structs) > 0 {
			packages = append(packages, *pkgInfo)
		}
		return nil
	})

	return packages, err
}

func scanPackage(pkgPath string) (*PackageInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go") &&
			!strings.HasSuffix(info.Name(), ".gen.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no pkgs")
	}

	var pkgInfo *PackageInfo

	for pkgName, pkg := range pkgs {
		for _, file := range pkg.Files {
			if pkgInfo == nil {
				pkgInfo = &PackageInfo{
					Name:          pkgName,
					Path:          pkgPath,
					Structs:       nil,
					Imports:       map[string]string{},
					NeededAliases: map[string]bool{},
				}
			}

			collectImports(pkgInfo, file)

			structs := findResetableStructs(file)
			if len(structs) > 0 {
				pkgInfo.Structs = append(pkgInfo.Structs, structs...)
			}
		}
	}

	if pkgInfo == nil || len(pkgInfo.Structs) == 0 {
		return nil, nil
	}

	markNeededAliases(pkgInfo)

	return pkgInfo, nil
}

func collectImports(pkg *PackageInfo, file *ast.File) {
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(path, "/")
			alias = parts[len(parts)-1]
		}

		if _, exists := pkg.Imports[alias]; !exists {
			pkg.Imports[alias] = path
		}
	}
}

func findResetableStructs(file *ast.File) []StructInfo {
	var structs []StructInfo

	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}
		if genDecl.Doc == nil {
			return true
		}

		hasResetComment := false
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, "generate:reset") {
				hasResetComment = true
				break
			}
		}
		if !hasResetComment {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			structs = append(structs, StructInfo{
				Name:   typeSpec.Name.Name,
				Fields: structType.Fields.List,
			})
		}

		return true
	})

	return structs
}

func markNeededAliases(pkg *PackageInfo) {
	for _, s := range pkg.Structs {
		for _, f := range s.Fields {
			if len(f.Names) == 0 {
				continue
			}
			inspectTypeForAliases(f.Type, pkg.NeededAliases)
		}
	}
}

func inspectTypeForAliases(expr ast.Expr, needed map[string]bool) {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			needed[x.Name] = true
		}
	case *ast.StarExpr:
		inspectTypeForAliases(t.X, needed)
	case *ast.ArrayType:
		inspectTypeForAliases(t.Elt, needed)
	case *ast.MapType:
		inspectTypeForAliases(t.Key, needed)
		inspectTypeForAliases(t.Value, needed)
	}
}
