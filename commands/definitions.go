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
}
