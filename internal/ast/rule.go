package ast

// QualifiedRule is a prelude (selectors) and set of declarations.
type QualifiedRule struct {
	Loc

	Selectors []*Selector

	Block Block
}

var _ Node = QualifiedRule{}

func (r QualifiedRule) isNode() {}
