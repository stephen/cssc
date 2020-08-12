package ast

// Node is any AST node.
type Node interface {
	Location() int
}

// Loc is a location in the source.
type Loc struct {
	Position int
}

// Location implements Node.
func (l Loc) Location() int { return l.Position }

// Stylesheet is a CSS stylesheet.
type Stylesheet struct {
	Rules []Rule
}

// Rule is a top-level rule in the stylesheet.
type Rule interface {
	Node
	rule()
}

// ImportAtRule represents an import statement.
type ImportAtRule struct {
	*Loc

	// Target is the URL or string location to import.
	URL string
}

var _ Rule = ImportAtRule{}

// rule implements Rule.
func (r ImportAtRule) rule() {}
