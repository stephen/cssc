package transformer

import (
	"strings"

	"github.com/stephen/cssc/internal/ast"
)

// Transform takes a pass over the input AST and runs various
// transforms.
func Transform(s *ast.Stylesheet) *ast.Stylesheet {
	t := transformer{
		variables:   make(map[string][]ast.Value),
		customMedia: make(map[string]*ast.MediaQueryList),
	}

	s.Nodes = t.transformNodes(s.Nodes)

	return s
}

// transformer takes a pass over the AST and makes
// modifications to the AST, depending on the settings.
type transformer struct {
	variables   map[string][]ast.Value
	customMedia map[string]*ast.MediaQueryList
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
					if strings.HasPrefix(decl.Property, "--") {
						t.variables[decl.Property] = decl.Values
						continue
					}

					newDecls = append(newDecls, decl)
				}

				declBlock.Declarations = newDecls
			}()

			// t.transform(node.Prelude)
			node.Block = t.transformBlock(node.Block)

			if node.Block == nil {
				continue
			}
			rv = append(rv, node)

		case *ast.AtRule:
			switch node.Name {
			case "custom-media":
				func() {

					if len(node.Preludes) != 2 {
						return
					}

					name, ok := node.Preludes[0].(*ast.Identifier)
					if !ok {
						return
					}

					values, ok := node.Preludes[1].(*ast.MediaQueryList)
					if !ok {
						return
					}

					t.customMedia[name.Value] = values
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

			// XXX: uh oh. what do we do if there is a list? i think we have to go back up and duplicate the current rule.
			newParts = append(newParts, replacement.Queries[0].Parts...)

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

				if len(v.Arguments) == 0 {
					// warning: expected at least one argument
					return
				}

				varName, ok := v.Arguments[0].(*ast.Identifier)
				if !ok {
					// warning: expected identifier
					return
				}

				vals, ok := t.variables[varName.Value]
				if !ok {
					// The first argument is the value, the second is a comma.
					if len(v.Arguments) > 2 {
						newValues = v.Arguments[2:]
						return
					}

					// warning: unknown variable (and no fallback)
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
