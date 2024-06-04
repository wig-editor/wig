package pipe

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/google/shlex"
)

type header struct {
	cmd         string
	interactive bool
	append      bool
}

type pipeDrv struct {
	e *mcwig.Editor
	// TODO: store cmds per-command. so it will be possible to keep many long running commands in same buffer
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	outBuf *mcwig.Buffer
}

func New(e *mcwig.Editor) *pipeDrv {
	return &pipeDrv{
		e: e,
	}
}

func (p *pipeDrv) parseHeader(buf *mcwig.Buffer) header {
	input := ""

	currentLine := buf.Lines.First()
	for currentLine != nil {
		if len(currentLine.Value) == 0 {
			break
		}
		if currentLine.Value[0] != '#' {
			break
		}

		input += string(currentLine.Value) + "\n"
		currentLine = currentLine.Next()
	}

	lines := strings.Split(input, "\n")
	result := make(map[string]string, 10)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			parts := strings.SplitN(strings.TrimPrefix(line, "#"), ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				result[key] = value
			}
		}
	}

	h := header{}
	if val, ok := result["cmd"]; ok {
		h.cmd = val
	}
	if val, ok := result["interactive"]; ok {
		h.interactive = isTrue(val)
	}
	if val, ok := result["append"]; ok {
		h.append = isTrue(val)
	}

	return h
}

func (p *pipeDrv) getCommand(cmd string) string {
	s := strings.Split(cmd, " ")
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

func (p *pipeDrv) buildArgs(cmd string, input string) []string {
	if strings.Contains(cmd, "%s") {
		cmd = strings.Replace(cmd, "%s", input, 1)
	} else {
		cmd = fmt.Sprintf("%s %s", cmd, input)
	}

	args, _ := shlex.Split(cmd)
	if len(args) > 0 {
		return args[1:]
	}

	return []string{}
}

func (p *pipeDrv) outBufferFor(buf *mcwig.Buffer) *mcwig.Buffer {
	if p.outBuf != nil {
		return p.outBuf
	}
	p.outBuf = p.e.BufferGetByName(fmt.Sprintf("[output] %s", buf.GetName()))
	return p.outBuf
}

func (p *pipeDrv) send(opts header, outBuf *mcwig.Buffer, input string) {
	if p.cmd != nil && opts.interactive {
		io.WriteString(p.stdin, input+"\n")
		return
	}

	if opts.interactive {
		p.cmd = exec.Command(p.getCommand(opts.cmd), p.buildArgs(opts.cmd, "")...)
		pin, _ := p.cmd.StdinPipe()
		p.stdin = pin
		io.WriteString(p.stdin, input+"\n")
	} else {
		p.cmd = exec.Command(p.getCommand(opts.cmd), p.buildArgs(opts.cmd, input)...)
	}

	pout, _ := p.cmd.StdoutPipe()
	perr, _ := p.cmd.StderrPipe()

	err := p.cmd.Start()
	if err != nil {
		outBuf.AppendStringLine(err.Error())
		p.e.Redraw()
	}

	go func() {
		scanner := bufio.NewScanner(pout)
		for scanner.Scan() {
			outBuf.AppendStringLine(scanner.Text())
			p.e.Redraw()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(perr)
		for scanner.Scan() {
			outBuf.AppendStringLine(scanner.Text())
			p.e.Redraw()
		}
	}()

	go func() {
		err := p.cmd.Wait()
		if err != nil {
			outBuf.AppendStringLine(err.Error())
			p.e.Redraw()
		}
		p.cmd = nil
		p.stdin = nil
	}()
}

func (p *pipeDrv) Exec(e *mcwig.Editor, buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
	outBuf := p.outBufferFor(buf)
	opts := p.parseHeader(buf)
	p.send(opts, outBuf, string(line.Value))
	p.e.EnsureBufferIsVisible(outBuf)
}

func (p *pipeDrv) ExecBuffer() {

}

func isTrue(val string) bool {
	var boolValues = []string{
		"1",
		"t",
		"T",
		"TRUE",
		"true",
		"True",
	}

	for _, row := range boolValues {
		if val == row {
			return true
		}
	}

	return false
}
