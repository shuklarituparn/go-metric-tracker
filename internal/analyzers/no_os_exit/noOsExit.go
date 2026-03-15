// Package osexit запрещает прямой вызов os.Exit внутри func main() пакета main.
package osexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `osexit запрещает прямой вызов os.Exit внутри func main() пакета main.

Зачем:
- main() становится проще тестировать (вынесите логику в run() и возвращайте error / код).
- избегаем "жёсткого" завершения процесса в середине main(), что усложняет сопровождение.

Правило:
- запрещено: os.Exit(...) внутри func main() (пакет main).
- разрешено:
  - os.Exit в других функциях (если вам очень надо)
  - log.Fatal(...) (он вызывает os.Exit внутри пакета log, а не напрямую)
`

var Analyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	var mainDecl *ast.FuncDecl
	ins.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fd := n.(*ast.FuncDecl)
		if fd.Recv == nil && fd.Name != nil && fd.Name.Name == "main" && fd.Type != nil && fd.Type.Results == nil {
			mainDecl = fd
		}
	})
	if mainDecl == nil || mainDecl.Body == nil {
		return nil, nil
	}

	ast.Inspect(mainDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if pass.TypesInfo == nil {
			return true
		}
		obj := calledFuncObject(pass.TypesInfo, call.Fun)
		if obj == nil {
			return true
		}

		fn, ok := obj.(*types.Func)
		if !ok || fn.Pkg() == nil {
			return true
		}
		if fn.Pkg().Path() == "os" && fn.Name() == "Exit" {
			pass.Reportf(call.Lparen, "запрещён прямой вызов os.Exit в func main(); вынесите логику в run() и используйте log.Fatal/возврат ошибки")
		}
		return true
	})

	return nil, nil
}

func calledFuncObject(info *types.Info, fun ast.Expr) types.Object {
	switch f := fun.(type) {
	case *ast.Ident:
		return info.Uses[f]
	case *ast.SelectorExpr:
		return info.Uses[f.Sel]
	default:
		return nil
	}
}
