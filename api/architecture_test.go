package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/swuecho/chat_backend"

// allowedImports lists, for every package in this module, the other module
// packages its non-test code may import. Adding an edge here is a deliberate
// architecture decision; see "Backend Boundaries" in AGENTS.md.
var allowedImports = map[string][]string{
	".":            {"apicontract", "apiopenapi", "config", "dto", "handler", "middleware", "sqlc_queries", "static", "svc"},
	"apicontract":  {"httpx"},
	"apiopenapi":   {"apicontract"},
	"auth":         {},
	"cmd/openapi":  {"apiopenapi", "handler"},
	"config":       {},
	"domain":       {},
	"dto":          {"domain", "validation"},
	"handler":      {"apicontract", "auth", "domain", "dto", "httpx", "middleware", "pkg/util", "provider", "requestctx", "svc", "validation"},
	"httpx":        {"domain", "dto", "pkg/util", "requestctx", "validation"},
	"llm/claude":   {"domain"},
	"llm/gemini":   {"domain"},
	"middleware":   {"auth", "dto", "requestctx", "sqlc_queries", "validation"},
	"pkg/util":     {},
	"provider":     {"domain", "llm/claude", "llm/gemini", "pkg/util"},
	"requestctx":   {},
	"sqlc_queries": {},
	"static":       {},
	"svc":          {"auth", "domain", "pkg/util", "provider", "sqlc_queries"},
	"testutil":     {},
	"validation":   {},
}

func TestModuleImportGraph(t *testing.T) {
	actual := map[string]map[string]bool{}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name := entry.Name(); path != "." && (strings.HasPrefix(name, ".") || name == "tmp" || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		pkg := filepath.ToSlash(filepath.Dir(path))
		if actual[pkg] == nil {
			actual[pkg] = map[string]bool{}
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if dep, ok := strings.CutPrefix(importPath, modulePath+"/"); ok {
				actual[pkg][dep] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	pkgs := make([]string, 0, len(actual))
	for pkg := range actual {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	for _, pkg := range pkgs {
		allowed, known := allowedImports[pkg]
		if !known {
			t.Errorf("package %q is not listed in allowedImports; add it with its intended dependencies", pkg)
			continue
		}
		allowedSet := map[string]bool{}
		for _, dep := range allowed {
			allowedSet[dep] = true
		}
		deps := make([]string, 0, len(actual[pkg]))
		for dep := range actual[pkg] {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			if !allowedSet[dep] {
				t.Errorf("package %q imports %q, which allowedImports does not permit", pkg, dep)
			}
		}
	}
}
