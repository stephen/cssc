package ast

// Node is any top-level stylesheet rule.
type Node interface {
	Location() int

	// isNode is only used for type discrimination.
	isNode()
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

// Comment represents a comment.
type Comment struct {
	Loc

	Text string
}

// Block is a set of declarations.
type Block struct {
	Loc

	Declarations []*Declaration
}

// Declaration is a property assignment, e.g. width: 2px.
type Declaration struct {
	Loc

	// Property is the property being assigned.
	Property string
}

var _ Node = Comment{}
var _ Node = Block{}
var _ Node = Declaration{}

func (r Comment) isNode()     {}
func (r Block) isNode()       {}
func (r Declaration) isNode() {}
