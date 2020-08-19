package ast

// QualifiedRule is a prelude (selectors) and set of declarations.
type QualifiedRule struct {
	Loc

	Prelude Prelude

	Block Block
}

// Prelude is the prelude for QualifiedRules.
type Prelude interface {
	Node

	isPrelude()
}

var _ Node = QualifiedRule{}

func (r QualifiedRule) isNode() {}
