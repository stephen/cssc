package ast

// Value is a css value, e.g. dimension, percentage, or number.
type Value interface {
	Node

	// isValue is only used for type discrimination.
	isValue()
}

// String is a string literal.
type String struct {
	Loc

	// Value is the string.
	Value string
}

// Dimension is a numeric value and a unit.
type Dimension struct {
	Loc

	// Value is the string representation for the value.
	Value string

	// Unit is the unit (e.g. rem, px) for the dimension.
	Unit string
}

// Percentage is a numeric percentage.
type Percentage struct {
	Loc

	// Value is the string representation for the value.
	Value string
}

// Number is a number literal. It can be either an integer or
// real number from https://www.w3.org/TR/css-values-4/.
type Number struct {
	Loc

	// Value is the string representation for the value.
	Value string
}

// Identifier is any string identifier value, e.g. inherit or left.
type Identifier struct {
	Loc

	// Value is the identifier.
	Value string
}

// Image is an image. Only one of the URL or Gradient fields
// can be validly non-zero.
// See: https://www.w3.org/TR/css-images-3/.
type Image struct {
	// URL is the referenced image.
	URL string

	// Gradient is a gradient defined by
	// https://www.w3.org/TR/css-images-3/#gradients.
	Gradient string
}

// HexColor is a hex color (e.g. #aabbccdd) defined by https://www.w3.org/TR/css-color-3/.
type HexColor struct {
	Loc

	// RGBA is the literal rgba value.
	RGBA string
}

// Function is a css function.
type Function struct {
	Loc

	// Name is the name of the function.
	Name string

	// Arguments is the set of values passed into the function.
	Arguments []Value
}

// IsMath returns whether or not this function supports math expressions
// as values.
func (f Function) IsMath() bool {
	_, ok := mathFunctions[f.Name]
	return ok
}

var mathFunctions = map[string]struct{}{
	"calc":  struct{}{},
	"min":   struct{}{},
	"max":   struct{}{},
	"clamp": struct{}{},
}

// MathExpression is a binary expression for math functions.
type MathExpression struct {
	Loc

	// Operator +, -, *, or /.
	Operator string

	Left  Value
	Right Value
}

// Comma is a single comma. Some declarations require commas,
// e.g. font-family fallbacks or transitions.
type Comma struct {
	Loc
}

func (String) isValue()         {}
func (Dimension) isValue()      {}
func (Percentage) isValue()     {}
func (Number) isValue()         {}
func (Function) isValue()       {}
func (MathExpression) isValue() {}
func (Comma) isValue()          {}
func (Identifier) isValue()     {}
func (HexColor) isValue()       {}

var _ Value = String{}
var _ Value = Dimension{}
var _ Value = Percentage{}
var _ Value = Number{}
var _ Value = Function{}
var _ Value = MathExpression{}
var _ Value = Comma{}
var _ Value = Identifier{}
var _ Value = HexColor{}
