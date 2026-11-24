package util

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// pkg/util is framework-independent: it must not depend on any other package
// in this module, so HTTP, auth, and transport concerns cannot leak into it.
func TestUtilDoesNotImportModulePackages(t *testing.T) {
	const modulePrefix = "github.com/swuecho/chat_backend/"
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
			if strings.HasPrefix(importPath, modulePrefix) {
				position := files.Position(imported.Pos())
				t.Errorf("%s:%d imports %s; pkg/util must stay free of module dependencies", path, position.Line, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
