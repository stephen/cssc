package ast

// Value is a css value, e.g. dimension, percentage, or number.
type Value interface {
	// isValue is only used for type discrimination.
	isValue()
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
	// Value is the string representation for the value.
	Value string
}

// Number is a number literal. It can be either an integer or
// real number from https://www.w3.org/TR/css-values-4/.
type Number struct {
	// Value is the string representation for the value.
	Value string
}

// Position is a 2D position, e.g. left, center, or bottom.
// See: https://www.w3.org/TR/css-values-4/#position.
type Position struct {
	// Value is the position.
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
	Keyword string

	// RGBA is the literal rgba value.
	RGBA string
}

func (Dimension) isValue()  {}
func (Percentage) isValue() {}
func (Number) isValue()     {}

var _ Value = Dimension{}
var _ Value = Percentage{}
var _ Value = Number{}
