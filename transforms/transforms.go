package transforms

// ImportRules controls transform behavior for @imports.
type ImportRules int

const (
	// ImportRulesPassthrough passes @imports down without changes. It is the default.
	ImportRulesPassthrough ImportRules = iota
	// ImportRulesInline inlines imported content where an @import statement is seen. In this
	// version, it ignores @supports rules and meedia queries.
	ImportRulesInline
)

// MediaFeatureRanges controls transform options for feature ranges,
// introduced in CSS Media Queries Level 4.
// See: https://www.w3.org/TR/mediaqueries-4/#mq-range-context.
type MediaFeatureRanges int

const (
	// MediaFeatureRangesPassthrough passes @imports down without changes. It is the default.
	MediaFeatureRangesPassthrough MediaFeatureRanges = iota
	// MediaFeatureRangesTransform transforms ranges into best-effort min- and max- values. When
	// > and < are used, we follow the guidance from https://www.w3.org/TR/mediaqueries-5/#mq-min-max and
	// use min/max with a change in .001 precision.
	MediaFeatureRangesTransform
)

// AnyLink controls transform options for :any-link selectors.
// introduced in CSS Selectors Level 4.
// See: https://www.w3.org/TR/selectors-4/#the-any-link-pseudo
type AnyLink int

const (
	// AnyLinkPassthrough passes :any-link down without changes. It is the default.
	AnyLinkPassthrough AnyLink = iota
	// AnyLinkTransform transforms :any-link selectors into selectors for both :visited and :link.
	AnyLinkTransform
)

// CustomProperties controls transform options for custom properteries (--var) and the var() function
// from CSS Variables Level 1.
// See: https://www.w3.org/TR/css-variables-1/.
type CustomProperties int

const (
	// CustomPropertiesPassthrough passes variable declarations and var() down without changes. It is the default.
	CustomPropertiesPassthrough CustomProperties = iota
	// CustomPropertiesTransformRoot will transform properties defiend in :root selectors. Custom property definitions
	// under any other selectors will be ignored and passed through.
	CustomPropertiesTransformRoot
)

// CustomMediaQueries controls transform options for @custom-media usage, specified in CSS Media Queries Level 5.
// See: https://www.w3.org/TR/mediaqueries-5/#custom-mq.
type CustomMediaQueries int

const (
	// CustomMediaQueriesPassthrough passes custom media query definitions and usages through. It is the default.
	CustomMediaQueriesPassthrough CustomMediaQueries = iota
	// CustomMediaQueriesTransform will transform custom media queries when used in @media rules.
	CustomMediaQueriesTransform
)

// Options sets options about what transforms to run. By default,
// no transforms are run.
type Options struct {
	ImportRules
	MediaFeatureRanges
	AnyLink
	CustomProperties
	CustomMediaQueries
}
