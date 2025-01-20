package mcwig

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var jjj = `{
  "processId": null,
  "rootPath": "/home/andrew/code/mcwig",
  "clientInfo": {
    "name": "mcwig",
    "version": "001"
  },
  "rootUri": "file:///home/andrew/code/mcwig",
  "capabilities": {
    "general": {
      "positionEncodings": [
        "utf-32",
        "utf-16"
      ]
    },
    "workspace": {
      "workspaceEdit": {
        "documentChanges": true,
        "resourceOperations": [
          "create",
          "rename",
          "delete"
        ]
      },
      "applyEdit": true,
      "symbol": {
        "symbolKind": {
          "valueSet": [
            1,
            2,
            3,
            4,
            5,
            6,
            7,
            8,
            9,
            10,
            11,
            12,
            13,
            14,
            15,
            16,
            17,
            18,
            19,
            20,
            21,
            22,
            23,
            24,
            25,
            26
          ]
        }
      },
      "executeCommand": {
        "dynamicRegistration": false
      },
      "didChangeWatchedFiles": {
        "dynamicRegistration": true
      },
      "workspaceFolders": true,
      "configuration": true,
      "diagnostics": {
        "refreshSupport": false
      },
      "fileOperations": {
        "didCreate": false,
        "willCreate": false,
        "didRename": true,
        "willRename": true,
        "didDelete": false,
        "willDelete": false
      }
    },
    "textDocument": {
      "declaration": {
        "dynamicRegistration": true,
        "linkSupport": true
      },
      "dn347ggVefinition": {
        "dynamicRegistration": true,
        "linkSupport": true
      },
      "references": {
        "dynamicRegistration": true
      },
      "implementation": {
        "dynamicRegistration": true,
        "linkSupport": true
      },
      "typeDefinition": {
        "dynamicRegistration": true,
        "linkSupport": true
      },
      "synchronization": {
        "willSave": true,
        "didSave": true,
        "willSaveWaitUntil": true
      },
      "documentSymbol": {
        "symbolKind": {
          "valueSet": [
            1,
            2,
            3,
            4,
            5,
            6,
            7,
            8,
            9,
            10,
            11,
            12,
            13,
            14,
            15,
            16,
            17,
            18,
            19,
            20,
            21,
            22,
            23,
            24,
            25,
            26
          ]
        },
        "hierarchicalDocumentSymbolSupport": true
      },
      "formatting": {
        "dynamicRegistration": true
      },
      "rangeFormatting": {
        "dynamicRegistration": true
      },
      "onTypeFormatting": {
        "dynamicRegistration": true
      },
      "rename": {
        "dynamicRegistration": true,
        "prepareSupport": true
      },
      "codeAction": {
        "dynamicRegistration": true,
        "isPreferredSupport": true,
        "codeActionLiteralSupport": {
          "codeActionKind": {
            "valueSet": [
              "",
              "quickfix",
              "refactor",
              "refactor.extract",
              "refactor.inline",
              "refactor.rewrite",
              "source",
              "source.organizeImports"
            ]
          }
        },
        "resolveSupport": {
          "properties": [
            "edit",
            "command"
          ]
        },
        "dataSupport": true
      },
      "completion": {
        "completionItem": {
          "snippetSupport": true,
          "documentationFormat": [
            "markdown",
            "plaintext"
          ],
          "resolveAdditionalTextEditsSupport": true,
          "insertReplaceSupport": true,
          "deprecatedSupport": true,
          "resolveSupport": {
            "properties": [
              "documentation",
              "detail",
              "additionalTextEdits",
              "command",
              "insertTextFormat",
              "insertTextMode"
            ]
          },
          "insertTextModeSupport": {
            "valueSet": [
              1,
              2
            ]
          }
        },
        "contextSupport": true,
        "dynamicRegistration": true
      },
      "signatureHelp": {
        "signatureInformation": {
          "parameterInformation": {
            "labelOffsetSupport": true
          }
        },
        "dynamicRegistration": true
      },
      "documentLink": {
        "dynamicRegistration": true,
        "tooltipSupport": true
      },
      "hover": {
        "contentFormat": [
          "markdown",
          "plaintext"
        ],
        "dynamicRegistration": true
      },
      "selectionRange": {
        "dynamicRegistration": true
      },
      "callHierarchy": {
        "dynamicRegistration": false
      },
      "typeHierarchy": {
        "dynamicRegistration": true
      },
      "publishDiagnostics": {
        "relatedInformation": true,
        "tagSupport": {
          "valueSet": [
            1,
            2
          ]
        },
        "versionSupport": true
      },
      "diagnostic": {
        "dynamicRegistration": false,
        "relatedDocumentSupport": false
      },
      "linkedEditingRange": {
        "dynamicRegistration": true
      }
    },
    "window": {
      "workDoneProgress": true,
      "showDocument": {
        "support": true
      }
    }
  },
  "initializationOptions": null,
  "workDoneToken": "1"
}`

type LspServerConfig struct {
	Cmd  []string // gopls serve -listen=127.0.0.1:101010
	Type string   // unix or tcp
	Addr string
}

type lspConn struct {
	rpcConn jsonrpc2.Conn
}

var lspConfigs = map[string]LspServerConfig{
	".go": {
		Cmd:  []string{"gopls", "-listen", "127.0.0.1:12345"}, // TODO: use unix file sockets
		Type: "tcp",
		Addr: "127.0.0.1:12345",
	},
}

func LspConfigByFileName(file string) (conf LspServerConfig, found bool) {
	fp := filepath.Ext(file)
	conf, found = lspConfigs[fp]
	return
}

type LspManager struct {
	e      *Editor
	conns  map[string]*lspConn
	ignore map[string]bool
}

func NewLspManager(e *Editor) *LspManager {
	return &LspManager{
		e:     e,
		conns: map[string]*lspConn{},
	}
}

func (l *LspManager) DidOpen(buf *Buffer) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	var client *lspConn
	var err error

	client, ok := l.conns[root]
	// initialize
	if !ok {
		conf, ok := LspConfigByFileName(buf.FilePath)
		if !ok {
			l.ignore[root] = true
		}

		// starts server and returns client conn
		client, err = l.startAndInitializeServer(conf)
		if err != nil {
			l.e.LogMessage("failed to start tcp server")
			l.e.EchoMessage("failed to start tcp server")
			return
		}

		l.conns[root] = client
	}

	client.didOpen(buf)
}

func (l *LspManager) Signature(buf *Buffer, cursor Cursor) (sign string) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	srReq := protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(buf.Cursor.Line),
				Character: uint32(buf.Cursor.Char),
			},
		},
	}
	var sigHelpResp protocol.SignatureHelp
	_, err := client.rpcConn.Call(context.Background(), protocol.MethodTextDocumentSignatureHelp, srReq, &sigHelpResp)
	if err != nil {
		l.e.LogError(err)
	}

	if len(sigHelpResp.Signatures) > 0 {
		sign = sigHelpResp.Signatures[0].Label
	}

	return
}

func (l *LspManager) Hover(buf *Buffer, cursor Cursor) (sign string) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	srReq := protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(buf.Cursor.Line),
				Character: uint32(buf.Cursor.Char),
			},
		},
	}
	var hover protocol.Hover
	_, err := client.rpcConn.Call(context.Background(), protocol.MethodTextDocumentHover, srReq, &hover)
	if err != nil {
		l.e.LogError(err)
	}

	result := strings.ReplaceAll(hover.Contents.Value, "\n", "")

	return result
}

func (l *LspManager) Definition(buf *Buffer, cursor Cursor) (filePath string, cur Cursor) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	definitionReq := protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(buf.Cursor.Line),
				Character: uint32(buf.Cursor.Char),
			},
		},
	}
	var definitionResp []protocol.Location
	_, err := client.rpcConn.Call(context.Background(), protocol.MethodTextDocumentDefinition, definitionReq, &definitionResp)
	if err != nil {
		l.e.EchoMessage(err.Error())
	}

	if len(definitionResp) == 0 {
		return
	}

	filePath = string(definitionResp[0].URI[7:])
	line := int(definitionResp[0].Range.Start.Line)
	ch := int(definitionResp[0].Range.Start.Character)

	return filePath, Cursor{
		Line: line,
		Char: ch,
	}
}

func (l *LspManager) startAndInitializeServer(conf LspServerConfig) (conn *lspConn, err error) {
	cmd := exec.Command(conf.Cmd[0], conf.Cmd[1:]...)

	pout, _ := cmd.StdoutPipe()
	perr, _ := cmd.StderrPipe()

	err = cmd.Start()
	if err != nil {
		l.e.LogError(err)
	}

	go func() {
		scanner := bufio.NewScanner(pout)
		for scanner.Scan() {
			l.e.LogMessage(scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(perr)
		for scanner.Scan() {
			l.e.LogMessage(scanner.Text())
		}
	}()

	go func() {
		cmd.Wait()
		l.e.LogMessage("lsp server exited")
		// cleanup all connections
		l.conns = make(map[string]*lspConn)
	}()

	// TODO: replace with wait channel
	time.Sleep(100 * time.Millisecond)

	tcpc, err := net.Dial(conf.Type, conf.Addr)
	if err != nil {
		l.e.LogError(err)
	}

	s := jsonrpc2.NewStream(tcpc)
	c := jsonrpc2.NewConn(s)

	handler := func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		return reply(ctx, nil, nil)
	}
	c.Go(context.Background(), handler)

	// initialize connection sequence
	r := &protocol.InitializeParams{}
	json.Unmarshal([]byte(jjj), r)

	var result protocol.InitializeResult
	_, err = c.Call(context.Background(), "initialize", r, &result)
	if err != nil {
		l.e.LogError(err)
	}

	_, err = c.Call(context.Background(), protocol.MethodInitialized, protocol.InitializedParams{}, nil)
	if err != nil {
		l.e.LogError(err)
	}

	return &lspConn{
		rpcConn: c,
	}, nil
}

func (l *lspConn) didOpen(buf *Buffer) {
	// didOpen
	didOpen := protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(fmt.Sprintf("file://%s", buf.FilePath)),
			LanguageID: protocol.GoLanguage,
			Version:    0,
			Text:       buf.String(),
		},
	}
	_, err := l.rpcConn.Call(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen, nil)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Println("DIDOPEN DONE", id, err)
	// didOpen
}
