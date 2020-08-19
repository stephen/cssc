package transformer

import (
	"strings"

	"github.com/stephen/cssc/internal/ast"
)

// Transform takes a pass over the input AST and runs various
// transforms.
func Transform(s *ast.Stylesheet) *ast.Stylesheet {
	t := transformer{
		variables: make(map[string][]ast.Value),
	}

	s.Nodes = t.transformNodes(s.Nodes)

	return s
}

// transformer takes a pass over the AST and makes
// modifications to the AST, depending on the settings.
type transformer struct {
	variables map[string][]ast.Value
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

		default:
			rv = append(rv, value)
		}
	}
	return rv
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

				if len(v.Arguments) != 1 {
					// XXX: do fallbacks
					// warning: expected single argument
					return
				}

				varName, ok := v.Arguments[0].(*ast.Identifier)
				if !ok {
					// warning: expected identifier
					return
				}

				vals, ok := t.variables[varName.Value]
				if !ok {
					// warning: unknown variable
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
