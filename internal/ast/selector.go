package ast

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
	Children []*Selector
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
