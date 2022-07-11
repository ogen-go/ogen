package json

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
)

type padLine struct {
	pad, line int
}

func log10(val int) (r int) {
	for val >= 10 {
		r++
		val = val / 10
	}
	return r
}

func (p padLine) Format(f fmt.State, verb rune) {
	padding := p.pad - log10(p.line)
	var buf [32]byte
	for i := 0; i < padding; i++ {
		buf[i] = ' '
	}
	b := strconv.AppendInt(buf[:padding], int64(p.line), 10)
	_, _ = f.Write(b)
}

func (l Lines) PrettyError(w io.Writer, filename, msg string, loc Location, context int) error {
	tw := w
	const (
		leftPad           = "  "
		verticalBorder    = "|"
		verticalPointer   = "\u2191"
		horizontalPointer = "\u2192"
	)

	// Line starts from 1, but index starts from 0.
	errLine := int(loc.Line) - 1
	if len(l.data) == 0 || errLine <= 0 || errLine >= len(l.lines) {
		return errors.New("line number is out of range")
	}
	plainColor := color.New(color.Reset)
	errColor := color.New(color.FgRed)

	if _, err := errColor.Fprintf(w, "%s- %s -> %s\n",
		leftPad,
		loc.WithFilename(filename),
		msg,
	); err != nil {
		return err
	}

	// Round up to the nearest odd number.
	if context%2 == 0 {
		context++
	}
	topContext, bottomContext := errLine-context/2, errLine+context/2

	padNum := 2
	if l := log10(bottomContext); l > 2 {
		padNum = l
	}

	line := func(n int) []byte {
		start, end := l.Line(n)
		if start < 0 || end < 0 {
			return []byte(`Cannot render line properly, please fill a bug report`)
		}
		return bytes.Trim(l.data[start:end], "\r\n")
	}
	printLine := func(leftPad string, n int, c *color.Color) error {
		lineNumber := padLine{
			pad: padNum,
			// Line number is 1-based.
			line: n + 1,
		}
		// Line number is 1-based, but index is 0-based.
		lineText := line(n + 1)
		_, err := c.Fprintf(tw, "\t%s%d %s %s\t\n", leftPad, lineNumber, verticalBorder, lineText)
		return err
	}

	// Print top context.
	for contextLine := topContext; contextLine < errLine; contextLine++ {
		if contextLine < 0 || contextLine >= len(l.lines) {
			continue
		}
		if err := printLine(leftPad, contextLine, plainColor); err != nil {
			return err
		}
	}

	// Print error line.
	if err := printLine(horizontalPointer+" ", errLine, errColor); err != nil {
		return err
	}

	// Print column pointer.
	if loc.Column > 0 {
		if _, err := errColor.Fprintf(tw,
			"\t%s%s %s %s%s\t\n",
			leftPad,
			strings.Repeat(" ", padNum+1),
			verticalBorder,
			strings.Repeat(" ", int(loc.Column-1)),
			verticalPointer,
		); err != nil {
			return err
		}
	}

	// Print bottom context.
	for contextLine := errLine + 1; contextLine <= bottomContext; contextLine++ {
		if contextLine < 0 || contextLine >= len(l.lines) {
			continue
		}
		if err := printLine(leftPad, contextLine, plainColor); err != nil {
			return err
		}
	}

	return nil
}
