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
		if node.Prelude != nil {
			p.s.WriteRune(' ')
			p.print(node.Prelude)
		}

		if node.Block != nil {
			p.s.WriteRune('{')
			p.print(node.Block)
			p.s.WriteRune('}')
		} else {
			p.s.WriteRune(';')
		}

	case *ast.ImportPrelude:
		p.s.WriteRune('"')
		p.s.WriteString(node.URL)
		p.s.WriteRune('"')

	case *ast.QualifiedRule:
		for _, s := range node.Selectors {
			p.print(s)
		}

		p.s.WriteRune('{')
		for _, decl := range node.Block.Declarations {
			p.print(decl)
		}
		p.s.WriteRune('}')

	case *ast.Declaration:
		p.s.WriteString(node.Property)
		p.s.WriteRune(':')
		for i, val := range node.Values {
			p.print(val)

			if i+1 < len(node.Values) {
				p.s.WriteRune(' ')
			}
		}

		if node.Important {
			p.s.WriteString("!important")
		}

		p.s.WriteRune(';')

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
		for i, arg := range node.Arguments {
			p.print(arg)
			if !node.IsMath() && i+1 < len(node.Arguments) {
				p.s.WriteString(",")
			}
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
		for _, part := range node.Parts {
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
		if len(node.Children) > 0 {
			p.s.WriteRune('(')
			for _, arg := range node.Children {
				p.print(arg)
			}
			p.s.WriteRune(')')
		}

	default:
		panic(fmt.Sprintf("unknown ast node: %s", reflect.TypeOf(in).String()))
	}

}
