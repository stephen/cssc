package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

func (p *parser) parseSelectorList() *ast.SelectorList {
	l := &ast.SelectorList{Loc: p.lexer.Location()}
	for {
		if p.lexer.Current == lexer.EOF {
			panic("unexpected EOF")
		}

		l.Selectors = append(l.Selectors, p.parseSelector())

		if p.lexer.Current == lexer.Comma {
			p.lexer.Next()
			continue
		}

		break
	}
	return l
}

func (p *parser) parseSelector() *ast.Selector {
	s := &ast.Selector{
		Loc: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			panic("unexpected EOF")

		case lexer.Comma:
			return s

		case lexer.LCurly:
			if p.innerSelectorList {
				panic("unexpected {")
			}
			return s

		case lexer.RParen:
			if !p.innerSelectorList {
				panic("unexpected )")
			}
			return s

		case lexer.Ident:
			s.Selectors = append(s.Selectors, &ast.TypeSelector{
				Loc:  p.lexer.Location(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Hash:
			s.Selectors = append(s.Selectors, &ast.IDSelector{
				Loc:  p.lexer.Location(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Delim:
			switch p.lexer.CurrentString {
			case ".":
				p.lexer.Next()
				s.Selectors = append(s.Selectors, &ast.ClassSelector{
					Loc:  p.lexer.Location(),
					Name: p.lexer.CurrentString,
				})
				p.lexer.Expect(lexer.Ident)

			case "+", "<", "~", "|":
				s.Selectors = append(s.Selectors, &ast.CombinatorSelector{
					Loc:      p.lexer.Location(),
					Operator: p.lexer.CurrentString,
				})
				p.lexer.Next()

			case "*":
				s.Selectors = append(s.Selectors, &ast.TypeSelector{
					Loc:  p.lexer.Location(),
					Name: p.lexer.CurrentString,
				})
				p.lexer.Next()

			default:
				panic("unknown delimiter" + p.lexer.Current.String() + p.lexer.CurrentString)
			}

		case lexer.LessThan:
			s.Selectors = append(s.Selectors, &ast.CombinatorSelector{
				Loc:      p.lexer.Location(),
				Operator: "<",
			})
			p.lexer.Next()

		case lexer.Colon:
			p.lexer.Next()

			// Wrap it in a PseudoElementSelector if there are two colons.
			var wrapper *ast.Loc
			if p.lexer.Current == lexer.Colon {
				wrapper = p.lexer.Location()
				p.lexer.Next()
			}

			pc := &ast.PseudoClassSelector{
				Loc:  p.lexer.Location(),
				Name: p.lexer.CurrentString,
			}
			p.lexer.Expect(lexer.Ident)

			if p.lexer.Current == lexer.LParen {
				p.lexer.Next()

				p.innerSelectorList = true
				pc.Children = p.parseSelectorList()
				p.innerSelectorList = false
				p.lexer.Expect(lexer.RParen)
			}

			if wrapper != nil {
				s.Selectors = append(s.Selectors, &ast.PseudoElementSelector{
					Loc:   wrapper,
					Inner: pc,
				})
				break
			}

			s.Selectors = append(s.Selectors, pc)

		case lexer.LBracket:
			p.lexer.Next()
			for p.lexer.Current != lexer.RBracket {
				p.lexer.Next()
			}
			p.lexer.Next()

		default:
			panic("unknown " + p.lexer.Current.String())
		}
	}
}
