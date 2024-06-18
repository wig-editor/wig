package mcwig

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProjectManager struct {
	root string
}

func NewProjectManager() ProjectManager {
	root, _ := os.Getwd()

	return ProjectManager{
		root: root,
	}
}

func (p ProjectManager) GetRoot() (root string) {
	return p.root
}

// Find project root by file path. Project root must contain .git directory in it.
// otherwise "working directory" will be returned.
func (p ProjectManager) FindRoot(buf *Buffer) (root string, err error) {
	root = p.root
	fp := filepath.Dir(buf.FilePath)

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = fp
	r, err := cmd.Output()
	if err != nil {
		return root, nil
	}

	return strings.TrimSpace(string(r)), nil
}
