package ast

// AtRule represents an import statement.
type AtRule struct {
	Loc

	Name string

	Prelude Prelude

	Block interface{}
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
	isPrelude()
}

var _ Node = AtRule{}

func (r AtRule) isNode() {}
