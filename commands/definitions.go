package commands

import (
	"github.com/firstrow/mcwig"
)

type CmdDefinition struct {
	Desc string
	Fn   interface{}
}

var AllCommands = map[string]CmdDefinition{
	"CmdFormatBuffer":  {Desc: "", Fn: CmdFormatBuffer},
	"CmdSearchProject": {Desc: "", Fn: CmdSearchProject},
	"CmdJumpForward":   {Desc: "", Fn: mcwig.CmdJumpForward},
	"CmdReloadBuffer":  {Desc: "", Fn: CmdReloadBuffer},
	"CmdNewBuffer":     {Desc: "", Fn: mcwig.CmdNewBuffer},
}

