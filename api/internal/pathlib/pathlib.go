package pathlib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Path struct {
	PathString string // the path string
}

func NewPath(pathString string) *Path {
	return &Path{
		PathString: pathString,
	}
}

func (p Path) Parts() []string {

	// Split path using slashes
	parts := strings.Split(p.PathString, "/")

	// Remove the empty substring at the start if path starts with a slash
	if parts[0] == "" {
		parts[0] = "/"
	}

	return parts
}

func (p *Path) Join(otherPath string) *Path {
	joined := filepath.Join(p.PathString, otherPath)
	return NewPath(joined)
}

func (p *Path) Parent() *Path {
	parent := filepath.Dir(p.PathString)
	return NewPath(parent)
}

func (p *Path) Name() string {
	return filepath.Base(p.PathString)
}

func (p *Path) RelativeTo(otherPath *Path) string {
	relativePath, _ := filepath.Rel(otherPath.PathString, p.PathString)
	return relativePath
}

func (p *Path) String() string {
	return p.PathString
}

func (p *Path) Exists() bool {
	_, err := os.Stat(p.PathString)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (p *Path) ReadText() (string, error) {
	if !p.Exists() {
		return "", fmt.Errorf("file %s does not exist", p.PathString)
	}
	content, err := os.ReadFile(p.PathString)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func (p *Path) WriteText(content string) error {
	err := os.WriteFile(p.PathString, []byte(content), 0644)
	return err
}
