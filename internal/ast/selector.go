package ast

// SelectorList is a type of prelude for QualifiedRule,
// containing a list of selectors separated by commas.
type SelectorList struct {
	Loc

	Selectors []*Selector
}

func (SelectorList) isPrelude() {}

var _ Prelude = SelectorList{}

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
	Node

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
	Arguments *SelectorList
}

// PseudoElementSelector selects a pseudo element, e.g. ::before or ::after.
type PseudoElementSelector struct {
	Loc

	Inner *PseudoClassSelector
}

// Whitespace represents any whitespace sequence. Whitespace is
// only kept in the AST when necessary for disambiguating syntax,
// e.g. in selectors.
type Whitespace struct {
	Loc
}

// AttributeSelector selects elements with the specified attributes matching.
// Note that the = token is implied if Value is non-zero.
type AttributeSelector struct {
	Loc

	// Property is the attribute to check.
	Property string

	// PreOperator can be ~, ^, $, *.
	// See: https://www.w3.org/TR/selectors-4/#attribute-representation.
	PreOperator string

	// Value is the value to match against.
	Value Value
}

var _ SelectorPart = TypeSelector{}
var _ SelectorPart = ClassSelector{}
var _ SelectorPart = IDSelector{}
var _ SelectorPart = CombinatorSelector{}
var _ SelectorPart = PseudoClassSelector{}
var _ SelectorPart = PseudoElementSelector{}
var _ SelectorPart = Whitespace{}
var _ SelectorPart = AttributeSelector{}

func (TypeSelector) isSelector()          {}
func (ClassSelector) isSelector()         {}
func (IDSelector) isSelector()            {}
func (CombinatorSelector) isSelector()    {}
func (PseudoClassSelector) isSelector()   {}
func (PseudoElementSelector) isSelector() {}
func (Whitespace) isSelector()            {}
func (AttributeSelector) isSelector()     {}
