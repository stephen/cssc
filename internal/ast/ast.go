package ast

// Node is any top-level stylesheet rule.
type Node interface {
	Location() Span
}

// Span is a range of text in the source.
type Span struct {
	// Start is the start of the range, inclusive.
	Start int

	// End is the end of the range, exclusive.
	End int
}

// Location implements Node.
func (l Span) Location() Span { return l }

// Stylesheet is a CSS stylesheet.
type Stylesheet struct {
	Nodes []Node

	Imports []ImportSpecifier
}

// ImportSpecifier is a pointer to an import at rule.
type ImportSpecifier struct {
	Value string

	// AtRule is a pointer to the at rule that specified this import.
	AtRule *AtRule
}

// Location implements Node.
func (l Stylesheet) Location() Span {
	if len(l.Nodes) == 0 {
		return Span{}
	}

	return Span{l.Nodes[0].Location().Start, l.Nodes[len(l.Nodes)-1].Location().End}
}

// Block can either be a block of rules or declarations.
// See https://www.w3.org/TR/css-syntax-3/#declaration-rule-list.
type Block interface {
	Node

	isBlock()
}

// DeclarationBlock is a block containing a set of declarations.
type DeclarationBlock struct {
	Span

	Declarations []Declarationish
}

// Declarationish is a Declaration or a Raw value.
type Declarationish interface {
	Node
	isDeclaration()
}

// QualifiedRuleBlock is a block containing a set of rules.
type QualifiedRuleBlock struct {
	Span

	Rules []*QualifiedRule
}

func (DeclarationBlock) isBlock()   {}
func (QualifiedRuleBlock) isBlock() {}

var _ Block = DeclarationBlock{}
var _ Block = QualifiedRuleBlock{}

// Declaration is a property assignment, e.g. width: 2px.
type Declaration struct {
	Span

	// Property is the property being assigned.
	Property string

	// Values is the list of values assigned to the declaration.
	Values []Value

	// Important is whether or not the declaration was marked !important.
	Important bool
}

func (Declaration) isDeclaration() {}
func (Raw) isDeclaration()         {}

var _ Declarationish = Declaration{}
var _ Declarationish = Raw{}
