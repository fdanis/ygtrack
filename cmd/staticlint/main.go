package main

import (
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
