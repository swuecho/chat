package pathlib

import (
	"io"
	"os"
	"testing"
)

func TestParts(t *testing.T) {
	path := NewPath("/usr/local/bin/python")
	parts := path.Parts()
	if len(parts) != 5 || parts[0] != "/" || parts[1] != "usr" || parts[2] != "local" || parts[3] != "bin" || parts[4] != "python" {
		t.Error("Parts incorrect")
	}
}

func TestJoin(t *testing.T) {
	path := NewPath("/usr/local/bin")
	newPath := path.Join("python")
	if newPath.PathString != "/usr/local/bin/python" {
		t.Error("Join incorrect")
	}
}

func TestParent(t *testing.T) {
	path := NewPath("/usr/local/bin/python")
	parent := path.Parent()
	if parent.PathString != "/usr/local/bin" {
		t.Error("Parent incorrect")
	}
}

func TestName(t *testing.T) {
	path := NewPath("/usr/local/bin/python")
	name := path.Name()
	if name != "python" {
		t.Error("Name incorrect")
	}
}

func TestRelativeTo(t *testing.T) {
	path := NewPath("/usr/local/bin/python")
	otherPath := NewPath("/usr/local")
	relativePath := path.RelativeTo(otherPath)
	if relativePath != "bin/python" {
		t.Error("RelativeTo incorrect")
	}
}

func TestExists(t *testing.T) {
	// create temporary file for testing
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal("Cannot create temporary file", err)
	}
	defer os.Remove(tempFile.Name())

	// test if file exists
	existingPath := NewPath(tempFile.Name())
	if !existingPath.Exists() {
		t.Error("Exists incorrect")
	}

	// test if non-existent file exists
	nonExistentPath := NewPath("nonexistentfile.txt")
	if nonExistentPath.Exists() {
		t.Error("Exists incorrect")
	}
}

func TestReadWriteText(t *testing.T) {
	// create temporary file for testing
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal("Cannot create temporary file", err)
	}
	defer os.Remove(tempFile.Name())

	// write text to file
	filePath := NewPath(tempFile.Name())
	err = filePath.WriteText("hello, world!")
	if err != nil {
		t.Error("WriteText failed:", err)
	}

	// read text from file
	f, err := os.Open(filePath.PathString)
	if err != nil {
		t.Fatal("Cannot open file", err)
	}
	defer f.Close()
	contentBytes, err := io.ReadAll(f)
	if err != nil {
		t.Fatal("Cannot read file", err)
	}
	content := string(contentBytes)
	if content != "hello, world!" {
		t.Error("ReadText incorrect")
	}
}
