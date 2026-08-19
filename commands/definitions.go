package commands

import (
	"github.com/firstrow/wig"
)

func init() {
	wig.AllCommands["CmdFormatBuffer"] = wig.CmdDefinition{Desc: "Format buffer", Fn: CmdFormatBuffer}
	wig.AllCommands["CmdSearchProject"] = wig.CmdDefinition{Desc: "Search project", Fn: CmdSearchProject}
	wig.AllCommands["CmdJumpForward"] = wig.CmdDefinition{Desc: "Jump forward", Fn: wig.CmdJumpForward}
	wig.AllCommands["CmdReloadBuffer"] = wig.CmdDefinition{Desc: "Reload buffer", Fn: CmdReloadBuffer}
	wig.AllCommands["CmdNewBuffer"] = wig.CmdDefinition{Desc: "New buffer", Fn: wig.CmdNewBuffer}
	wig.AllCommands["CmdExit"] = wig.CmdDefinition{Desc: "Quit editor", Fn: wig.CmdExit}
	wig.AllCommands["CmdCommandPalettePicker"] = wig.CmdDefinition{Desc: "Command palette", Fn: CmdCommandPalettePicker}
	wig.AllCommands["CmdFindProjectFilePicker"] = wig.CmdDefinition{Desc: "Find file", Fn: CmdFindProjectFilePicker}
	wig.AllCommands["CmdBufferPicker"] = wig.CmdDefinition{Desc: "Buffer picker", Fn: CmdBufferPicker}
	wig.AllCommands["CmdBufferNext"] = wig.CmdDefinition{Desc: "Next buffer", Fn: wig.CmdBufferNext}
	wig.AllCommands["CmdBufferPrev"] = wig.CmdDefinition{Desc: "Previous buffer", Fn: wig.CmdBufferPrev}
	wig.AllCommands["CmdBufferLast"] = wig.CmdDefinition{Desc: "Last buffer", Fn: wig.CmdBufferLast}
	wig.AllCommands["CmdWindowNext"] = wig.CmdDefinition{Desc: "Next window", Fn: wig.CmdWindowNext}
	wig.AllCommands["CmdKillBuffer"] = wig.CmdDefinition{Desc: "Kill buffer", Fn: wig.CmdKillBuffer}
	wig.AllCommands["CmdSaveFile"] = wig.CmdDefinition{Desc: "Save file", Fn: wig.CmdSaveFile}
	wig.AllCommands["CmdGitHunkNext"] = wig.CmdDefinition{Desc: "Next git hunk", Fn: CmdGitHunkNext}
	wig.AllCommands["CmdGitHunkPrev"] = wig.CmdDefinition{Desc: "Previous git hunk", Fn: CmdGitHunkPrev}
	wig.AllCommands["CmdGitHunkRevert"] = wig.CmdDefinition{Desc: "Revert git hunk", Fn: CmdGitHunkRevert}
	wig.AllCommands["CmdDeleteLine"] = wig.CmdDefinition{Desc: "Delete line", Fn: wig.CmdDeleteLine}
	wig.AllCommands["CmdGitHunkPreview"] = wig.CmdDefinition{Desc: "Preview git hunk", Fn: CmdGitHunkPreview}
	wig.AllCommands["CmdMRUBufferPicker"] = wig.CmdDefinition{Desc: "MRU Buffer Picker", Fn: CmdMRUBufferPicker}
	wig.AllCommands["CmdCheckHealth"] = wig.CmdDefinition{Desc: "Check health of dependencies", Fn: CmdCheckHealth}
	wig.AllCommands["checkhealth"] = wig.CmdDefinition{Desc: "Check health of dependencies", Fn: CmdCheckHealth}

	// Command-line basics
	wig.AllCommands["q"] = wig.CmdDefinition{Desc: "Quit", Fn: wig.CmdExit}
	wig.AllCommands["q!"] = wig.CmdDefinition{Desc: "Quit", Fn: wig.CmdExit}
	wig.AllCommands["w"] = wig.CmdDefinition{Desc: "Save", Fn: wig.CmdSaveFile}
	wig.AllCommands["wq"] = wig.CmdDefinition{Desc: "Save and quit", Fn: func(ctx wig.Context) {
		wig.CmdSaveFile(ctx)
		wig.CmdExit(ctx)
	}}
	wig.AllCommands["bd"] = wig.CmdDefinition{Desc: "Delete buffer", Fn: wig.CmdKillBuffer}
	wig.AllCommands["bn"] = wig.CmdDefinition{Desc: "Next buffer", Fn: wig.CmdBufferNext}
	wig.AllCommands["bp"] = wig.CmdDefinition{Desc: "Previous buffer", Fn: wig.CmdBufferPrev}
	wig.AllCommands["bl"] = wig.CmdDefinition{Desc: "Last buffer", Fn: wig.CmdBufferLast}
	wig.AllCommands["CmdGitView"] = wig.CmdDefinition{Desc: "Git status panel", Fn: CmdGitView}
	wig.AllCommands["gs"] = wig.CmdDefinition{Desc: "Git status", Fn: CmdGitView}
	wig.AllCommands["CmdGitBlame"] = wig.CmdDefinition{Desc: "Git blame", Fn: CmdGitBlame}
	wig.AllCommands["blame"] = wig.CmdDefinition{Desc: "Git blame", Fn: CmdGitBlame}

	// Window management commands
	wig.AllCommands["vs"] = wig.CmdDefinition{Desc: "Vertical split", Fn: wig.CmdWindowVSplit}
	wig.AllCommands["sp"] = wig.CmdDefinition{Desc: "Horizontal split", Fn: wig.CmdWindowVSplit} // wig only has VSplit for now
	wig.AllCommands["only"] = wig.CmdDefinition{Desc: "Close other windows", Fn: wig.CmdWindowCloseOther}
	wig.AllCommands["close"] = wig.CmdDefinition{Desc: "Close window", Fn: wig.CmdWindowClose}
}
