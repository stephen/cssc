package ast

// Node is any top-level stylesheet rule.
type Node interface {
	Location() int
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

// Location implements Node.
func (l Stylesheet) Location() int { return 0 }

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

	// Values is the list of values assigned to the declaration.
	Values []Value

	// Important is whether or not the declaration was marked !important.
	Important bool
}
