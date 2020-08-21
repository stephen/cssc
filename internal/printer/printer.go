package printer

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/stephen/cssc/internal/ast"
)

type printer struct {
	s strings.Builder
}

// Print prints the input AST node into CSS. It should have deterministic
// output.
func Print(in ast.Node) string {
	p := printer{}

	p.print(in)

	return p.s.String()
}

// print prints the current ast node to the printer output.
func (p *printer) print(in ast.Node) {
	switch node := in.(type) {
	case *ast.Stylesheet:
		for _, n := range node.Nodes {
			p.print(n)
		}

	case *ast.AtRule:
		p.s.WriteRune('@')
		p.s.WriteString(node.Name)
		if len(node.Preludes) > 0 {
			p.s.WriteRune(' ')
			for i, prelude := range node.Preludes {
				p.print(prelude)

				if i+1 < len(node.Preludes) {
					p.s.WriteRune(' ')
				}
			}
		}

		if node.Block != nil {
			p.s.WriteRune('{')
			p.print(node.Block)
			p.s.WriteRune('}')
		} else {
			p.s.WriteRune(';')
		}

	case *ast.SelectorList:
		for i, s := range node.Selectors {
			p.print(s)

			if i+1 < len(node.Selectors) {
				p.s.WriteRune(',')
			}
		}

	case *ast.KeyframeSelectorList:
		for _, s := range node.Selectors {
			p.print(s)
		}

	case *ast.QualifiedRule:
		p.print(node.Prelude)

		p.s.WriteRune('{')
		p.print(node.Block)
		p.s.WriteRune('}')

	case *ast.QualifiedRuleBlock:
		for _, r := range node.Rules {
			p.print(r)
		}

	case *ast.DeclarationBlock:
		for i, d := range node.Declarations {
			p.print(d)

			if i+1 < len(node.Declarations) {
				p.s.WriteRune(';')
			}
		}

	case *ast.Declaration:
		p.s.WriteString(node.Property)
		p.s.WriteRune(':')
		for i, val := range node.Values {
			p.print(val)

			// Print space if we're not the last value and the previous or current
			// value was not a comma.
			if i+1 < len(node.Values) {
				if _, nextIsComma := node.Values[i+1].(*ast.Comma); !nextIsComma {
					if _, isComma := val.(*ast.Comma); !isComma {
						p.s.WriteRune(' ')
					}
				}
			}
		}

		if node.Important {
			p.s.WriteString("!important")
		}

	case *ast.Comma:
		p.s.WriteRune(',')

	case *ast.Dimension:
		p.s.WriteString(node.Value)
		p.s.WriteString(node.Unit)

	case *ast.Percentage:
		p.s.WriteString(node.Value)
		p.s.WriteRune('%')

	case *ast.Number:
		p.s.WriteString(node.Value)

	case *ast.String:
		p.s.WriteRune('"')
		p.s.WriteString(node.Value)
		p.s.WriteRune('"')

	case *ast.Identifier:
		p.s.WriteString(node.Value)

	case *ast.Function:
		p.s.WriteString(node.Name)
		p.s.WriteRune('(')
		for _, arg := range node.Arguments {
			p.print(arg)
		}
		p.s.WriteRune(')')

	case *ast.Comment:
		p.s.WriteString("/*")
		p.s.WriteString(node.Text)
		p.s.WriteString("*/")

	case *ast.MathOperator:
		p.s.WriteString(node.Operator)

	case *ast.Whitespace:
		p.s.WriteRune(' ')

	case *ast.Selector:
		for i, part := range node.Parts {
			if _, isWhitespace := part.(*ast.Whitespace); i+1 >= len(node.Parts) && isWhitespace {
				continue
			}

			p.print(part)
		}

	case *ast.AttributeSelector:
		p.s.WriteRune('[')
		p.s.WriteString(node.Property)
		if node.Value != nil {
			p.s.WriteRune('=')
			p.print(node.Value)
		}
		p.s.WriteRune(']')

	case *ast.TypeSelector:
		p.s.WriteString(node.Name)

	case *ast.ClassSelector:
		p.s.WriteRune('.')
		p.s.WriteString(node.Name)

	case *ast.IDSelector:
		p.s.WriteRune('#')
		p.s.WriteString(node.Name)

	case *ast.CombinatorSelector:
		p.s.WriteString(node.Operator)

	case *ast.PseudoElementSelector:
		p.s.WriteRune(':')
		p.print(node.Inner)

	case *ast.HexColor:
		p.s.WriteRune('#')
		p.s.WriteString(node.RGBA)

	case *ast.PseudoClassSelector:
		p.s.WriteRune(':')
		p.s.WriteString(node.Name)
		if node.Arguments != nil {
			p.s.WriteRune('(')
			p.print(node.Arguments)
			p.s.WriteRune(')')
		}

	case *ast.ANPlusB:
		if node.A != "" {
			p.s.WriteString(node.A)
		}
		p.s.WriteRune('n')
		if node.B != "" {
			p.s.WriteRune('+')
			p.s.WriteString(node.B)
		}

	case *ast.MediaQueryList:
		for i, q := range node.Queries {
			p.print(q)

			if i+1 < len(node.Queries) {
				p.s.WriteRune(',')
			}
		}

	case *ast.MediaQuery:
		for i, part := range node.Parts {
			p.print(part)

			if i+1 < len(node.Parts) {
				p.s.WriteRune(' ')
			}
		}

	case *ast.MediaFeaturePlain:
		p.s.WriteRune('(')
		p.print(node.Property)
		if node.Value != nil {
			p.s.WriteRune(':')
			p.print(node.Value)
		}
		p.s.WriteRune(')')

	case *ast.MediaFeatureRange:
		p.s.WriteRune('(')
		if node.LeftValue != nil {
			p.print(node.LeftValue)
			p.s.WriteString(node.Operator)
		}
		p.print(node.Property)
		if node.RightValue != nil {
			p.s.WriteString(node.Operator)
			p.print(node.RightValue)
		}
		p.s.WriteRune(')')

	default:
		panic(fmt.Sprintf("unknown ast node: %s", reflect.TypeOf(in).String()))
	}

}
