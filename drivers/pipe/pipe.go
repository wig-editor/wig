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

type Options struct {
	IsPrompt bool
	Cmd      []string
}

type pipeDrv struct {
	opts  Options
	stdin io.WriteCloser
	cmd   *exec.Cmd
	e     *mcwig.Editor
}

func New(e *mcwig.Editor, opts Options) *pipeDrv {
	return &pipeDrv{
		e:    e,
		opts: opts,
	}
}

func (p *pipeDrv) getCommand() string {
	header := string(p.e.ActiveBuffer().Lines.First().Value)
	s := strings.Split(header, " ")
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

func (p *pipeDrv) buildArgs(input string) []string {
	header := string(p.e.ActiveBuffer().Lines.First().Value)

	if strings.Contains(header, "%s") {
		header = strings.Replace(header, "%s", input, 1)
	} else {
		header = fmt.Sprintf("%s %s", header, input)
	}

	args, _ := shlex.Split(header)

	if len(args) > 0 {
		return args[1:]
	}

	return []string{}
}

func (p *pipeDrv) send(input string) {
	if p.cmd != nil && p.opts.IsPrompt {
		io.WriteString(p.stdin, input+"\n")
		return
	}

	if p.opts.IsPrompt {
		p.cmd = exec.Command(p.getCommand(), p.buildArgs("")...)
		pin, _ := p.cmd.StdinPipe()
		p.stdin = pin
		io.WriteString(p.stdin, input+"\n")
	} else {
		p.cmd = exec.Command(p.getCommand(), p.buildArgs(input)...)
	}

	pout, _ := p.cmd.StdoutPipe()
	perr, _ := p.cmd.StderrPipe()

	err := p.cmd.Start()
	if err != nil {
		p.e.BufferGetByName("[output]").AppendStringLine(err.Error())
		p.e.Redraw()
	}

	go func() {
		scanner := bufio.NewScanner(pout)
		for scanner.Scan() {
			p.e.BufferGetByName("[output]").AppendStringLine(scanner.Text())
			p.e.Redraw()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(perr)
		for scanner.Scan() {
			p.e.BufferGetByName("[output]").AppendStringLine(scanner.Text())
			p.e.Redraw()
		}
	}()

	go func() {
		err := p.cmd.Wait()
		if err != nil {
			p.e.BufferGetByName("[output]").AppendStringLine(err.Error())
			p.e.Redraw()
		}
		p.cmd = nil
		p.stdin = nil
	}()
}

func (p *pipeDrv) Exec(e *mcwig.Editor, buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
	p.send(string(line.Value))
}

func (p *pipeDrv) ExecBuffer() {

}
