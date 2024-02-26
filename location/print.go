package location

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"golang.org/x/exp/constraints"

	"github.com/ogen-go/ogen/internal/xslices"
)

// ColorFunc defines a simple printer callback.
type ColorFunc func(w io.Writer, s string, args ...any) (int, error)

// PrintListingOptions is a set of options for PrintListing.
type PrintListingOptions struct {
	// Context is the number of lines to print before and after the error line.
	//
	// If is zero, the default value 5 is used.
	Context int
	// MsgColor sets message color.
	MsgColor ColorFunc
	// TextColor sets text color.
	PlainColor ColorFunc
}

// WithoutColor creates a copy of the options with disabled color.
func (o PrintListingOptions) WithoutColor() PrintListingOptions {
	o.MsgColor = fmt.Fprintf
	o.PlainColor = fmt.Fprintf
	return o
}

const defaultContext = 3

func (o *PrintListingOptions) setDefaults() {
	if o.Context == 0 {
		o.Context = defaultContext
	}
	if o.MsgColor == nil {
		o.MsgColor = color.New(color.FgRed).Fprintf
	}
	if o.PlainColor == nil {
		o.PlainColor = color.New(color.Reset).Fprintf
	}
}

const (
	// BugLine is a fallback line when the line is not available.
	BugLine = `Cannot render line properly, please fill a bug report`

	leftPad           = "  "
	verticalBorder    = "|"
	horizontalPointer = "\u2192"
)

// PrintListing prints given message with line number and file listing to the writer.
//
// The context parameter defines the number of lines to print before and after.
func (f File) PrintListing(w io.Writer, msg string, pos Position, opts PrintListingOptions) error {
	opts.setDefaults()
	return f.PrintHighlights(w, msg, []Highlight{
		{Pos: pos, Color: opts.MsgColor},
	}, opts)
}

// Highlight is a highlighted position.
type Highlight struct {
	Pos   Position
	Color ColorFunc
}

// clamp keeps val in given boundaries.
func clamp[T constraints.Integer](val, lo, hi T) T {
	switch {
	case val < lo:
		return lo
	case val > hi:
		return hi
	default:
		return val
	}
}

func log10(val int) (r int) {
	for val >= 10 {
		r++
		val /= 10
	}
	return r
}

type lineNumberPad struct {
	pad, line int
}

func (p lineNumberPad) Format(f fmt.State, verb rune) {
	padding := p.pad - log10(p.line)
	var buf [32]byte
	for i := 0; i < padding; i++ {
		buf[i] = ' '
	}
	b := strconv.AppendInt(buf[:padding], int64(p.line), 10)
	_, _ = f.Write(b)
}

// PrintHighlights prints all given highlights.
func (f File) PrintHighlights(w io.Writer, msg string, highlights []Highlight, opts PrintListingOptions) error {
	opts.setDefaults()

	if len(highlights) < 1 {
		return errors.New("empty highlights")
	}

	var (
		l = f.Lines

		first      = highlights[0]
		lowestIdx  = first.Pos.Line - 1
		highestIdx = lowestIdx
	)
	for idx, h := range highlights {
		// Line starts from 1, but index starts from 0.
		hlightIdx := h.Pos.Line - 1
		if len(l.data) == 0 || hlightIdx < 0 || hlightIdx > len(l.lines) {
			return errors.Errorf("highlight %d: line number %d is out of range [0, %d)", idx, hlightIdx, len(l.lines))
		}

		switch {
		case hlightIdx < lowestIdx:
			lowestIdx = hlightIdx
		case hlightIdx > highestIdx:
			highestIdx = hlightIdx
		}
	}

	lowestIdx = clamp(lowestIdx-opts.Context, 0, len(l.lines))
	highestIdx = clamp(highestIdx+opts.Context+1, 0, len(l.lines))
	padNum := clamp(log10(highestIdx), 2, math.MaxInt)

	var (
		filename     = f.HumanName()
		formattedMsg string
	)
	if msg != "" {
		formattedMsg = " -> " + msg
	}

	if _, err := opts.MsgColor(w, "%s- %s%s\n",
		leftPad,
		first.Pos.WithFilename(filename),
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
		lineNumber := lineNumberPad{
			pad: padNum,
			// Line number is 1-based.
			line: n + 1,
		}
		// Line number is 1-based, but index is 0-based.
		lineText := line(n + 1)
		_, err := colored(w, "\t%s%d %s %s\t\n", leftPad, lineNumber, verticalBorder, lineText)
		return err
	}

	// Print lines.
	for idx := lowestIdx; idx <= highestIdx; idx++ {
		var (
			lineColor = opts.PlainColor
			pad       = leftPad
		)

		highlight, ok := xslices.FindFunc(highlights, func(h Highlight) bool {
			return h.Pos.Line-1 == idx
		})
		if ok {
			lineColor = highlight.Color
			pad = horizontalPointer + " "
		}

		if err := printLine(pad, idx, lineColor); err != nil {
			return err
		}

		// TODO(tdakkota): column pointer?
	}

	return nil
}
