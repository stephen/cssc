package ast

// AtRule represents an import statement.
type AtRule struct {
	Span

	Name string

	Preludes []AtPrelude

	Block Block
}

func (String) isAtPrelude()     {}
func (Identifier) isAtPrelude() {}

var _ AtPrelude = String{}
var _ AtPrelude = Identifier{}

// AtPrelude is the set of arguments for an at-rule.
// The interface is only used for type discrimination.
type AtPrelude interface {
	Node

	isAtPrelude()
}

var _ Node = AtRule{}

func (r AtRule) isNode() {}
