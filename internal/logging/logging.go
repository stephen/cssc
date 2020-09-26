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
func LocationErrorf(source *sources.Source, span ast.Span, f string, args ...interface{}) error {
	return &locationError{fmt.Errorf(f, args...), false, source, span}
}

// LocationWarnf adds a warning from a specific location.
func LocationWarnf(source *sources.Source, span ast.Span, f string, args ...interface{}) error {
	return &locationError{fmt.Errorf(f, args...), true, source, span}
}

// locationError is an error that happened at a specific location
// in the source.
type locationError struct {
	inner error

	warning bool

	Source *sources.Source
	ast.Span
}

// Unwrap satisfies errors.Unwrap.
func (l *locationError) Unwrap() error {
	return l.inner
}

// Error implements error. It's relatively slow because it needs to
// rescan the source to figure out line and column numbers. The output
// looks like:
// file.css:1:4
// there's a problem here:
//   (contents) or (other thing)
//    ~~~~~~~~
func (l *locationError) Error() string {
	// Unfortunately, AnnotateSourceSpan will also call LineAndCol, so we'll
	// end up duplicating that effort. It's probably okay since we're unlikely to be
	// generating high-throughput errors.
	lineNumber, col := l.Source.LineAndCol(l.Span)

	return fmt.Sprintf("%s:%d:%d\n%s:\n%s", l.Source.Path, lineNumber, col, l.inner.Error(), AnnotateSourceSpan(l.Source, l.Span))
}

// AnnotateSourceSpan annotates a span from a single line in the source code.
// The output looks like:
//   (contents) or (other thing)
//    ~~~~~~~~
func AnnotateSourceSpan(source *sources.Source, span ast.Span) string {
	lineNumber, col := source.LineAndCol(span)
	lineStart := source.Lines[lineNumber-1]

	lineEnd := len(source.Content)
	for i, ch := range source.Content[span.Start:] {
		if ch == '\n' {
			lineEnd = i + span.Start
			break
		}
	}

	line := source.Content[lineStart:lineEnd]

	tabCount := strings.Count(line, "\t")
	withoutTabs := strings.ReplaceAll(line, "\t", "  ")

	indent := strings.Repeat(" ", int(col)+tabCount-1)
	underline := strings.Repeat("~", span.End-span.Start)[:]
	excessMarker := ""
	if excess := span.End - lineEnd; excess > 0 {
		underline = underline[:len(underline)-excess]
		excessMarker = ">"
	}

	return fmt.Sprintf("\t%s\n\t%s%s%s", withoutTabs, indent, underline, excessMarker)
}
