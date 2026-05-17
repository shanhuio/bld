package jsonx

import (
	"bytes"
	"io"
)

// tabPrinter is a write filter that supports shift tabbing
type tabPrinter struct {
	out     io.Writer
	e       error
	midLine bool

	indent    int
	indentStr string
}

func newTabPrinter(out io.Writer) *tabPrinter {
	return &tabPrinter{
		out:       out,
		indentStr: "    ",
	}
}

func (p *tabPrinter) write(buf []byte) {
	if p.e != nil {
		return
	}
	_, p.e = p.out.Write(buf)
}

func (p *tabPrinter) writeBytes(buf []byte) {
	if len(buf) == 0 {
		return
	}

	if !p.midLine {
		for j := 0; j < p.indent; j++ {
			p.write([]byte(p.indentStr))
		}
	}

	p.midLine = true
	p.write(buf)
}

func (p *tabPrinter) writeEndl() {
	p.write([]byte("\n"))
	p.midLine = false
}

// Write writes the buf. It adds indent before each line.
func (p *tabPrinter) Write(buf []byte) (int, error) {
	lines := bytes.Split(buf, []byte("\n"))

	for i, line := range lines {
		if i > 0 {
			p.writeEndl()
		}

		p.writeBytes(line)
	}

	return len(buf), nil
}

// Tab indents in one level
func (p *tabPrinter) Tab() { p.indent++ }

// ShiftTab indents out one level
func (p *tabPrinter) ShiftTab() {
	if p.indent > 0 {
		p.indent--
	}
}

// Err returns the first error on printing.
func (p *tabPrinter) Err() error { return p.e }
