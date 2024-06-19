package config

import (
	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/commands"
	"github.com/firstrow/mcwig/ui"
)

func DefaultKeyMap() mcwig.ModeKeyMap {
	return mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			"ctrl+e": mcwig.CmdScrollDown,
			"ctrl+y": mcwig.CmdScrollUp,
			"h":      mcwig.CmdCursorLeft,
			"l":      mcwig.CmdCursorRight,
			"j":      mcwig.CmdCursorLineDown,
			"k":      mcwig.CmdCursorLineUp,
			"i":      mcwig.CmdInsertMode,
			"v":      mcwig.CmdVisualMode,
			"V":      mcwig.CmdVisualLineMode,
			"a":      mcwig.CmdInsertModeAfter,
			"A":      mcwig.CmdAppendLine,
			"w":      mcwig.CmdForwardWord,
			"b":      mcwig.CmdBackwardWord,
			"x":      mcwig.CmdDeleteCharForward,
			"X":      mcwig.CmdDeleteCharBackward,
			"^":      mcwig.CmdCursorFirstNonBlank,
			"$":      mcwig.CmdGotoLineEnd,
			"0":      mcwig.CmdCursorBeginningOfTheLine,
			"o":      mcwig.CmdLineOpenBelow,
			"O":      mcwig.CmdLineOpenAbove,
			"J":      mcwig.CmdJoinNextLine,
			"p":      mcwig.CmdYankPut,
			"P":      mcwig.CmdYankPutBefore,
			"r":      mcwig.CmdReplaceChar,
			"f":      mcwig.CmdForwardToChar,
			"t":      mcwig.CmdForwardBeforeChar,
			"F":      mcwig.CmdBackwardChar,
			"n":      mcwig.CmdSearchNext,
			"N":      mcwig.CmdSearchPrev,
			":":      ui.CommandLineInit,
			";":      commands.CmdBufferPicker,
			"/":      ui.SearchPromptInit,
			"c": mcwig.KeyMap{
				"c": mcwig.CmdChangeLine,
				"w": mcwig.CmdChangeWord,
				"f": mcwig.CmdChangeTo,
				"t": mcwig.CmdChangeBefore,
				"$": mcwig.CmdChangeEndOfLine,
			},
			"d": mcwig.KeyMap{
				"d": mcwig.CmdDeleteLine,
				"w": mcwig.CmdDeleteWord,
				"f": mcwig.CmdDeleteTo,
				"t": mcwig.CmdDeleteBefore,
			},
			"y": mcwig.KeyMap{
				"y": mcwig.CmdYank,
			},
			"g": mcwig.KeyMap{
				"g": mcwig.CmdGotoLine0,
			},
			"ctrl+c": mcwig.KeyMap{
				"ctrl+c": commands.CmdExecute,
				"ctrl+x": mcwig.CmdExit,
			},
			"ctrl+w": mcwig.KeyMap{
				"v":      mcwig.CmdWindowVSplit,
				"w":      mcwig.CmdWindowNext,
				"q":      mcwig.CmdWindowClose,
				"ctrl+w": mcwig.CmdWindowNext,
				"t":      mcwig.CmdWindowToggleLayout,
			},
			"Space": mcwig.KeyMap{
				"b": mcwig.KeyMap{
					"b": commands.CmdBufferPicker,
					"k": mcwig.CmdKillBuffer,
				},
				"f": commands.CmdFindProjectFilePicker,
				"F": commands.CmdCurrentBufferDirFilePicker,
			},
		},
		mcwig.MODE_VISUAL: mcwig.KeyMap{
			"ctrl+e": mcwig.WithSelection(mcwig.CmdScrollDown),
			"ctrl+y": mcwig.WithSelection(mcwig.CmdScrollUp),
			"w":      mcwig.WithSelection(mcwig.CmdForwardWord),
			"b":      mcwig.WithSelection(mcwig.CmdBackwardWord),
			"h":      mcwig.WithSelection(mcwig.CmdCursorLeft),
			"l":      mcwig.WithSelection(mcwig.CmdCursorRight),
			"j":      mcwig.WithSelection(mcwig.CmdCursorLineDown),
			"k":      mcwig.WithSelection(mcwig.CmdCursorLineUp),
			"f":      mcwig.WithSelectionToChar(mcwig.CmdForwardToChar),
			"t":      mcwig.WithSelectionToChar(mcwig.CmdForwardBeforeChar),
			"$":      mcwig.WithSelection(mcwig.CmdGotoLineEnd),
			"0":      mcwig.WithSelection(mcwig.CmdCursorBeginningOfTheLine),
			"x":      mcwig.CmdSelectinDelete,
			"d":      mcwig.CmdSelectinDelete,
			"y":      mcwig.CmdYank,
			"c":      mcwig.CmdSelectionChange,
			"Esc":    mcwig.CmdNormalMode,
			"g": mcwig.KeyMap{
				"g": mcwig.WithSelection(mcwig.CmdGotoLine0),
			},
		},
		mcwig.MODE_VISUAL_LINE: mcwig.KeyMap{
			"j":   mcwig.WithSelection(mcwig.CmdCursorLineDown),
			"k":   mcwig.WithSelection(mcwig.CmdCursorLineUp),
			"h":   mcwig.CmdCursorLeft,
			"l":   mcwig.CmdCursorRight,
			"Esc": mcwig.CmdNormalMode,
			"x":   mcwig.CmdSelectinDelete,
			"d":   mcwig.CmdSelectinDelete,
			"y":   mcwig.CmdYank,
		},
		mcwig.MODE_INSERT: mcwig.KeyMap{
			"Esc":    mcwig.CmdNormalMode,
			"ctrl+f": mcwig.CmdCursorRight,
			"ctrl+b": mcwig.CmdCursorLeft,
			"ctrl+j": mcwig.CmdCursorLineDown,
			"ctrl+k": mcwig.CmdCursorLineUp,
		},
	}
}
