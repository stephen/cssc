package ast

// Node is any AST node.
type Node interface {
	Location() int
	node()
}

// Loc is a location in the source.
type Loc struct {
	Position int
}

// Location implements Node.
func (l Loc) Location() int { return l.Position }

// Stylesheet is a CSS stylesheet.
type Stylesheet struct {
	Nodes []Node
}

// ImportAtRule represents an import statement.
type ImportAtRule struct {
	*Loc

	// Target is the URL or string location to import.
	URL string
}

// Comment represents a comment.
type Comment struct {
	*Loc

	Text string
}

var _ Node = ImportAtRule{}
var _ Node = Comment{}

func (r ImportAtRule) node() {}
func (r Comment) node()      {}
