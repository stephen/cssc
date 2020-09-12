package logging

import (
	"fmt"
	"strings"

	"github.com/stephen/cssc/internal/sources"
)

// LocationError is an error that happened at a specific location
// in the source.
type LocationError struct {
	Message string

	Source *sources.Source
	start  int
	length int
}

// NewLocationError returns a new error from a source location.
func NewLocationError(source *sources.Source, start, end int, f string, args ...interface{}) error {
	return &LocationError{fmt.Sprintf(f, args...), source, start, end - start}
}

// Error implements error. It's relatively slow because it needs to
// rescan the source to figure out line and column numbers. The output
// looks like:
// file.css:1:1
// there's a problem here:
//   contents
//   ~~~~~~~~
func (l *LocationError) Error() string {
	lineNumber, lineStart := 1, 0
	for i, ch := range l.Source.Content[:l.start] {
		if ch == '\n' {
			lineNumber++
			lineStart = i + 1
		}
	}

	lineEnd := len(l.Source.Content)
	for i, ch := range l.Source.Content[l.start:] {
		if ch == '\n' {
			lineEnd = i + l.start
			break
		}
	}

	line := l.Source.Content[lineStart:lineEnd]
	col := l.start - lineStart

	tabCount := strings.Count(line, "\t")
	withoutTabs := strings.ReplaceAll(line, "\t", "  ")

	indent := strings.Repeat(" ", col+tabCount)
	underline := strings.Repeat("~", l.length)

	return fmt.Sprintf("%s:%d:%d\n%s:\n\t%s\n\t%s%s", l.Source.Path, lineNumber, col, l.Message, withoutTabs, indent, underline)
}
