package wig

type CmdDefinition struct {
	Desc string
	Fn   interface{}
}

var AllCommands = map[string]CmdDefinition{}
