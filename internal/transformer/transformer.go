package transformer

import (
	"fmt"
	"strings"

	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/logging"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/transforms"
)

// Options is the set of options for transformation.
type Options struct {
	// OriginalSource is used to report error locations.
	// XXX: work even if original source is not passed by having janky errors.
	OriginalSource *sources.Source

	// Reporter is the reporter for errors and warnings.
	Reporter logging.Reporter

	// Options is the set of transform options. Some transforms may need additional context passed in.
	transforms.Options

	// ImportReplacements is the set of import references to inline. ImportReplacements must be non-nil
	// if ImportRules is set to ImportRulesInline.
	ImportReplacements map[*ast.AtRule]*ast.Stylesheet
}

// Transform takes a pass over the input AST and runs various
// transforms.
func Transform(s *ast.Stylesheet, opts Options) *ast.Stylesheet {
	t := &transformer{
		Options:     opts,
		customMedia: make(map[string]*ast.MediaQuery),
	}

	if opts.Reporter == nil {
		t.Reporter = logging.DefaultReporter
	}

	if opts.CustomProperties != transforms.CustomPropertiesPassthrough {
		t.variables = make(map[string][]ast.Value)
	}

	if opts.ImportReplacements == nil && opts.ImportRules == transforms.ImportRulesInline {
		t.Reporter.AddError(fmt.Errorf("ImportRules is set to ImportRulesInline, but ImportReplacements is not set"))
	}

	s.Nodes = t.transformNodes(s.Nodes)

	return s
}

// transformer takes a pass over the AST and makes
// modifications to the AST, depending on the settings.
type transformer struct {
	Options

	variables   map[string][]ast.Value
	customMedia map[string]*ast.MediaQuery
}

func (t *transformer) addError(loc ast.Loc, fmt string, args ...interface{}) {
	// XXX: we don't have the end locations anymore, but we should...
	t.Reporter.AddError(logging.LocationErrorf(t.OriginalSource, loc.Location().Position, loc.Location().Position+1, fmt, args...))
}

func (t *transformer) addWarn(loc ast.Loc, fmt string, args ...interface{}) {
	// XXX: we don't have the end locations anymore, but we should...
	t.Reporter.AddError(logging.LocationWarnf(t.OriginalSource, loc.Location().Position, loc.Location().Position+1, fmt, args...))
}

func (t *transformer) transformSelectors(nodes []*ast.Selector) []*ast.Selector {
	newNodes := make([]*ast.Selector, 0, len(nodes))
	for _, n := range nodes {
		newParts := make([]ast.SelectorPart, 0, len(n.Parts))
		for index, p := range n.Parts {
			switch part := p.(type) {
			case *ast.PseudoClassSelector:
				if part.Name != "any-link" || t.AnyLink == transforms.AnyLinkPassthrough {
					newParts = append(newParts, p)
					break
				}

				// Replace one of them with :link.
				newParts = append(
					newParts,
					&ast.PseudoClassSelector{Name: "link"},
				)

				// Make a duplicate with :visited.
				duplicate := *n
				duplicate.Parts[index] = &ast.PseudoClassSelector{Name: "visited"}
				newNodes = append(newNodes, &duplicate)

			default:
				newParts = append(newParts, p)
			}
		}

		n.Parts = newParts
		newNodes = append(newNodes, n)
	}
	return newNodes
}

func (t *transformer) transformNodes(nodes []ast.Node) []ast.Node {
	rv := make([]ast.Node, 0, len(nodes))
	for _, value := range nodes {
		switch node := value.(type) {
		case *ast.QualifiedRule:
			func() {
				selList, ok := node.Prelude.(*ast.SelectorList)
				if !ok {
					return
				}

				isRoot := false
				for _, sel := range selList.Selectors {
					if len(sel.Parts) != 1 {
						isRoot = true
						break
					}
				}

				if !isRoot {
					return
				}

				rootSel, ok := selList.Selectors[0].Parts[0].(*ast.PseudoClassSelector)
				if !ok {
					return
				}

				if rootSel.Name != "root" {
					return
				}

				declBlock, ok := node.Block.(*ast.DeclarationBlock)
				if !ok {
					return
				}

				newDecls := make([]*ast.Declaration, 0, len(declBlock.Declarations))
				for _, decl := range declBlock.Declarations {
					if strings.HasPrefix(decl.Property, "--") && t.variables != nil {
						t.variables[decl.Property] = decl.Values
						continue
					}

					newDecls = append(newDecls, decl)
				}

				declBlock.Declarations = newDecls
			}()

			selList, ok := node.Prelude.(*ast.SelectorList)
			if !ok {
				t.addError(node.Prelude.Location(), "expected selector list for qualified rule")
			}
			selList.Selectors = t.transformSelectors(selList.Selectors)
			node.Block = t.transformBlock(node.Block)

			if node.Block == nil {
				continue
			}
			rv = append(rv, node)

		case *ast.AtRule:
			switch node.Name {
			case "import":
				if t.ImportReplacements == nil {
					rv = append(rv, node)
				}

				imported, ok := t.ImportReplacements[node]
				if !ok {
					rv = append(rv, node)
					break
				}
				rv = append(rv, imported.Nodes...)

			case "custom-media":
				func() {
					if len(node.Preludes) != 2 {
						return
					}

					name, ok := node.Preludes[0].(*ast.Identifier)
					if !ok {
						return
					}

					query, ok := node.Preludes[1].(*ast.MediaQuery)
					if !ok {
						return
					}

					t.customMedia[name.Value] = query
				}()

			case "media":
				mq := node.Preludes[0].(*ast.MediaQueryList)
				mq.Queries = t.transformMediaQueries(mq.Queries)
				rv = append(rv, node)

			default:
				rv = append(rv, node)
			}

		default:
			rv = append(rv, value)
		}
	}
	return rv
}

func (t *transformer) transformMediaQueries(queries []*ast.MediaQuery) []*ast.MediaQuery {
	newQueries := make([]*ast.MediaQuery, 0, len(queries))
	for _, q := range queries {
		q.Parts = t.transformMediaQueryParts(q.Parts)
		newQueries = append(newQueries, q)
	}
	return newQueries
}

func (t *transformer) transformMediaFeatureRange(part *ast.MediaFeatureRange) []ast.MediaQueryPart {
	if t.MediaFeatureRanges == transforms.MediaFeatureRangesPassthrough {
		return []ast.MediaQueryPart{part}
	}

	var newParts []ast.MediaQueryPart
	if part.LeftValue != nil {
		direction := "min"
		if part.Operator == ">=" {
			direction = "max"
		}

		if part.Operator == "<" || part.Operator == ">" {
			t.addWarn(part.Location(), "< and > not yet supported for transformation")
		}

		newParts = append(newParts, &ast.MediaFeaturePlain{
			// XXX: replace this allocation with a lookup.
			Property: &ast.Identifier{Value: fmt.Sprintf("%s-%s", direction, part.Property.Value)},
			Value:    part.LeftValue,
		})
	}

	if part.RightValue != nil {
		direction := "max"
		if part.Operator == ">=" {
			direction = "min"
		}

		if part.Operator == "<" || part.Operator == ">" {
			t.addWarn(part.Location(), "< and > not yet supported for transformation")
		}

		newParts = append(newParts, &ast.MediaFeaturePlain{
			// XXX: replace this allocation with a lookup.
			Property: &ast.Identifier{Value: fmt.Sprintf("%s-%s", direction, part.Property.Value)},
			Value:    part.RightValue,
		})
	}

	return newParts
}

func (t *transformer) transformMediaQueryParts(parts []ast.MediaQueryPart) []ast.MediaQueryPart {
	newParts := make([]ast.MediaQueryPart, 0, len(parts))
	for _, p := range parts {
		switch part := p.(type) {
		case *ast.MediaFeaturePlain:
			if part.Value != nil || !strings.HasPrefix(part.Property.Value, "--") {
				newParts = append(newParts, p)
				break
			}

			replacement, ok := t.customMedia[part.Property.Value]
			if !ok {
				newParts = append(newParts, p)
				break
			}

			newParts = append(newParts, replacement.Parts...)
		case *ast.MediaFeatureRange:
			newParts = append(newParts, t.transformMediaFeatureRange(part)...)

		default:
			newParts = append(newParts, p)
		}
	}
	return newParts
}

func (t *transformer) transformBlock(block ast.Block) ast.Block {
	switch node := block.(type) {
	case *ast.QualifiedRuleBlock:
		// 	for _, d := range node.Rules {
		// 		// t.transform(d)
		// 	}
		if len(node.Rules) == 0 {
			return nil
		}

	case *ast.DeclarationBlock:
		node.Declarations = t.transformDeclarations(node.Declarations)
		if len(node.Declarations) == 0 {
			return nil
		}

	default:
		panic("unknown block")
	}

	return block
}

func (t *transformer) transformDeclarations(decls []*ast.Declaration) []*ast.Declaration {
	newDecls := make([]*ast.Declaration, 0, len(decls))
	for _, d := range decls {
		d.Values = t.transformValues(d.Values)
		newDecls = append(newDecls, d)
	}

	return newDecls
}

func (t *transformer) transformValues(values []ast.Value) []ast.Value {
	rv := make([]ast.Value, 0, len(values))
	for _, value := range values {
		switch v := value.(type) {
		case *ast.Function:
			newValues := []ast.Value{v}
			func() {
				if v.Name != "var" {
					return
				}

				if t.variables == nil {
					return
				}

				if len(v.Arguments) == 0 {
					t.addError(v.Location(), "expected at least one argument to var()")
					return
				}

				varName, ok := v.Arguments[0].(*ast.Identifier)
				if !ok {
					t.addError(v.Location(), "expected identifier as argument to var()")
					return
				}

				vals, ok := t.variables[varName.Value]
				if !ok {
					// The first argument is the value, the second is a comma.
					if len(v.Arguments) > 2 {
						newValues = v.Arguments[2:]
						return
					}

					t.addWarn(v.Location(), "use of undefined variable without fallback: %s", varName.Value)
					return
				}

				newValues = vals
			}()

			rv = append(rv, newValues...)

		default:
			rv = append(rv, v)
		}
	}

	return rv
}
