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
	wig.AllCommands["CmdBufferNext"] = wig.CmdDefinition{Desc: "Next buffer", Fn: CmdBufferNext}
	wig.AllCommands["CmdBufferPrev"] = wig.CmdDefinition{Desc: "Previous buffer", Fn: CmdBufferPrev}
	wig.AllCommands["CmdBufferLast"] = wig.CmdDefinition{Desc: "Last buffer", Fn: CmdBufferLast}

	// Command-line basics
	wig.AllCommands["q"] = wig.CmdDefinition{Desc: "Quit", Fn: wig.CmdExit}
	wig.AllCommands["q!"] = wig.CmdDefinition{Desc: "Quit", Fn: wig.CmdExit}
	wig.AllCommands["w"] = wig.CmdDefinition{Desc: "Save", Fn: wig.CmdSaveFile}
	wig.AllCommands["wq"] = wig.CmdDefinition{Desc: "Save and quit", Fn: func(ctx wig.Context) {
		wig.CmdSaveFile(ctx)
		wig.CmdExit(ctx)
	}}
	wig.AllCommands["bd"] = wig.CmdDefinition{Desc: "Delete buffer", Fn: wig.CmdKillBuffer}
	wig.AllCommands["bn"] = wig.CmdDefinition{Desc: "Next buffer", Fn: CmdBufferNext}
	wig.AllCommands["bp"] = wig.CmdDefinition{Desc: "Previous buffer", Fn: CmdBufferPrev}
	wig.AllCommands["bl"] = wig.CmdDefinition{Desc: "Last buffer", Fn: CmdBufferLast}
}
