package api

// ImportRules controls transform behavior for @imports.
type ImportRules int

const (
	// ImportRulesPassthrough passes @imports down without changes. It is the default.
	ImportRulesPassthrough ImportRules = iota
	// ImportRulesInline inlines imported content where an @import statement is seen. In this
	// version, it ignores @supports rules and meedia queries.
	ImportRulesInline
)

// TransformOptions sets options about what transforms to run. By default,
// no transforms are run.
type TransformOptions struct {
	ImportRules
}
