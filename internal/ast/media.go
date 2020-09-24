package ast

// MediaQueryList is a comma-separated list of media queries.
type MediaQueryList struct {
	Span

	Queries []*MediaQuery
}

func (MediaQueryList) isAtPrelude() {}

var _ AtPrelude = MediaQueryList{}

// MediaQuery is a single media query.
type MediaQuery struct {
	Span

	Parts []MediaQueryPart
}

// isAtPrelude implements AtPrelude for @custom-media rules.
func (MediaQuery) isAtPrelude() {}

var _ AtPrelude = MediaQuery{}

// MediaQueryPart is a part of a media query, e.g. a MediaFeature,
// MediaType, or MediaCombinator.
type MediaQueryPart interface {
	Node

	isMediaQueryPart()
}

func (Identifier) isMediaQueryPart()        {}
func (MediaFeaturePlain) isMediaQueryPart() {}
func (MediaFeatureRange) isMediaQueryPart() {}
func (MediaInParens) isMediaQueryPart()     {}

var _ MediaQueryPart = Identifier{}
var _ MediaQueryPart = MediaFeaturePlain{}
var _ MediaQueryPart = MediaFeatureRange{}
var _ MediaQueryPart = MediaInParens{}

// MediaInParens is a media expression in parenthesis. It is
// different from MediaQuery in that it implements MediaQueryPart.
type MediaInParens struct {
	Span

	Parts []MediaQueryPart
}

// MediaType is a specific media type.
type MediaType struct {
	Span

	Value string
}

// MediaFeature is fine-grained test for a media feature,
// enclosed in parenthesis.
type MediaFeature interface {
	Node
	MediaQueryPart

	isMediaFeature()
}

func (MediaFeaturePlain) isMediaFeature() {}
func (MediaFeatureRange) isMediaFeature() {}

var _ MediaFeature = MediaFeaturePlain{}
var _ MediaFeature = MediaFeatureRange{}

// MediaFeaturePlain is a equivalence check.
// e.g. (width: 500px) or (color).
type MediaFeaturePlain struct {
	Span

	Property *Identifier
	Value    Value
}

// MediaFeatureRange is a type of media feature that looks
// like value < name < value or value > name > value.
type MediaFeatureRange struct {
	Span

	Property *Identifier

	LeftValue  Value
	Operator   string
	RightValue Value
}
