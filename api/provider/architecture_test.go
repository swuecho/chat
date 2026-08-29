package provider

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProviderDoesNotImportGeneratedQueries(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if importPath == "github.com/swuecho/chat_backend/sqlc_queries" {
				position := files.Position(imported.Pos())
				t.Errorf("%s:%d imports generated SQLC queries; map persistence records before the provider boundary", path, position.Line)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
