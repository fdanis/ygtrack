package main

import (
	"errors"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
)

func main() {

	analyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shift.Analyzer,
		asmdecl.Analyzer,
		atomicalign.Analyzer,
		timeformat.Analyzer,
		sortslice.Analyzer,
		bools.Analyzer,
		///.....
		usesgenerics.Analyzer,
		ExitCheckAnalyzer,
	}

	for k, v := range staticcheck.Analyzers {
		if strings.HasPrefix(k, "SA1") {
			analyzers = append(analyzers, v)
		}
	}
	for k, v := range staticcheck.Analyzers {
		if strings.HasPrefix(k, "ST") {
			analyzers = append(analyzers, v)
			break
		}
	}

	multichecker.Main(
		analyzers...,
	)
}

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
										return nil, errors.New("call os.Exit in main function")
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
