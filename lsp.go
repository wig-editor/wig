package wig

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type lspConn struct {
	rpcConn jsonrpc2.Conn
}

var lspConfigs = map[string]LanguageServerConfig{}

func init() {
	lcfg := LoadLanguagesConfig()
	for _, v := range lcfg.Languages {
		exts, _ := v.GetFileTypes()
		servers := v.GetLanguageServers()
		if len(servers) == 0 {
			continue
		}
		c := lcfg.LanguageServers[servers[0]]
		for _, ext := range exts {
			ekey := "." + ext
			lspConfigs[ekey] = c
		}
	}
}

func LspConfigByFileName(file string) (conf LanguageServerConfig, found bool) {
	fp := filepath.Ext(file)
	conf, found = lspConfigs[fp]
	return
}

type LspManager struct {
	rw          sync.Mutex
	e           *Editor
	conns       map[string]*lspConn
	ignore      map[string]bool
	diagnostics map[string]map[uint32][]protocol.Diagnostic
	ctxCancel   context.CancelFunc
	timeout     time.Duration
}

func NewLspManager(e *Editor) *LspManager {
	r := &LspManager{
		e:           e,
		conns:       map[string]*lspConn{},
		ignore:      map[string]bool{},
		diagnostics: map[string]map[uint32][]protocol.Diagnostic{},
		timeout:     500 * time.Millisecond,
	}

	go func() {
		for event := range e.Events.Subscribe() {
			switch e := event.Msg.(type) {
			case EventTextChange:
				r.DidChange(e)
			case EventKeyPressed:
				if r.ctxCancel != nil {
					r.ctxCancel()
				}
			}
			event.Wg.Done()
		}
	}()

	return r
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
		go func() {
			conf, ok := LspConfigByFileName(buf.FilePath)
			if !ok {
				l.ignore[root] = true
				return
			}

			// starts server and returns client conn
			client, err = l.startAndInitializeServer(conf, buf)
			if err != nil {
				l.e.LogMessage("failed to start tcp server")
				l.e.EchoMessage("failed to start tcp server")
				return
			}

			l.conns[root] = client
			client.didOpen(buf)
		}()

		return
	}

	client.didOpen(buf)
}

func (l *lspConn) didOpen(buf *Buffer) {
	didOpen := protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(fmt.Sprintf("file://%s", buf.FilePath)),
			LanguageID: protocol.GoLanguage,
			Version:    0,
			Text:       buf.String(),
		},
	}
	err := l.rpcConn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
	if err != nil {
		EditorInst.LogError(err)
	}
}

func (l *LspManager) DidChange(event EventTextChange) {
	root, _ := l.e.Projects.FindRoot(event.Buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	rrange := &protocol.Range{
		Start: protocol.Position{Line: uint32(event.Start.Line), Character: uint32(event.Start.Char)},
		End:   protocol.Position{Line: uint32(event.End.Line), Character: uint32(event.End.Char)},
	}

	req := protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(fmt.Sprintf("file://%s", event.Buf.FilePath)),
			},
			Version: int32(time.Now().Unix()),
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{
				Range: rrange,
				Text:  event.Text,
			},
		},
	}

	err := client.rpcConn.Notify(context.Background(), protocol.MethodTextDocumentDidChange, req)
	if err != nil {
		l.e.LogError(err)
	}
}

func (l *LspManager) DidClose(buf *Buffer) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	req := protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
		},
	}

	err := client.rpcConn.Notify(context.Background(), protocol.MethodTextDocumentDidClose, req)
	if err != nil {
		l.e.LogError(err)
	}
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

	cur := CursorGet(EditorInst, buf)
	srReq := protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(cur.Line),
				Character: uint32(cur.Char),
			},
		},
	}
	var sigHelpResp protocol.SignatureHelp

	ctx, cancel := context.WithTimeout(context.TODO(), l.timeout)
	l.ctxCancel = cancel
	defer cancel()

	_, err := client.rpcConn.Call(ctx, protocol.MethodTextDocumentSignatureHelp, srReq, &sigHelpResp)
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

	cur := CursorGet(EditorInst, buf)
	srReq := protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(cur.Line),
				Character: uint32(cur.Char),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), l.timeout)
	l.ctxCancel = cancel
	defer cancel()

	var hover protocol.Hover
	_, err := client.rpcConn.Call(ctx, protocol.MethodTextDocumentHover, srReq, &hover)
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

	currentCur := CursorGet(EditorInst, buf)
	pos := protocol.Position{
		Line:      uint32(currentCur.Line),
		Character: uint32(currentCur.Char),
	}

	definitionReq := protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: pos,
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), l.timeout)
	l.ctxCancel = cancel
	defer cancel()

	var definitionResp []protocol.Location
	_, err := client.rpcConn.Call(ctx, protocol.MethodTextDocumentDefinition, definitionReq, &definitionResp)

	if err != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), l.timeout)
		l.ctxCancel = cancel
		defer cancel()

		var definitionResp2 protocol.Location
		_, err2 := client.rpcConn.Call(ctx, protocol.MethodTextDocumentDefinition, definitionReq, &definitionResp2)
		if err2 != nil {
			l.e.EchoMessage(err2.Error())
		} else {
			definitionResp = append(definitionResp, definitionResp2)
		}
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

func (l *LspManager) Completion(buf *Buffer) (res CompletionItems) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	cur := CursorGet(EditorInst, buf)
	req := protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(cur.Line),
				Character: uint32(cur.Char),
			},
		},
		Context: &protocol.CompletionContext{
			// TriggerCharacter: ".",
			// TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
			TriggerKind: protocol.CompletionTriggerKindInvoked,
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), l.timeout)
	l.ctxCancel = cancel
	defer cancel()
	_, err := client.rpcConn.Call(ctx, protocol.MethodTextDocumentCompletion, req, &res)
	if err != nil {
		l.e.LogMessage("lsp completion error:", err.Error())
	}

	res2 := map[string]any{}
	_, err = client.rpcConn.Call(ctx, protocol.MethodTextDocumentCompletion, req, &res2)
	if err != nil {
		l.e.LogMessage("lsp completion error:", err.Error())
	}
	fmt.Println(res2)

	return
}

// TODO: return copy of diagnostics
func (m *LspManager) Diagnostics(buf *Buffer, lineNum int) []protocol.Diagnostic {
	m.rw.Lock()
	defer m.rw.Unlock()

	if val, ok := m.diagnostics[buf.FilePath]; ok {
		return val[uint32(lineNum)]
	}

	return nil
}

type pipeWrapper struct {
	reader io.Reader
	writer io.Writer
	closer io.Closer
}

func (pw *pipeWrapper) Read(p []byte) (n int, err error) {
	return pw.reader.Read(p)
}

func (pw *pipeWrapper) Write(p []byte) (n int, err error) {
	return pw.writer.Write(p)
}

func (pw *pipeWrapper) Close() error {
	return nil
}

func (l *LspManager) startAndInitializeServer(conf LanguageServerConfig, buf *Buffer) (conn *lspConn, err error) {
	cmd := exec.Command(conf.Command, conf.Args...)

	pin, _ := cmd.StdinPipe()
	pout, _ := cmd.StdoutPipe()
	// perr, _ := cmd.StderrPipe()

	err = cmd.Start()
	if err != nil {
		l.e.LogError(err)
	}

	go func() {
		cmd.Wait()
		l.e.LogMessage("lsp server exited")
		// cleanup all connections
		l.conns = make(map[string]*lspConn)
	}()

	st := &pipeWrapper{
		reader: pout,
		writer: pin,
	}

	s := jsonrpc2.NewStream(st)
	c := jsonrpc2.NewConn(s)

	handler := func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		// fmt.Println("got LSP hanlder req", req.Method(), string(req.Params()))
		if req.Method() == "workspace/configuration" {
			// TODO: move out to language config
			var resp []any
			if conf.Command == "gopls" {
				resp = []any{
					map[string]any{
						"analysisProgressReporting": true,
						"buildFlags":                []any{},
						"codelenses": map[string]any{
							"gc_details":         false,
							"generate":           true,
							"regenerate_cgo":     true,
							"tidy":               true,
							"upgrade_dependency": true,
							"test":               true,
							"vendor":             true,
						},
						"completeFunctionCalls": true,
						"completionBudget":      "100ms",
						"diagnosticsDelay":      "1s",
						"directoryFilters":      []any{},
						"gofumpt":               false,
						"hoverKind":             "SynopsisDocumentation",
						"importShortcut":        "Both",
						"linkTarget":            "pkg.go.dev",
						"linksInHover":          true,
						"local":                 "",
						"matcher":               "Fuzzy",
						"standaloneTags": []any{
							"ignore",
						},
						"symbolMatcher":   "FastFuzzy",
						"symbolScope":     "all",
						"symbolStyle":     "Dynamic",
						"usePlaceholders": true,
						"verboseOutput":   true,
					},
				}
			}

			return reply(ctx, resp, nil)
		}

		if req.Method() == "textDocument/publishDiagnostics" {
			rest := protocol.PublishDiagnosticsParams{}
			json.Unmarshal(req.Params(), &rest)

			filepath := rest.URI.Filename()

			l.rw.Lock()
			l.diagnostics[filepath] = map[uint32][]protocol.Diagnostic{}
			for _, r := range rest.Diagnostics {
				l.diagnostics[filepath][r.Range.Start.Line] = append(l.diagnostics[filepath][r.Range.Start.Line], r)
			}
			l.rw.Unlock()

			// TODO: redraw only if modified buffer is visible
			// TODO: schedule redraw. e.g. one redraw per 3ms.
			l.e.Redraw()
		}

		return reply(ctx, nil, nil)
	}

	c.Go(context.Background(), handler)

	// initialize connection sequence
	r := &protocol.InitializeParams{}
	json.Unmarshal([]byte(lspServerInitJson), r)
	fileRoot, _ := l.e.Projects.FindRoot(buf)
	r.RootURI = protocol.DocumentURI(fmt.Sprintf("file://%s", fileRoot))
	r.WorkspaceFolders = append(r.WorkspaceFolders, protocol.WorkspaceFolder{
		Name: fileRoot,
		URI:  fmt.Sprintf("file://%s", fileRoot),
	})

	var result protocol.InitializeResult
	_, err = c.Call(context.Background(), protocol.MethodInitialize, r, &result)
	if err != nil {
		l.e.LogError(err)
	}

	err = c.Notify(context.Background(), protocol.MethodInitialized, protocol.InitializedParams{})
	if err != nil {
		l.e.LogError(err)
	}

	return &lspConn{
		rpcConn: c,
	}, nil
}

