package ast

// AtRule represents an import statement.
type AtRule struct {
	Loc

	Name string

	Prelude Prelude

	Block *Block
}

// ImportPrelude is the target for an import statement.
type ImportPrelude struct {
	Loc

	URL string
}

var _ Prelude = ImportPrelude{}

func (ImportPrelude) isPrelude() {}

// Prelude is the set of arguments for an at-rule.
// The interface is only used for type discrimination.
type Prelude interface {
	Node

	isPrelude()
}

var _ Node = AtRule{}

func (r AtRule) isNode() {}
