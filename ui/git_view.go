package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/git"
)

// DiffHighlighter provides syntax highlighting for git diff and commit
// inspection buffers.
type DiffHighlighter struct {
	Buf *wig.Buffer
}

func (h *DiffHighlighter) Build()                          {}
func (h *DiffHighlighter) TextChanged(wig.EventTextChange) {}

func (h *DiffHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	if h.Buf == nil {
		return nil
	}

	nodes := wig.List[wig.HighlighterNode]{}

	line := wig.CursorLineByNum(h.Buf, int(startLine))
	for lineNum := startLine; line != nil && lineNum <= endLine; lineNum++ {
		text := line.Value.String()
		lineLen := uint32(len([]rune(text)))

		if lineLen > 0 {
			var nodeName string
			switch {
			case strings.HasPrefix(text, "diff --git"), strings.HasPrefix(text, "index "),
				strings.HasPrefix(text, "commit "), strings.HasPrefix(text, "Author:"),
				strings.HasPrefix(text, "Date:"):
				nodeName = "ui.statusline"
			case strings.HasPrefix(text, "---") || strings.HasPrefix(text, "+++"):
				nodeName = "ui.linenr"
			case strings.HasPrefix(text, "@@"):
				nodeName = "ui.linenr.selected"
			case strings.HasPrefix(text, "+"):
				nodeName = "diff.plus"
			case strings.HasPrefix(text, "-"):
				nodeName = "diff.minus"
			}

			if nodeName != "" {
				nodes.PushBack(wig.HighlighterNode{
					NodeName:  nodeName,
					StartLine: lineNum,
					StartChar: 0,
					EndLine:   lineNum,
					EndChar:   lineLen,
				})
			}
		}

		line = line.Next()
	}

	if nodes.First() == nil {
		return nil
	}

	return &wig.HighlighterCursor{Cursor: nodes.First()}
}

// ── git status popup widget ──────────────────────────────

// gitViewEntry represents a single line in the left panel of the git view popup.
type gitViewEntry struct {
	kind     string // "header", "separator", "empty", "blank", "file", "branch", "stash"
	item     git.StatusItem
	filePath string
	text     string
}

// GitViewWidget is a dual-panel popup for git status and diff.
// Left panel (40%) shows git status items; right panel (60%) shows
// the diff of the currently selected item.
type GitViewWidget struct {
	e            *wig.Editor
	keymap       *wig.KeyHandler
	entries      []gitViewEntry
	activeIdx    int
	scrollOffset int
	diffLines    []string
	diffScroll   int
	diffTitle    string
	pendingStash *git.StatusItem
}

func (w *GitViewWidget) Plane() wig.RenderPlane  { return wig.PlaneEditor }
func (w *GitViewWidget) Mode() wig.Mode          { return wig.MODE_NORMAL }
func (w *GitViewWidget) Keymap() *wig.KeyHandler { return w.keymap }

// GitViewPopupInit creates and pushes the git view popup widget.
func GitViewPopupInit(ctx wig.Context) *GitViewWidget {
	w := &GitViewWidget{
		e:         ctx.Editor,
		activeIdx: 0,
	}
	w.refresh()
	w.updateDiff()

	w.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc": func(ctx wig.Context) {
				if w.pendingStash != nil {
					w.pendingStash = nil
					ctx.Editor.EchoMessage("Cancelled")
					ctx.Editor.Redraw()
					return
				}
				ctx.Editor.PopUi()
			},
			"q": func(ctx wig.Context) {
				ctx.Editor.PopUi()
			},
			"j":       func(ctx wig.Context) { w.moveDown(ctx) },
			"k":       func(ctx wig.Context) { w.moveUp(ctx) },
			"Down":    func(ctx wig.Context) { w.moveDown(ctx) },
			"Up":      func(ctx wig.Context) { w.moveUp(ctx) },
			"Tab":     func(ctx wig.Context) { w.moveDown(ctx) },
			"Backtab": func(ctx wig.Context) { w.moveUp(ctx) },
			"Home":    func(ctx wig.Context) { w.moveToFirst(ctx) },
			"End":     func(ctx wig.Context) { w.moveToLast(ctx) },
			"Enter":   func(ctx wig.Context) { w.handleEnter(ctx) },
			"s":       func(ctx wig.Context) { w.handleStage(ctx) },
			"d":       func(ctx wig.Context) { w.handleDiffAction(ctx) },
			"z":       func(ctx wig.Context) { w.handleStash(ctx) },
			"r":       func(ctx wig.Context) { w.handleRefresh(ctx) },
			"c":       func(ctx wig.Context) { w.handleCommit(ctx) },
			"p":       func(ctx wig.Context) { w.handlePush(ctx) },
			"S":       func(ctx wig.Context) { w.handleStashPop(ctx) },
			"ctrl+d":  func(ctx wig.Context) { w.scrollDiffDown(ctx, 10) },
			"ctrl+u":  func(ctx wig.Context) { w.scrollDiffUp(ctx, 10) },
		},
	})

	ctx.Editor.PushUi(w)
	ctx.Editor.Redraw()
	return w
}

// ── data ──

func (w *GitViewWidget) refresh() {
	w.pendingStash = nil
	items := git.StatusItems()

	w.entries = make([]gitViewEntry, 0, len(items))
	for _, it := range items {
		var text string
		kind := it.Type
		switch it.Type {
		case "header":
			text = fmt.Sprintf("── %s ", it.Label)
			pad := 50 - len([]rune(text))
			if pad > 0 {
				text += strings.Repeat("─", pad)
			}
		case "separator", "blank":
			text = ""
		case "empty":
			text = fmt.Sprintf("   %s", it.Label)
		case "file":
			text = fmt.Sprintf("  %s  %s", it.Code, it.FilePath)
		case "branch":
			text = fmt.Sprintf("  %s", it.Label)
		case "stash":
			text = fmt.Sprintf("  %s: %s", it.StashRef, it.Label)
		default:
			text = it.Label
		}

		w.entries = append(w.entries, gitViewEntry{
			kind:     kind,
			item:     it,
			filePath: it.FilePath,
			text:     text,
		})
	}

	w.activeIdx = 0
	for i, e := range w.entries {
		if e.kind == "file" || e.kind == "branch" || e.kind == "stash" {
			w.activeIdx = i
			break
		}
	}

	w.scrollOffset = 0
	w.diffScroll = 0
}

func (w *GitViewWidget) updateDiff() {
	w.diffLines = nil
	w.diffScroll = 0
	w.diffTitle = ""

	if w.activeIdx < 0 || w.activeIdx >= len(w.entries) {
		w.diffLines = []string{"(no selection)"}
		return
	}

	entry := w.entries[w.activeIdx]

	switch entry.kind {
	case "file":
		diffLines := git.DiffLines(entry.item)
		if len(diffLines) == 0 {
			w.diffLines = []string{"(no diff)"}
		} else {
			w.diffLines = diffLines
		}
		w.diffTitle = entry.filePath
	case "branch":
		w.diffLines = []string{
			fmt.Sprintf("Branch: %s", entry.item.Label),
			"",
			"Press Enter to checkout this branch.",
		}
		w.diffTitle = entry.item.StashRef
	case "stash":
		diffOut := git.Run("stash", "show", "-p", entry.item.StashRef)
		if strings.TrimSpace(diffOut) == "" {
			w.diffLines = []string{"(empty stash)"}
		} else {
			w.diffLines = strings.Split(diffOut, "\n")
		}
		w.diffTitle = entry.item.StashRef
	default:
		w.diffLines = []string{"(select a file, branch, or stash)"}
	}
}

// ── navigation ──

func (w *GitViewWidget) isSelectable(idx int) bool {
	if idx < 0 || idx >= len(w.entries) {
		return false
	}
	k := w.entries[idx].kind
	return k == "file" || k == "branch" || k == "stash"
}

func (w *GitViewWidget) moveDown(ctx wig.Context) {
	for i := w.activeIdx + 1; i < len(w.entries); i++ {
		if w.isSelectable(i) {
			w.activeIdx = i
			w.updateDiff()
			ctx.Editor.Redraw()
			return
		}
	}
}

func (w *GitViewWidget) moveUp(ctx wig.Context) {
	for i := w.activeIdx - 1; i >= 0; i-- {
		if w.isSelectable(i) {
			w.activeIdx = i
			w.updateDiff()
			ctx.Editor.Redraw()
			return
		}
	}
}

func (w *GitViewWidget) moveToFirst(ctx wig.Context) {
	for i := 0; i < len(w.entries); i++ {
		if w.isSelectable(i) {
			w.activeIdx = i
			w.updateDiff()
			ctx.Editor.Redraw()
			return
		}
	}
}

func (w *GitViewWidget) moveToLast(ctx wig.Context) {
	for i := len(w.entries) - 1; i >= 0; i-- {
		if w.isSelectable(i) {
			w.activeIdx = i
			w.updateDiff()
			ctx.Editor.Redraw()
			return
		}
	}
}

func (w *GitViewWidget) scrollDiffDown(ctx wig.Context, n int) {
	w.diffScroll += n
	ctx.Editor.Redraw()
}

func (w *GitViewWidget) scrollDiffUp(ctx wig.Context, n int) {
	w.diffScroll -= n
	if w.diffScroll < 0 {
		w.diffScroll = 0
	}
	ctx.Editor.Redraw()
}

// ── action handlers ──

func (w *GitViewWidget) handleEnter(ctx wig.Context) {
	if w.activeIdx < 0 || w.activeIdx >= len(w.entries) {
		return
	}
	entry := w.entries[w.activeIdx]

	switch entry.kind {
	case "file":
		w.pendingStash = nil
		filePath := entry.filePath
		if !filepath.IsAbs(filePath) {
			rootDir := ctx.Editor.Projects.GetRoot()
			filePath = filepath.Join(rootDir, filePath)
		}
		ctx.Editor.PopUi()
		buf, err := ctx.Editor.OpenFile(filePath)
		if err != nil {
			ctx.Editor.EchoMessage("Cannot open: " + err.Error())
			return
		}
		ctx.Buf = buf
		ctx.Editor.ActiveWindow().VisitBuffer(ctx)
	case "branch":
		w.pendingStash = nil
		err := git.SwitchBranch(entry.item)
		branchName := entry.item.StashRef
		if branchName == "" {
			branchName = entry.item.FilePath
		}
		if err != nil {
			ctx.Editor.EchoMessage(err.Error())
		} else {
			for _, b := range ctx.Editor.Buffers {
				if b.FilePath != "" && !strings.HasPrefix(b.FilePath, "[") {
					_ = wig.BufferReloadFile(b)
					if b.Highlighter != nil {
						b.Highlighter.Build()
					}
				}
			}
			w.refresh()
			w.updateDiff()
			ctx.Editor.EchoMessage(fmt.Sprintf("Switched to branch '%s'", branchName))
		}
		ctx.Editor.Redraw()
	case "stash":
		itemCopy := entry.item
		w.pendingStash = &itemCopy
		ctx.Editor.EchoMessage(fmt.Sprintf("Stash %s: [S] pop  [d] drop  [Esc] cancel", entry.item.StashRef))
		ctx.Editor.Redraw()
	}
}

func (w *GitViewWidget) handleStage(ctx wig.Context) {
	w.pendingStash = nil
	if w.activeIdx < 0 || w.activeIdx >= len(w.entries) {
		return
	}
	entry := w.entries[w.activeIdx]
	if entry.kind != "file" {
		return
	}
	oldPath := entry.filePath
	git.StageItem(entry.item)
	w.refresh()
	for i, e := range w.entries {
		if e.filePath == oldPath {
			w.activeIdx = i
			break
		}
	}
	w.updateDiff()
	ctx.Editor.EchoMessage("Toggled stage for: " + oldPath)
	ctx.Editor.Redraw()
}

func (w *GitViewWidget) handleDiffAction(ctx wig.Context) {
	if w.pendingStash != nil {
		stashRef := w.pendingStash.StashRef
		git.StashAction(*w.pendingStash, "drop")
		w.pendingStash = nil
		w.refresh()
		w.updateDiff()
		ctx.Editor.EchoMessage("Stash dropped: " + stashRef)
		ctx.Editor.Redraw()
		return
	}
	w.diffScroll += 5
	ctx.Editor.Redraw()
}

func (w *GitViewWidget) handleStash(ctx wig.Context) {
	w.pendingStash = nil
	git.StashUnstaged()
	w.refresh()
	w.updateDiff()
	ctx.Editor.EchoMessage("Stashed unstaged changes")
	ctx.Editor.Redraw()
}

func (w *GitViewWidget) handleRefresh(ctx wig.Context) {
	w.pendingStash = nil
	w.refresh()
	w.updateDiff()
	ctx.Editor.EchoMessage("Git status refreshed")
	ctx.Editor.Redraw()
}

func (w *GitViewWidget) handleCommit(ctx wig.Context) {
	ctx.Editor.PopUi()
	GitShowCommitBuffer(ctx)
}

func (w *GitViewWidget) handlePush(ctx wig.Context) {
	ctx.Editor.EchoMessage("running: git push origin HEAD")
	ctx.Editor.Redraw()
	git.Run("push", "origin", "HEAD")
	ctx.Editor.EchoMessage("done")
	ctx.Editor.Redraw()
}

func (w *GitViewWidget) handleStashPop(ctx wig.Context) {
	if w.pendingStash != nil {
		stashRef := w.pendingStash.StashRef
		git.StashAction(*w.pendingStash, "pop")
		w.pendingStash = nil
		w.refresh()
		w.updateDiff()
		ctx.Editor.EchoMessage("Stash popped: " + stashRef)
		ctx.Editor.Redraw()
	}
}

// ── rendering ──

func (w *GitViewWidget) Render(view wig.View) {
	vw, vh := view.Size()

	boxW := int(float32(vw) * 0.92)
	boxH := int(float32(vh) * 0.88)
	if boxW > vw-2 {
		boxW = vw - 2
	}
	if boxH > vh-2 {
		boxH = vh - 2
	}
	if boxW < 40 {
		boxW = 40
	}
	if boxH < 10 {
		boxH = 10
	}

	x := (vw - boxW) / 2
	y := (vh - boxH) / 2

	bgStyle := wig.Color("ui.background")
	borderStyle := wig.Color("ui.window.border")

	for row := 0; row < boxH; row++ {
		view.SetContent(x, y+row, strings.Repeat(" ", boxW), bgStyle)
	}

	for col := 0; col < boxW; col++ {
		view.SetContent(x+col, y, "─", borderStyle)
		view.SetContent(x+col, y+boxH-1, "─", borderStyle)
	}
	for row := 0; row < boxH; row++ {
		view.SetContent(x, y+row, "│", borderStyle)
		view.SetContent(x+boxW-1, y+row, "│", borderStyle)
	}
	view.SetContent(x, y, "┌", borderStyle)
	view.SetContent(x+boxW-1, y, "┐", borderStyle)
	view.SetContent(x, y+boxH-1, "└", borderStyle)
	view.SetContent(x+boxW-1, y+boxH-1, "┘", borderStyle)

	view.SetContent(x+2, y, " Git Status ", borderStyle)

	hint := " [j/k] Move  [Enter] Open  [s] Stage  [c] Commit  [p] Push  [z] Stash  [r] Refresh  [Esc] Close "
	hintRunes := []rune(hint)
	maxHint := boxW - 16
	if maxHint > 0 && len(hintRunes) > maxHint {
		hint = string(hintRunes[:maxHint])
	}
	hintX := x + boxW - 2 - len([]rune(hint))
	if hintX > x+14 {
		view.SetContent(hintX, y, hint, borderStyle)
	}

	innerX := x + 1
	innerY := y + 1
	innerW := boxW - 2
	innerH := boxH - 2

	leftW := innerW * 4 / 10
	if leftW < 20 {
		leftW = 20
	}
	rightW := innerW - leftW - 1
	if rightW < 20 {
		rightW = 20
		leftW = innerW - rightW - 1
		if leftW < 10 {
			leftW = 10
		}
	}

	sepX := innerX + leftW
	for row := 0; row < innerH; row++ {
		view.SetContent(sepX, innerY+row, "│", borderStyle)
	}

	w.clampLeftScroll(innerH)
	w.clampDiffScroll(innerH - 2)

	w.renderLeftPanel(view, innerX, innerY, leftW, innerH)

	rightX := sepX + 1
	w.renderRightPanel(view, rightX, innerY, rightW, innerH)
}

func (w *GitViewWidget) clampLeftScroll(height int) {
	if w.activeIdx < w.scrollOffset {
		w.scrollOffset = w.activeIdx
	}
	if w.activeIdx >= w.scrollOffset+height {
		w.scrollOffset = w.activeIdx - height + 1
	}
	if w.scrollOffset < 0 {
		w.scrollOffset = 0
	}
}

func (w *GitViewWidget) clampDiffScroll(height int) {
	if height < 0 {
		height = 0
	}
	maxScroll := len(w.diffLines) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if w.diffScroll > maxScroll {
		w.diffScroll = maxScroll
	}
	if w.diffScroll < 0 {
		w.diffScroll = 0
	}
}

func (w *GitViewWidget) renderLeftPanel(view wig.View, x, y, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	bgStyle := wig.Color("ui.background")

	for row := 0; row < height; row++ {
		entryIdx := w.scrollOffset + row
		var entry *gitViewEntry
		if entryIdx >= 0 && entryIdx < len(w.entries) {
			entry = &w.entries[entryIdx]
		}

		isActive := entryIdx == w.activeIdx && entry != nil
		selectable := entry != nil && (entry.kind == "file" || entry.kind == "branch" || entry.kind == "stash")

		var style = bgStyle
		var text string

		if entry != nil {
			switch entry.kind {
			case "header":
				style = wig.Color("ui.statusline")
				text = entry.text
			case "empty":
				style = wig.Color("ui.linenr")
				text = entry.text
			case "file":
				style = wig.Color("ui.text")
				text = entry.text
			case "branch":
				if entry.item.Status == "active_branch" {
					style = wig.Color("ui.linenr.selected")
				} else {
					style = wig.Color("ui.text")
				}
				text = entry.text
			case "stash":
				style = wig.Color("ui.text")
				text = entry.text
			}

			if isActive && selectable {
				style = wig.ApplyBg("ui.selection.primary", style)
			}
		}

		runes := []rune(text)
		if len(runes) > width {
			runes = runes[:width]
		}
		padded := string(runes)
		if len(runes) < width {
			padded += strings.Repeat(" ", width-len(runes))
		}

		view.SetContent(x, y+row, padded, style)
	}
}

func (w *GitViewWidget) renderRightPanel(view wig.View, x, y, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	bgStyle := wig.Color("ui.background")

	if len(w.diffLines) == 0 {
		w.diffLines = []string{"(no diff)"}
	}

	titleText := w.diffTitle
	if titleText == "" {
		titleText = "Diff"
	}
	titleStyle := wig.Color("ui.statusline")
	titleRunes := []rune(titleText)
	if len(titleRunes) > width-1 {
		titleRunes = titleRunes[:width-1]
	}
	titlePadded := string(titleRunes) + strings.Repeat(" ", width-len(titleRunes))
	view.SetContent(x, y, titlePadded, titleStyle)

	sepStyle := wig.Color("ui.window.border")
	if height > 1 {
		view.SetContent(x, y+1, strings.Repeat("─", width), sepStyle)
	}

	bodyY := y + 2
	bodyH := height - 2
	if bodyH < 0 {
		bodyH = 0
	}

	for row := 0; row < bodyH; row++ {
		lineIdx := w.diffScroll + row
		if lineIdx >= len(w.diffLines) {
			view.SetContent(x, bodyY+row, strings.Repeat(" ", width), bgStyle)
			continue
		}

		line := strings.ReplaceAll(w.diffLines[lineIdx], "\t", "    ")
		style := gitViewDiffLineStyle(line)

		runes := []rune(line)
		if len(runes) > width {
			runes = runes[:width]
		}
		padded := string(runes)
		if len(runes) < width {
			padded += strings.Repeat(" ", width-len(runes))
		}

		view.SetContent(x, bodyY+row, padded, style)
	}
}

func gitViewDiffLineStyle(line string) tcell.Style {
	switch {
	case strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index "):
		return wig.Color("ui.statusline")
	case strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ "):
		return wig.Color("ui.linenr")
	case strings.HasPrefix(line, "@@"):
		return wig.Color("ui.linenr.selected")
	case strings.HasPrefix(line, "+"):
		return wig.Color("diff.plus")
	case strings.HasPrefix(line, "-"):
		return wig.Color("diff.minus")
	default:
		return wig.Color("default")
	}
}

// ── commit message buffer ───────────────────────────────

func exitModeOrClose(ctx wig.Context) {
	if ctx.Buf.Mode() != wig.MODE_NORMAL {
		wig.CmdNormalMode(ctx)
		return
	}

	wig.CmdKillBuffer(ctx)
}

func GitShowCommitBuffer(ctx wig.Context) {
	contents := git.Run("status", "-v")
	diffBufName := "[git: edit commit message]"
	dBuf := ctx.Editor.BufferFindByFilePath(diffBufName, true)
	dBuf.ResetLines()

	dBuf.Append("")
	dBuf.Append("")
	dBuf.Append("# Please enter your commit message and press ctrl+c for commit or Esc for exit.")
	dBuf.Append("")
	for l := range strings.SplitSeq(contents, "\n") {
		dBuf.Append(fmt.Sprintf("# %s", l))
	}
	dBuf.Highlighter = &DiffHighlighter{Buf: dBuf}

	dBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc":    exitModeOrClose,
			"ctrl+c": gitCommitFinish,
		},
		wig.MODE_INSERT: wig.KeyMap{
			"Esc":    exitModeOrClose,
			"ctrl+c": gitCommitFinish,
		},
	})

	ctx.Buf = dBuf
	wig.CmdEnterInsertMode(ctx)
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: 0, Char: 0})
}

func gitCommitFinish(ctx wig.Context) {
	ctx.Buf.FilePath = "/tmp/commit_msg.txt"
	ctx.Buf.Save()
	git.Run("commit", "-F", "/tmp/commit_msg.txt", "--cleanup=strip")
	wig.CmdKillBuffer(ctx)
	wig.EditorInst.EchoMessage("commit done")
}
