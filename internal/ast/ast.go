package ast

// Node is any top-level stylesheet rule.
type Node interface {
	Location() Loc
}

// Loc is a location in the source.
type Loc struct {
	Position int
}

// Location implements Node.
func (l Loc) Location() Loc { return l }

// Stylesheet is a CSS stylesheet.
type Stylesheet struct {
	Nodes []Node
}

// Location implements Node.
func (l Stylesheet) Location() Loc { return Loc{} }

// Comment represents a comment.
type Comment struct {
	Loc

	Text string
}

// Block can either be a block of rules or declarations.
// See https://www.w3.org/TR/css-syntax-3/#declaration-rule-list.
type Block interface {
	Node

	isBlock()
}

// DeclarationBlock is a block containing a set of declarations.
type DeclarationBlock struct {
	Loc

	Declarations []*Declaration
}

// QualifiedRuleBlock is a block containing a set of rules.
type QualifiedRuleBlock struct {
	Loc

	Rules []*QualifiedRule
}

func (DeclarationBlock) isBlock()   {}
func (QualifiedRuleBlock) isBlock() {}

var _ Block = DeclarationBlock{}
var _ Block = QualifiedRuleBlock{}

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
