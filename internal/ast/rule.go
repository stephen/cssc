package ast

// QualifiedRule is a prelude (selectors) and set of declarations.
type QualifiedRule struct {
	Span

	Prelude Prelude

	Block Block
}

// Location implements Node.
func (n *QualifiedRule) Location() *Span { return &n.Span }

// Prelude is the prelude for QualifiedRules.
type Prelude interface {
	Node

	isPrelude()
}

var _ Node = &QualifiedRule{}
