package config

import (
	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/commands"
	"github.com/firstrow/mcwig/ui"
)

func DefaultKeyMap() mcwig.ModeKeyMap {
	return mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			// personal config
			"F2": commands.CmdFormatBufferAndSave,
			"F5": commands.CmdMakeRun,

			"ctrl+e": mcwig.CmdScrollDown,
			"ctrl+y": mcwig.CmdScrollUp,
			"h":      mcwig.CmdCursorLeft,
			"l":      mcwig.CmdCursorRight,
			"j":      mcwig.CmdCursorLineDown,
			"k":      mcwig.CmdCursorLineUp,
			"i":      mcwig.CmdEnterInsertMode,
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
			"u":      mcwig.CmdUndo,
			"ctrl+r": mcwig.CmdRedo,
			":":      ui.CmdLineInit,
			"/":      ui.CmdSearchPromptInit,
			";":      commands.CmdBufferPicker,
			"*":      commands.CmdSearchWordUnderCursor,
			"q":      mcwig.CmdMacroRecord,
			"@":      mcwig.CmdMacroPlay,
			"c": mcwig.KeyMap{
				"$": mcwig.CmdChangeEndOfLine,
				"c": mcwig.CmdChangeLine,
				"w": mcwig.CmdChangeWord,
				"a": mcwig.KeyMap{
					"w": mcwig.CmdChangeWORD,
				},
				"i": mcwig.CmdChangeInsideBlock,
				"f": mcwig.CmdChangeTo,
				"t": mcwig.CmdChangeBefore,
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
				"d": commands.CmdGotoDefinition,
				"O": commands.CmdGotoDefinitionOtherWindow,
				"o": commands.CmdViewDefinitionOtherWindow,
				"c": mcwig.CmdToggleComment,
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
			"]": mcwig.KeyMap{
				"]": mcwig.CmdJumpForward,
			},
			"[": mcwig.KeyMap{
				"[": mcwig.CmdJumpBack,
			},
			"Space": mcwig.KeyMap{
				"/": commands.CmdSearchProject,
				"?": commands.CmdCommandPalettePicker,
				"`": mcwig.CmdBufferCycle,
				"*": commands.CmdProjectSearchWordUnderCursor,
				"h": commands.CmdLspHover,
				"e": commands.CmdLspShowDiagnostics,
				"b": mcwig.KeyMap{
					"b": commands.CmdBufferPicker,
					"k": mcwig.CmdKillBuffer,
				},
				"f": commands.CmdFindProjectFilePicker,
				"F": commands.CmdCurrentBufferDirFilePicker,
				"s": mcwig.KeyMap{
					"s": commands.CmdSearchLine,
				},
				"t": commands.CmdThemeSelect,
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
			"$":      mcwig.WithSelection(mcwig.CmdGotoLineEnd),
			"0":      mcwig.WithSelection(mcwig.CmdCursorBeginningOfTheLine),
			"f":      mcwig.CmdForwardToChar,
			"t":      mcwig.CmdForwardBeforeChar,
			"x":      mcwig.CmdSelectionDelete,
			"d":      mcwig.CmdSelectionDelete,
			"y":      mcwig.CmdYank,
			"c":      mcwig.CmdSelectionChange,
			"Esc":    mcwig.CmdNormalMode,
			"*":      commands.CmdSearchWordUnderCursor,
			"g": mcwig.KeyMap{
				"g": mcwig.WithSelection(mcwig.CmdGotoLine0),
				"c": mcwig.CmdToggleComment,
			},
		},
		mcwig.MODE_VISUAL_LINE: mcwig.KeyMap{
			"j":   mcwig.WithSelection(mcwig.CmdCursorLineDown),
			"k":   mcwig.WithSelection(mcwig.CmdCursorLineUp),
			"h":   mcwig.CmdCursorLeft,
			"l":   mcwig.CmdCursorRight,
			"Esc": mcwig.CmdNormalMode,
			"x":   mcwig.CmdSelectionDelete,
			"d":   mcwig.CmdSelectionDelete,
			"y":   mcwig.CmdYank,
			"g": mcwig.KeyMap{
				"g": mcwig.WithSelection(mcwig.CmdGotoLine0),
				"c": mcwig.CmdToggleComment,
			},
		},
		mcwig.MODE_INSERT: mcwig.KeyMap{
			"Esc":    mcwig.CmdExitInsertMode,
			"ctrl+f": mcwig.CmdCursorRight,
			"ctrl+b": mcwig.CmdCursorLeft,
			"ctrl+j": mcwig.CmdCursorLineDown,
			"ctrl+k": mcwig.CmdCursorLineUp,
		},
	}
}

