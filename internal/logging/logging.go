package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/sources"
)

// Reporter is an interface for reporting errors and warnings.
type Reporter interface {
	AddError(error)
}

// DefaultReporter is the default reporter, which writes to stderr.
var DefaultReporter = WriterReporter{os.Stderr}

// WriterReporter is a simple adapter for writing logs to an io.Writer.
// The default reporter is a WriterReporter(os.Stderr).
type WriterReporter struct {
	io.Writer
}

// AddError implements Reporter.
func (w WriterReporter) AddError(err error) {
	fmt.Fprintln(w, err.Error())
}

// LocationErrorf adds an error from a specific location.
func LocationErrorf(source *sources.Source, start, end int, f string, args ...interface{}) error {
	return &locationError{fmt.Errorf(f, args...), false, source, start, end - start}
}

// LocationWarnf adds a warning from a specific location.
func LocationWarnf(source *sources.Source, start, end int, f string, args ...interface{}) error {
	return &locationError{fmt.Errorf(f, args...), true, source, start, end - start}
}

// locationError is an error that happened at a specific location
// in the source.
type locationError struct {
	inner error

	warning bool

	Source *sources.Source
	start  int
	length int
}

// Unwrap satisfies errors.Unwrap.
func (l *locationError) Unwrap() error {
	return l.inner
}

// Error implements error. It's relatively slow because it needs to
// rescan the source to figure out line and column numbers. The output
// looks like:
// file.css:1:1
// there's a problem here:
//   contents
//   ~~~~~~~~
func (l *locationError) Error() string {
	lineNumber, col := l.Source.LineAndCol(ast.Span{Start: l.start, End: l.start + l.length})
	lineStart := l.Source.Lines[lineNumber-1]

	lineEnd := len(l.Source.Content)
	for i, ch := range l.Source.Content[l.start:] {
		if ch == '\n' {
			lineEnd = i + l.start
			break
		}
	}

	line := l.Source.Content[lineStart:lineEnd]

	tabCount := strings.Count(line, "\t")
	withoutTabs := strings.ReplaceAll(line, "\t", "  ")

	indent := strings.Repeat(" ", int(col)+tabCount-1)
	underline := strings.Repeat("~", l.length)

	return fmt.Sprintf("%s:%d:%d\n%s:\n\t%s\n\t%s%s", l.Source.Path, lineNumber, col, l.inner.Error(), withoutTabs, indent, underline)
}
