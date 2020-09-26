package ast

// QualifiedRule is a prelude (selectors) and set of declarations.
type QualifiedRule struct {
	Span

	Prelude Prelude

	Block Block
}

// Prelude is the prelude for QualifiedRules.
type Prelude interface {
	Node

	isPrelude()
}

var _ Node = QualifiedRule{}
