package ast

// Node is any AST node.
type Node interface {
	Location() int
	node()
}

// Loc is a location in the source.
type Loc struct {
	Position int
}

// Location implements Node.
func (l Loc) Location() int { return l.Position }

// Stylesheet is a CSS stylesheet.
type Stylesheet struct {
	Nodes []Node
}

// ImportAtRule represents an import statement.
type ImportAtRule struct {
	Loc

	// Target is the URL or string location to import.
	URL string
}

// MediaAtRule represents a @media rule.
type MediaAtRule struct {
	Loc

	// Query
	// XXX: model out the media query.
	Query string
}

// Comment represents a comment.
type Comment struct {
	Loc

	Text string
}

// QualifiedRule is a prelude (selectors) and set of declarations.
type QualifiedRule struct {
	Loc

	Selectors *SelectorList

	Block *Block
}

// Block is a set of declarations.
type Block struct {
	Loc

	Declarations []*Declaration
}

// Declaration is a property assignment.
type Declaration struct {
	Loc

	// Property is the property being assigned.
	Property string
}

// SelectorList is a list of selectors, separated by commas, e.g.
// .class, .another-class.
type SelectorList struct {
	Loc

	Selectors []*Selector
}

// Selector represents a single selector. From the selectors level 4
// spec, a selector is a flat representation of complex-selector,
// compound-selector, type-selector, combinator, etc, since we mostly
// just want tokens to work with.
type Selector struct {
	Loc

	Selectors []interface{}
}

// TypeSelector selects a single type, e.g. div, body, or html.
type TypeSelector struct {
	Loc

	Name string
}

// ClassSelector selects a single class, e.g. .test or .Thing.
type ClassSelector struct {
	Loc

	Name string
}

// IDSelector selects a single ID, e.g. #container.
type IDSelector struct {
	Loc

	Name string
}

// CombinatorSelector operates between two selectors.
type CombinatorSelector struct {
	Loc

	// The combinator operation, i.e. >, +, ~, or |.
	Operator string
}

// PseudoClassSelector selects a pseudo class, e.g. :not() or :hover.
type PseudoClassSelector struct {
	Loc

	// Name is the name of the pseudo selector.
	Name string

	// Children holds any arguments to the selector, if specified.
	Children *SelectorList
}

// PseudoElementSelector selects a pseudo element, e.g. ::before or ::after.
type PseudoElementSelector struct {
	Loc

	Inner *PseudoClassSelector
}

var _ Node = ImportAtRule{}
var _ Node = Comment{}
var _ Node = Block{}
var _ Node = Declaration{}
var _ Node = QualifiedRule{}

func (r ImportAtRule) node()  {}
func (r Comment) node()       {}
func (r Block) node()         {}
func (r Declaration) node()   {}
func (r QualifiedRule) node() {}
