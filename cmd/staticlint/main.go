// Package staticlint implements a custom multichecker for static analysis.
//
// Usage:
//
//	go run ./cmd/staticlint <packages>
//
// This multichecker includes:
//   - Standard analyzers from golang.org/x/tools/go/analysis/passes
//   - All "SA" analyzers from staticcheck.io
//   - At least one analyzer from other staticcheck classes
//   - Two public analyzers (unused, nilerr)
//   - Custom analyzer prohibiting direct os.Exit call in main function of main package
//
// Custom analyzer:
//
//	Prohibits direct usage of os.Exit in main() of main package. Use error handling and return instead.
//
// Example:
//
//	go run ./cmd/staticlint ./...
package main

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/gostaticanalysis/nilerr"
	"github.com/gostaticanalysis/unused"
)

// osExitInMainAnalyzer prohibits direct os.Exit call in main() of main package.
var osExitInMainAnalyzer = &analysis.Analyzer{
	Name: "customNoOsExitInMain",
	Doc:  "prohibits direct os.Exit call in main() of main package",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		workDir, err := os.Getwd()
		if err != nil {
			return nil, nil
		}
		workDir = filepath.Clean(workDir)

		for _, file := range pass.Files {
			if !isProjectFile(pass, file, workDir) || isGeneratedFile(pass, file) {
				continue
			}

			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name.Name != "main" || fn.Recv != nil {
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
							pass.Reportf(call.Lparen, "direct os.Exit call in main() is prohibited")
						}
					}
					return true
				})
			}
		}
		return nil, nil
	},
}

func isProjectFile(pass *analysis.Pass, file *ast.File, workDir string) bool {
	if !file.Pos().IsValid() {
		return false
	}

	fname := pass.Fset.File(file.Pos()).Name()
	fname = filepath.Clean(fname)

	rel, err := filepath.Rel(workDir, fname)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}

	if strings.Contains(fname, string(filepath.Separator)+"vendor"+string(filepath.Separator)) ||
		strings.Contains(fname, string(filepath.Separator)+"node_modules"+string(filepath.Separator)) ||
		strings.Contains(fname, string(filepath.Separator)+".git"+string(filepath.Separator)) {
		return false
	}

	return true
}

func isGeneratedFile(pass *analysis.Pass, file *ast.File) bool {
	if file.Pos().IsValid() {
		fname := pass.Fset.File(file.Pos()).Name()
		if strings.HasSuffix(fname, ".pb.go") {
			return true
		}
	}

	if file.Doc != nil {
		for _, comment := range file.Doc.List {
			if strings.Contains(comment.Text, "Code generated") ||
				strings.Contains(comment.Text, "DO NOT EDIT") {
				return true
			}
		}
	}

	return false
}

func wrapAnalyzer(analyzer *analysis.Analyzer) *analysis.Analyzer {
	if analyzer.Run == nil {
		return analyzer
	}

	originalRun := analyzer.Run
	wrappedAnalyzer := *analyzer
	wrappedAnalyzer.Run = func(pass *analysis.Pass) (interface{}, error) {
		workDir, err := os.Getwd()
		if err != nil {
			workDir = ""
		}
		workDir = filepath.Clean(workDir)

		filteredFiles := make([]*ast.File, 0, len(pass.Files))
		for _, file := range pass.Files {
			if isProjectFile(pass, file, workDir) && !isGeneratedFile(pass, file) {
				filteredFiles = append(filteredFiles, file)
			}
		}

		if len(filteredFiles) == 0 {
			return nil, nil
		}

		filteredPass := &analysis.Pass{
			Analyzer:          pass.Analyzer,
			Fset:              pass.Fset,
			Files:             filteredFiles,
			OtherFiles:        pass.OtherFiles,
			IgnoredFiles:      pass.IgnoredFiles,
			Pkg:               pass.Pkg,
			TypesInfo:         pass.TypesInfo,
			TypesSizes:        pass.TypesSizes,
			ResultOf:          pass.ResultOf,
			Report:            pass.Report,
			ImportObjectFact:  pass.ImportObjectFact,
			ExportObjectFact:  pass.ExportObjectFact,
			ImportPackageFact: pass.ImportPackageFact,
			ExportPackageFact: pass.ExportPackageFact,
		}

		return originalRun(filteredPass)
	}

	return &wrappedAnalyzer
}

func main() {
	var analyzers []*analysis.Analyzer

	// Standard analyzers
	standardAnalyzers := []*analysis.Analyzer{

		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		nilfunc.Analyzer,
		shadow.Analyzer,
		sigchanyzer.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
	}

	for _, analyzer := range standardAnalyzers {
		analyzers = append(analyzers, wrapAnalyzer(analyzer))
	}

	// SA class
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, wrapAnalyzer(v.Analyzer))
	}

	// S class
	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, wrapAnalyzer(v.Analyzer))
		break // one enough
	}

	// ST class (Style)
	for _, v := range stylecheck.Analyzers {
		analyzers = append(analyzers, wrapAnalyzer(v.Analyzer))
		break // one enough
	}

	// Public analyzers
	analyzers = append(analyzers, wrapAnalyzer(unused.Analyzer))
	analyzers = append(analyzers, wrapAnalyzer(nilerr.Analyzer))

	// Custom analyzer
	analyzers = append(analyzers, osExitInMainAnalyzer)

	multichecker.Main(analyzers...)
}
