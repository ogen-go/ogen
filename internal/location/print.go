package location

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
		val /= 10
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

// ColorFunc defines a simple printer callback.
type ColorFunc func(w io.Writer, s string, args ...any) (int, error)

// PrintListingOptions is a set of options for PrintListing.
type PrintListingOptions struct {
	// Context is the number of lines to print before and after the error line.
	//
	// If is zero, the default value 5 is used.
	Context int
	// If is nil, the default value color.New(color.FgRed) is used.
	ErrColor ColorFunc
	// If is nil, the default value color.New(color.Reset) is used.
	PlainColor ColorFunc
}

// WithoutColor creates a copy of the options with disabled color.
func (o PrintListingOptions) WithoutColor() PrintListingOptions {
	o.ErrColor = fmt.Fprintf
	o.PlainColor = fmt.Fprintf
	return o
}

func (o PrintListingOptions) contextLines(errLine int) (padNum, top, bottom int) {
	context := o.Context

	// Round up to the nearest odd number.
	if context%2 == 0 {
		context++
	}
	top, bottom = errLine-context/2, errLine+context/2

	padNum = 2
	if l := log10(bottom); l > 2 {
		padNum = l
	}

	return padNum, top, bottom
}

func (o *PrintListingOptions) setDefaults() {
	if o.Context == 0 {
		o.Context = 5
	}
	if o.ErrColor == nil {
		o.ErrColor = color.New(color.FgRed).Fprintf
	}
	if o.PlainColor == nil {
		o.PlainColor = color.New(color.Reset).Fprintf
	}
}

// BugLine is a fallback line when the line is not available.
const BugLine = `Cannot render line properly, please fill a bug report`

// PrintListing prints given message with line number and file listing to the writer.
//
// The context parameter defines the number of lines to print before and after.
func (f File) PrintListing(w io.Writer, msg string, loc Position, opts PrintListingOptions) error {
	opts.setDefaults()
	l := f.Lines

	// Line starts from 1, but index starts from 0.
	errLine := loc.Line - 1
	if len(l.data) == 0 || errLine < 0 || errLine >= len(l.lines) {
		return errors.New("line number is out of range")
	}

	const (
		leftPad           = "  "
		verticalBorder    = "|"
		verticalPointer   = "\u2191"
		horizontalPointer = "\u2192"
	)
	var (
		filename = f.humanName()

		plainColor          = opts.PlainColor
		errColor            = opts.ErrColor
		padNum, top, bottom = opts.contextLines(loc.Line)
	)

	var formattedMsg string
	if msg != "" {
		formattedMsg = " -> " + msg
	}

	if _, err := errColor(w, "%s- %s%s\n",
		leftPad,
		loc.WithFilename(filename),
		formattedMsg,
	); err != nil {
		return err
	}

	line := func(n int) []byte {
		start, end := l.Line(n)
		if start < 0 || end < 0 {
			return []byte(BugLine)
		}
		return bytes.Trim(l.data[start:end], "\r\n")
	}
	printLine := func(leftPad string, n int, colored ColorFunc) error {
		lineNumber := padLine{
			pad: padNum,
			// Line number is 1-based.
			line: n + 1,
		}
		// Line number is 1-based, but index is 0-based.
		lineText := line(n + 1)
		_, err := colored(w, "\t%s%d %s %s\t\n", leftPad, lineNumber, verticalBorder, lineText)
		return err
	}

	// Print top context.
	for contextLine := top; contextLine < errLine; contextLine++ {
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
		if _, err := errColor(w,
			"\t%s%s %s %s%s\t\n",
			leftPad,
			strings.Repeat(" ", padNum+1),
			verticalBorder,
			strings.Repeat(" ", loc.Column-1),
			verticalPointer,
		); err != nil {
			return err
		}
	}

	// Print bottom context.
	for contextLine := errLine + 1; contextLine <= bottom; contextLine++ {
		if contextLine < 0 || contextLine >= len(l.lines) {
			continue
		}
		if err := printLine(leftPad, contextLine, plainColor); err != nil {
			return err
		}
	}

	return nil
}
