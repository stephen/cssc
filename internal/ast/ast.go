package ast

// Node is any top-level stylesheet rule.
type Node interface {
	Location() int

	// isNode is only used for type discrimination.
	isNode()
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

	Parts []SelectorPart
}

// SelectorPart is a part of a complex selector. It maybe be e.g.
// a class or id selector, or a + or < combinator, or a pseudoselector.
//
// The interface is only used for type discrimination.
type SelectorPart interface {
	isSelector()
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

var _ SelectorPart = TypeSelector{}
var _ SelectorPart = ClassSelector{}
var _ SelectorPart = IDSelector{}
var _ SelectorPart = CombinatorSelector{}
var _ SelectorPart = PseudoClassSelector{}
var _ SelectorPart = PseudoElementSelector{}

func (TypeSelector) isSelector()          {}
func (ClassSelector) isSelector()         {}
func (IDSelector) isSelector()            {}
func (CombinatorSelector) isSelector()    {}
func (PseudoClassSelector) isSelector()   {}
func (PseudoElementSelector) isSelector() {}

var _ Node = AtRule{}
var _ Node = Comment{}
var _ Node = Block{}
var _ Node = Declaration{}
var _ Node = QualifiedRule{}

func (r AtRule) isNode()        {}
func (r Comment) isNode()       {}
func (r Block) isNode()         {}
func (r Declaration) isNode()   {}
func (r QualifiedRule) isNode() {}
