package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/firstrow/wig"
)

// WorkspaceCacheEntry stores the file paths and active file for a single
// workspace so that the workspace layout can be restored in a later
// editor session.
type WorkspaceCacheEntry struct {
	Files      []string `json:"files"`
	ActiveFile string   `json:"active_file"`
}

// WorkspaceCache is the on-disk persistence store for workspace state.
// It is saved to ~/.config/wig/workspaces.json on exit and loaded on
// startup or when the workspace picker is opened.
type WorkspaceCache struct {
	Workspaces      map[int]WorkspaceCacheEntry `json:"workspaces"`
	ActiveWorkspace int                         `json:"active_workspace"`
}

// LoadWorkspaceCache reads the workspace cache from disk. If the file
// does not exist or is corrupt, an empty cache is returned.
func LoadWorkspaceCache() *WorkspaceCache {
	home, _ := os.UserHomeDir()
	p := filepath.Join(home, ".config", "wig", "workspaces.json")

	data, err := os.ReadFile(p)
	if err != nil {
		return &WorkspaceCache{Workspaces: make(map[int]WorkspaceCacheEntry)}
	}

	var cache WorkspaceCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return &WorkspaceCache{Workspaces: make(map[int]WorkspaceCacheEntry)}
	}
	if cache.Workspaces == nil {
		cache.Workspaces = make(map[int]WorkspaceCacheEntry)
	}
	return &cache
}

// Save writes the workspace cache to disk.
func (c *WorkspaceCache) Save() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "wig")
	os.MkdirAll(dir, 0755)
	p := filepath.Join(dir, "workspaces.json")
	data, _ := json.MarshalIndent(c, "", "  ")
	os.WriteFile(p, data, 0644)
}

// CaptureWorkspace records the file paths open in a workspace into the
// cache. Only real file-backed buffers are stored; special buffers
// like [Messages] or [git] are skipped. Duplicate paths are deduplicated.
func (c *WorkspaceCache) CaptureWorkspace(num int, ws *wig.Workspace) {
	if ws == nil {
		return
	}
	entry := WorkspaceCacheEntry{}
	seen := make(map[string]bool)
	for _, win := range ws.Windows {
		if win == nil {
			continue
		}
		buf := win.Buffer()
		if buf == nil {
			continue
		}
		fp := buf.FilePath
		if fp == "" || strings.HasPrefix(fp, "[") || seen[fp] {
			continue
		}
		seen[fp] = true
		entry.Files = append(entry.Files, fp)
	}
	if ws.ActiveWindow != nil {
		buf := ws.ActiveWindow.Buffer()
		if buf != nil {
			entry.ActiveFile = buf.FilePath
		}
	}
	c.Workspaces[num] = entry
}

// CaptureAll captures the state of every workspace that has at least
// one window. Workspaces that were never used in this session are
// skipped so that their cached entries from a previous session are
// preserved.
func (c *WorkspaceCache) CaptureAll(editor *wig.Editor) {
	for i := range editor.Workspaces {
		ws := &editor.Workspaces[i]
		if len(ws.Windows) == 0 {
			continue
		}
		c.CaptureWorkspace(i, ws)
	}
	c.ActiveWorkspace = editor.ActiveWorkspace
}

// RestoreWorkspace opens cached files into the target workspace if it
// is currently empty (no real file-backed buffers). If the workspace
// already has file buffers, the cache is skipped to avoid duplicates.
func (c *WorkspaceCache) RestoreWorkspace(editor *wig.Editor, num int) {
	entry, ok := c.Workspaces[num]
	if !ok || len(entry.Files) == 0 {
		return
	}

	ws := editor.GetWorkspace(num)

	// Skip if workspace already has real file buffers
	hasFiles := false
	for _, win := range ws.Windows {
		if win == nil {
			continue
		}
		buf := win.Buffer()
		if buf != nil && buf.FilePath != "" && !strings.HasPrefix(buf.FilePath, "[") {
			hasFiles = true
			break
		}
	}
	if hasFiles {
		return
	}

	var activeBuf *wig.Buffer
	for _, fp := range entry.Files {
		buf, _ := editor.OpenFile(fp)
		if buf != nil && fp == entry.ActiveFile {
			activeBuf = buf
		}
	}

	if activeBuf == nil && len(entry.Files) > 0 {
		for _, b := range editor.Buffers {
			if b.FilePath == entry.Files[0] {
				activeBuf = b
				break
			}
		}
	}

	if activeBuf != nil && ws.ActiveWindow != nil {
		ctx := editor.NewContext()
		ctx.Buf = activeBuf
		ws.ActiveWindow.VisitBuffer(ctx)
	}
}
