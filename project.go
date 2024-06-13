package mcwig

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProjectManager struct {
}

func NewProjectManager() ProjectManager {
	return ProjectManager{}
}

// Find project root by file path. Project root must contain .git directory in it.
// otherwise "working directory" will be returned.
func (p ProjectManager) FindRoot(buf *Buffer) (root string, err error) {
	fp := filepath.Dir(buf.FilePath)

	root, _ = os.Getwd()

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = fp
	r, err := cmd.Output()
	if err != nil {
		return
	}

	return strings.TrimSpace(string(r)), nil
}
