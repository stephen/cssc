package ast

// AtRule represents an import statement.
type AtRule struct {
	Loc

	Name string

	Prelude AtPrelude

	Block *Block
}

func (String) isAtPrelude() {}

var _ AtPrelude = String{}

// AtPrelude is the set of arguments for an at-rule.
// The interface is only used for type discrimination.
type AtPrelude interface {
	Node

	isAtPrelude()
}

var _ Node = AtRule{}

func (r AtRule) isNode() {}
