package commands

import (
	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/ui"
)

type CmdDefinition struct {
	Desc string
	Fn   interface{}
}

var AllCommands = map[string]CmdDefinition{
	"CmdJoinNextLine":               {Desc: "", Fn: mcwig.CmdJoinNextLine},
	"CmdScrollUp":                   {Desc: "", Fn: mcwig.CmdScrollUp},
	"CmdScrollDown":                 {Desc: "", Fn: mcwig.CmdScrollDown},
	"CmdCursorLeft":                 {Desc: "", Fn: mcwig.CmdCursorLeft},
	"CmdCursorRight":                {Desc: "", Fn: mcwig.CmdCursorRight},
	"CmdCursorLineUp":               {Desc: "", Fn: mcwig.CmdCursorLineUp},
	"CmdCursorLineDown":             {Desc: "", Fn: mcwig.CmdCursorLineDown},
	"CmdCursorBeginningOfTheLine":   {Desc: "", Fn: mcwig.CmdCursorBeginningOfTheLine},
	"CmdCursorFirstNonBlank":        {Desc: "", Fn: mcwig.CmdCursorFirstNonBlank},
	"CmdInsertMode":                 {Desc: "", Fn: mcwig.CmdInsertMode},
	"CmdVisualMode":                 {Desc: "", Fn: mcwig.CmdVisualMode},
	"CmdVisualLineMode":             {Desc: "", Fn: mcwig.CmdVisualLineMode},
	"CmdInsertModeAfter":            {Desc: "", Fn: mcwig.CmdInsertModeAfter},
	"CmdNormalMode":                 {Desc: "", Fn: mcwig.CmdNormalMode},
	"CmdGotoLine0":                  {Desc: "", Fn: mcwig.CmdGotoLine0},
	"CmdGotoLineEnd":                {Desc: "", Fn: mcwig.CmdGotoLineEnd},
	"CmdForwardWord":                {Desc: "", Fn: mcwig.CmdForwardWord},
	"CmdBackwardWord":               {Desc: "", Fn: mcwig.CmdBackwardWord},
	"CmdReplaceChar":                {Desc: "", Fn: mcwig.CmdReplaceChar},
	"CmdForwardToChar":              {Desc: "", Fn: mcwig.CmdForwardToChar},
	"CmdForwardBeforeChar":          {Desc: "", Fn: mcwig.CmdForwardBeforeChar},
	"CmdBackwardChar":               {Desc: "", Fn: mcwig.CmdBackwardChar},
	"CmdDeleteCharForward":          {Desc: "", Fn: mcwig.CmdDeleteCharForward},
	"CmdDeleteCharBackward":         {Desc: "", Fn: mcwig.CmdDeleteCharBackward},
	"CmdAppendLine":                 {Desc: "", Fn: mcwig.CmdAppendLine},
	"CmdNewLine":                    {Desc: "", Fn: mcwig.CmdNewLine},
	"CmdLineOpenBelow":              {Desc: "", Fn: mcwig.CmdLineOpenBelow},
	"CmdLineOpenAbove":              {Desc: "", Fn: mcwig.CmdLineOpenAbove},
	"CmdDeleteLine":                 {Desc: "", Fn: mcwig.CmdDeleteLine},
	"CmdDeleteWord":                 {Desc: "", Fn: mcwig.CmdDeleteWord},
	"CmdChangeWord":                 {Desc: "", Fn: mcwig.CmdChangeWord},
	"CmdChangeTo":                   {Desc: "", Fn: mcwig.CmdChangeTo},
	"CmdChangeBefore":               {Desc: "", Fn: mcwig.CmdChangeBefore},
	"CmdChangeEndOfLine":            {Desc: "", Fn: mcwig.CmdChangeEndOfLine},
	"CmdChangeLine":                 {Desc: "", Fn: mcwig.CmdChangeLine},
	"CmdDeleteTo":                   {Desc: "", Fn: mcwig.CmdDeleteTo},
	"CmdDeleteBefore":               {Desc: "", Fn: mcwig.CmdDeleteBefore},
	"CmdSelectionChange":            {Desc: "", Fn: mcwig.CmdSelectionChange},
	"CmdSelectinDelete":             {Desc: "", Fn: mcwig.CmdSelectinDelete},
	"CmdSaveFile":                   {Desc: "", Fn: mcwig.CmdSaveFile},
	"CmdWindowVSplit":               {Desc: "", Fn: mcwig.CmdWindowVSplit},
	"CmdWindowNext":                 {Desc: "", Fn: mcwig.CmdWindowNext},
	"CmdWindowToggleLayout":         {Desc: "", Fn: mcwig.CmdWindowToggleLayout},
	"CmdWindowClose":                {Desc: "", Fn: mcwig.CmdWindowClose},
	"CmdExit":                       {Desc: "", Fn: mcwig.CmdExit},
	"CmdYank":                       {Desc: "", Fn: mcwig.CmdYank},
	"CmdYankPut":                    {Desc: "", Fn: mcwig.CmdYankPut},
	"CmdYankPutBefore":              {Desc: "", Fn: mcwig.CmdYankPutBefore},
	"CmdKillBuffer":                 {Desc: "", Fn: mcwig.CmdKillBuffer},
	"CmdEnsureCursorVisible":        {Desc: "", Fn: mcwig.CmdEnsureCursorVisible},
	"CmdBufferPicker":               {Desc: "", Fn: CmdBufferPicker},
	"CmdExecute":                    {Desc: "", Fn: CmdExecute},
	"CmdFindProjectFilePicker":      {Desc: "", Fn: CmdFindProjectFilePicker},
	"CmdCurrentBufferDirFilePicker": {Desc: "", Fn: CmdCurrentBufferDirFilePicker},
	"CmdLineInit":                   {Desc: "", Fn: ui.CmdLineInit},
	"CmdSearchPromptInit":           {Desc: "", Fn: ui.CmdSearchPromptInit},
}
