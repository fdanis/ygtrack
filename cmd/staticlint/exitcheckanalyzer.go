package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check os.exit in main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		for _, d := range file.Decls {
			if f, ok := d.(*ast.FuncDecl); ok {
				if f.Name.Name != "main" {
					continue
				}
				for _, stmt := range f.Body.List {
					if call, ok := stmt.(*ast.ExprStmt); ok {
						if ce, ok := call.X.(*ast.CallExpr); ok {
							if ff, ok := ce.Fun.(*ast.SelectorExpr); ok {
								if p, ok := ff.X.(*ast.Ident); ok {
									if ff.Sel.Name == "Exit" && p.Name == "os" {
										pass.Reportf(ff.Sel.NamePos, "call os.Exit in main function")
										return nil, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil, nil
}
