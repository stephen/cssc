package lexer

import (
	"sort"

	"github.com/stephen/cssc/internal/ast"
)

// Source is a container for a file and its contents.
type Source struct {
	// Path is the path of the source file.
	Path string

	// Content is the content of the file.
	Content string

	// lines is the offset of the beginning of every line. This
	// is useful for quickly finding the line and column for a
	// given byte offset (ast.Loc).
	lines []int
}

// LineAndCol computes the 1-index line and column for a given
// ast.Loc (byte offset in the file).
func (s *Source) LineAndCol(loc ast.Loc) (int32, int32) {
	line := sort.Search(len(s.lines), func(i int) bool {
		return loc.Position < s.lines[i]
	})

	return int32(line), int32(loc.Position - s.lines[line-1] + 1)
}
