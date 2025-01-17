package mcwig

import (
	"path/filepath"

	"go.lsp.dev/jsonrpc2"
)

type LspServerConfig struct {
	Cmd  []string // gopls serve -listen=127.0.0.1:101010
	Type string   // unix or tcp
	Addr string
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
	conns  map[string]lspConn
	ignore map[string]bool
}

func NewLspManager(e *Editor) *LspManager {
	return &LspManager{
		e:     e,
		conns: map[string]lspConn{},
	}
}

func (l *LspManager) DidOpen(buf *Buffer) {
	root, _ := l.e.Projects.FindRoot(buf)

	_, ignore := l.conns[root]
	if ignore {
		return
	}

	var client lspConn
	var err error

	client, ok := l.conns[root]
	// initialize

	if !ok {
		conf, ok := LspConfigByFileName(buf.FilePath)
		if !ok {
			l.ignore[root] = true
		}

		client, err = l.startServer(conf)
		if err != nil {
			l.e.LogMessage("failed to start tcp server")
			l.e.EchoMessage("failed to start tcp server")
			return
		}

		l.conns[root] = client
	}

	client.didOpen(buf)
}

func (l *LspManager) startServer(sconf LspServerConfig) (lspConn, error) {
	return lspConn{}, nil
}

///////////////////////////////////
// Lsp Procotol Connection
////////////////////////////////////

type lspConn struct {
	rpc jsonrpc2.Conn
}

func (c *lspConn) didOpen(buf *Buffer) {
}
