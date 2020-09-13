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
	// MediaFeatureRangesTransform transforms ranges into best-effort min- and max- values. In this version,
	// it only supports <= and >= syntax and will fail to transform < and > syntax.
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

// Options sets options about what transforms to run. By default,
// no transforms are run.
type Options struct {
	ImportRules
	MediaFeatureRanges
	AnyLink
}
