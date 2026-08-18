package commands

import (
	"github.com/firstrow/wig"
)

type CmdDefinition struct {
	Desc string
	Fn   interface{}
}

var AllCommands = map[string]CmdDefinition{}

func init() {
	AllCommands["CmdFormatBuffer"] = CmdDefinition{Desc: "Format buffer", Fn: CmdFormatBuffer}
	AllCommands["CmdSearchProject"] = CmdDefinition{Desc: "Search project", Fn: CmdSearchProject}
	AllCommands["CmdJumpForward"] = CmdDefinition{Desc: "Jump forward", Fn: wig.CmdJumpForward}
	AllCommands["CmdReloadBuffer"] = CmdDefinition{Desc: "Reload buffer", Fn: CmdReloadBuffer}
	AllCommands["CmdNewBuffer"] = CmdDefinition{Desc: "New buffer", Fn: wig.CmdNewBuffer}
	AllCommands["CmdExit"] = CmdDefinition{Desc: "Quit editor", Fn: wig.CmdExit}
	AllCommands["CmdCommandPalettePicker"] = CmdDefinition{Desc: "Command palette", Fn: CmdCommandPalettePicker}
	AllCommands["CmdFindProjectFilePicker"] = CmdDefinition{Desc: "Find file", Fn: CmdFindProjectFilePicker}
	AllCommands["CmdBufferPicker"] = CmdDefinition{Desc: "Buffer picker", Fn: CmdBufferPicker}
	AllCommands["CmdBufferNext"] = CmdDefinition{Desc: "Next buffer", Fn: CmdBufferNext}
	AllCommands["CmdBufferPrev"] = CmdDefinition{Desc: "Previous buffer", Fn: CmdBufferPrev}
	AllCommands["CmdBufferLast"] = CmdDefinition{Desc: "Last buffer", Fn: CmdBufferLast}

	// Command-line basics
	AllCommands["q"] = CmdDefinition{Desc: "Quit", Fn: wig.CmdExit}
	AllCommands["q!"] = CmdDefinition{Desc: "Quit", Fn: wig.CmdExit}
	AllCommands["w"] = CmdDefinition{Desc: "Save", Fn: wig.CmdSaveFile}
	AllCommands["wq"] = CmdDefinition{Desc: "Save and quit", Fn: func(ctx wig.Context) {
		wig.CmdSaveFile(ctx)
		wig.CmdExit(ctx)
	}}
	AllCommands["bd"] = CmdDefinition{Desc: "Delete buffer", Fn: wig.CmdKillBuffer}
	AllCommands["bn"] = CmdDefinition{Desc: "Next buffer", Fn: CmdBufferNext}
	AllCommands["bp"] = CmdDefinition{Desc: "Previous buffer", Fn: CmdBufferPrev}
	AllCommands["bl"] = CmdDefinition{Desc: "Last buffer", Fn: CmdBufferLast}
}
