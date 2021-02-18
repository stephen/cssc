package ast

// Value is a css value, e.g. dimension, percentage, or number.
type Value interface {
	Node

	// isValue is only used for type discrimination.
	isValue()
}

// String is a string literal.
type String struct {
	Span

	// Value is the string.
	Value string
}

// Dimension is a numeric value and a unit. Dimension can
// also represent Percentages (% unit) or Numbers (empty string unit).
type Dimension struct {
	Span

	// Value is the string representation for the value.
	Value string

	// Unit is the unit (e.g. rem, px) for the dimension. If Unit
	// is empty, then it's a CSS number type.
	Unit string
}

// Percentage is a numeric percentage.
type Percentage struct {
	Span

	// Value is the string representation for the value.
	Value string
}

// Identifier is any string identifier value, e.g. inherit or left.
type Identifier struct {
	Span

	// Value is the identifier.
	Value string
}

// HexColor is a hex color (e.g. #aabbccdd) defined by https://www.w3.org/TR/css-color-3/.
type HexColor struct {
	Span

	// RGBA is the literal rgba value.
	RGBA string
}

// Brackets is a bracketized value.
type Brackets struct {
	Span

	// Values is the inner values, space-separated.
	Values []Value
}

// Function is a css function.
type Function struct {
	Span

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
	Span

	// Operator +, -, *, or /.
	Operator string

	Left  Value
	Right Value
}

// MathParenthesizedExpression is a parenthesized math expression.
type MathParenthesizedExpression struct {
	Span

	Value Value
}

// Raw is an otherwise non-structured but valid value.
type Raw struct {
	Span

	Value string
}

// Comma is a single comma. Some declarations require commas,
// e.g. font-family fallbacks or transitions.
type Comma struct {
	Span
}

func (String) isValue()                      {}
func (Dimension) isValue()                   {}
func (Brackets) isValue()                    {}
func (Function) isValue()                    {}
func (MathExpression) isValue()              {}
func (MathParenthesizedExpression) isValue() {}
func (Comma) isValue()                       {}
func (Identifier) isValue()                  {}
func (HexColor) isValue()                    {}
func (Raw) isValue()                         {}

var _ Value = String{}
var _ Value = Dimension{}
var _ Value = Brackets{}
var _ Value = Function{}
var _ Value = MathExpression{}
var _ Value = MathParenthesizedExpression{}
var _ Value = Comma{}
var _ Value = Identifier{}
var _ Value = HexColor{}
var _ Value = Raw{}
