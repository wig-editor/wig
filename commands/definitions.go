package commands

import (
	"github.com/firstrow/wig"
)

type CmdDefinition struct {
	Desc string
	Fn   interface{}
}

var AllCommands = map[string]CmdDefinition{
	"CmdFormatBuffer":  {Desc: "", Fn: CmdFormatBuffer},
	"CmdSearchProject": {Desc: "", Fn: CmdSearchProject},
	"CmdJumpForward":   {Desc: "", Fn: wig.CmdJumpForward},
	"CmdReloadBuffer":  {Desc: "", Fn: CmdReloadBuffer},
	"CmdNewBuffer":     {Desc: "", Fn: wig.CmdNewBuffer},
}

