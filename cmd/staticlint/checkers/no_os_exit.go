// Package checkers содержит кастомные анализаторы для статического анализа Go-кода.
package checkers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer анализатор, запрещающий использование os.Exit() в main
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "Запрещает использование os.Exit() в функции main пакета main",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверяем, что анализируем файл пакета main
		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				ast.Inspect(fn, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok {
								if ident.Name == "os" && sel.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "Использование os.Exit() в main запрещено")
								}
							}
						}
					}
					return true
				})
			}
			return true
		})
	}

	return nil, nil
}
