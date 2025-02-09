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
	e           *Editor
	conns       map[string]*lspConn
	ignore      map[string]bool
	diagnostics map[string]map[uint32][]protocol.Diagnostic
}

func NewLspManager(e *Editor) *LspManager {
	r := &LspManager{
		e:           e,
		conns:       map[string]*lspConn{},
		ignore:      map[string]bool{},
		diagnostics: map[string]map[uint32][]protocol.Diagnostic{},
	}

	go func() {
		events := e.Events.Subscribe()
		for {
			select {
			case msg := <-events:
				switch event := msg.(type) {
				case EventTextChange:
					fmt.Println("change:", event.Start, event.End, "text", event.Text)
					r.DidChange(event)
				}

			}
		}
	}()

	return r
}

// TODO: check for supported file extension before doing anything
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
			return
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

	req := protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(fmt.Sprintf("file://%s", event.Buf.FilePath)),
			},
			Version: int32(time.Now().Unix()),
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(event.Start.Line), Character: uint32(event.Start.Char)},
					End:   protocol.Position{Line: uint32(event.End.Line), Character: uint32(event.End.Char)},
				},
				Text: event.Text,
			},
		},
	}

	_, err := client.rpcConn.Call(context.Background(), protocol.MethodTextDocumentDidChange, req, nil)
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

	_, err := client.rpcConn.Call(context.Background(), protocol.MethodTextDocumentDidClose, req, nil)
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

func (l *LspManager) Completion(buf *Buffer) (sign string) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.ignore[root]
	if ignore {
		return
	}

	client, ok := l.conns[root]
	if !ok {
		return
	}

	req := protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.URI(fmt.Sprintf("file://%s", buf.FilePath)),
			},
			Position: protocol.Position{
				Line:      uint32(buf.Cursor.Line),
				Character: uint32(buf.Cursor.Char),
			},
		},
		Context: &protocol.CompletionContext{
			TriggerKind: protocol.CompletionTriggerKindInvoked,
		},
	}

	var result protocol.CompletionList
	_, err := client.rpcConn.Call(context.Background(), protocol.MethodTextDocumentCompletion, req, &result)
	if err != nil {
		l.e.LogError(err)
	}

	return ""
}

func (m *LspManager) Diagnostics(buf *Buffer, lineNum int) []protocol.Diagnostic {
	if val, ok := m.diagnostics[buf.FilePath]; ok {
		return val[uint32(lineNum)]
	}
	return nil
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
			l.e.LogMessage("lsp:" + scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(perr)
		for scanner.Scan() {
			l.e.LogMessage("lsp:" + scanner.Text())
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
		// data := map[string]interface{}{}
		// json.Unmarshal(req.Params(), &data)
		if req.Method() == "textDocument/publishDiagnostics" {
			rest := protocol.PublishDiagnosticsParams{}
			json.Unmarshal(req.Params(), &rest)

			filepath := rest.URI.Filename()
			l.diagnostics[filepath] = map[uint32][]protocol.Diagnostic{}

			for _, r := range rest.Diagnostics {
				l.diagnostics[filepath][r.Range.Start.Line] = append(l.diagnostics[filepath][r.Range.Start.Line], r)
			}
		}

		l.e.Redraw()

		return reply(ctx, nil, nil)
	}
	c.Go(context.Background(), handler)

	// initialize connection sequence
	r := &protocol.InitializeParams{}
	json.Unmarshal([]byte(lspServerInitJson), r)

	var result protocol.InitializeResult
	_, err = c.Call(context.Background(), "initialize", r, &result)
	if err != nil {
		l.e.LogError(err)
	}
	// fmt.Printf("%+v", result)

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
