package wig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
func (c *WorkspaceCache) CaptureWorkspace(num int, ws *Workspace) {
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
		// Drop files that no longer exist on disk so deleted files never
		// poison future session restores.
		if _, err := os.Stat(fp); err != nil {
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

	// A workspace can end a session holding only blank scratch buffers:
	// no-args startups seed ws1 via CmdNewBuffer, and switching through
	// the workspace picker seeds empty windows elsewhere. Overwriting the
	// cached entry with an empty one here silently erases the previous
	// session's file list. Keep the existing cached entry instead;
	// clearing a workspace is done explicitly with Delete in the picker.
	if len(entry.Files) == 0 {
		return
	}
	c.Workspaces[num] = entry
}

// CaptureAll captures the state of every workspace that has at least
// one window. Workspaces that were never used in this session are
// skipped so that their cached entries from a previous session are
// preserved.
func (c *WorkspaceCache) CaptureAll(editor *Editor) {
	for i := range editor.Workspaces {
		ws := &editor.Workspaces[i]
		if len(ws.Windows) == 0 {
			continue
		}
		c.CaptureWorkspace(i, ws)
	}
	c.ActiveWorkspace = editor.ActiveWorkspace
}

// RestoreWorkspace rebuilds the target workspace from cache if it is
// currently empty (no real file-backed buffers): every cached file gets
// its own window so the previous split layout is recreated. If the
// workspace already has file buffers, the cache is skipped to avoid
// duplicates.
func (c *WorkspaceCache) RestoreWorkspace(editor *Editor, num int) {
	entry, ok := c.Workspaces[num]
	if !ok || len(entry.Files) == 0 {
		c.ensureWorkspaceBuffer(editor, num)
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

	// Recreate one window per restored file so the previous split layout
	// reappears. Opening a file alone only adds a hidden buffer to
	// editor.Buffers; without a window displaying it, the file stays
	// invisible on screen ("dismissed").
	type restoredWin struct {
		fp  string
		win *Window
	}
	var restored []restoredWin
	for _, fp := range entry.Files {
		// Files can vanish from disk between capture and restore. Skip
		// them instead of ending up with a nil-buffer window.
		if _, err := os.Stat(fp); err != nil {
			continue
		}
		buf, _ := editor.OpenFile(fp)
		if buf == nil {
			continue
		}
		win := CreateWindow(nil)
		ctx := editor.NewContext()
		ctx.Buf = buf
		win.VisitBuffer(ctx)
		restored = append(restored, restoredWin{fp: fp, win: win})
	}

	if len(restored) == 0 {
		// Every cached file was deleted from disk: still give the
		// workspace's active window a valid buffer.
		c.ensureWorkspaceBuffer(editor, num)
		return
	}

	ws.Windows = make([]*Window, 0, len(restored))
	for _, r := range restored {
		ws.Windows = append(ws.Windows, r.win)
	}
	// Focus the window showing the previously active file; fall back to
	// the first surviving window.
	ws.ActiveWindow = restored[0].win
	for _, r := range restored {
		if r.fp == entry.ActiveFile {
			ws.ActiveWindow = r.win
			break
		}
	}
}

// ensureWorkspaceBuffer attaches a fresh empty buffer to the workspace's
// active window if it currently has none. Without this, restoring a
// workspace whose cached files were deleted from disk leaves a window with
// a nil buffer, which panics on the next input or render tick.
func (c *WorkspaceCache) ensureWorkspaceBuffer(editor *Editor, num int) {
	ws := editor.GetWorkspace(num)
	if ws.ActiveWindow == nil || ws.ActiveWindow.Buffer() != nil {
		return
	}
	ctx := editor.NewContext()
	ctx.Buf = NewBuffer()
	ws.ActiveWindow.VisitBuffer(ctx)
}
