package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

func (p *parser) parseSelectorList() *ast.SelectorList {
	l := &ast.SelectorList{
		Loc: p.lexer.Location(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
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

	prevRetainWhitespace := p.lexer.RetainWhitespace
	p.lexer.RetainWhitespace = true
	defer func() {
		p.lexer.RetainWhitespace = prevRetainWhitespace
	}()

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.Whitespace:
			s.Parts = append(s.Parts, &ast.Whitespace{Loc: p.lexer.Location()})
			p.lexer.Next()

		case lexer.Comma:
			return s

		case lexer.LCurly:
			if p.innerSelectorList {
				p.lexer.Errorf("unexpected {")
			}
			return s

		case lexer.RParen:
			if !p.innerSelectorList {
				p.lexer.Errorf("unexpected )")
			}
			return s

		case lexer.Ident:
			s.Parts = append(s.Parts, &ast.TypeSelector{
				Loc:  p.lexer.Location(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Hash:
			s.Parts = append(s.Parts, &ast.IDSelector{
				Loc:  p.lexer.Location(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Delim:
			switch p.lexer.CurrentString {
			case ".":
				p.lexer.Next()
				s.Parts = append(s.Parts, &ast.ClassSelector{
					Loc:  p.lexer.Location(),
					Name: p.lexer.CurrentString,
				})
				p.lexer.Expect(lexer.Ident)

			case "+", ">", "~", "|":
				s.Parts = append(s.Parts, &ast.CombinatorSelector{
					Loc:      p.lexer.Location(),
					Operator: p.lexer.CurrentString,
				})
				p.lexer.Next()

			case "*":
				s.Parts = append(s.Parts, &ast.TypeSelector{
					Loc:  p.lexer.Location(),
					Name: p.lexer.CurrentString,
				})
				p.lexer.Next()

			default:
				p.lexer.Errorf("unexpected delimeter: %s", p.lexer.CurrentString)
			}

		case lexer.Colon:
			p.lexer.Next()

			// Wrap it in a PseudoElementSelector if there are two colons.
			var wrapper bool
			var wrapperLocation ast.Loc
			if p.lexer.Current == lexer.Colon {
				wrapper = true
				wrapperLocation = p.lexer.Location()
				p.lexer.Next()
			}

			pc := &ast.PseudoClassSelector{
				Loc:  p.lexer.Location(),
				Name: p.lexer.CurrentString,
			}

			switch p.lexer.Current {
			case lexer.Ident:
				p.lexer.Next()

			case lexer.FunctionStart:
				p.lexer.Next()

				if pc.Name == "nth-child" || pc.Name == "nth-last-child" || pc.Name == "nth-of-type" || pc.Name == "nth-last-of-type" {
					switch p.lexer.Current {
					case lexer.Number:
						pc.Arguments = p.parseANPlusB()
					case lexer.Ident:
						if p.lexer.CurrentString == "n" {
							pc.Arguments = p.parseANPlusB()
							break
						}

						if p.lexer.CurrentString != "even" && p.lexer.CurrentString != "odd" {
							p.lexer.Errorf("expected even, odd, or an+b syntax")
						}
						pc.Arguments = &ast.Identifier{
							Loc:   p.lexer.Location(),
							Value: p.lexer.CurrentString,
						}
						p.lexer.Next()
					}
					p.lexer.Expect(lexer.RParen)

				} else {
					p.innerSelectorList = true
					pc.Arguments = p.parseSelectorList()
					p.innerSelectorList = false
					p.lexer.Expect(lexer.RParen)
				}

			default:
				p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
			}

			if wrapper {
				s.Parts = append(s.Parts, &ast.PseudoElementSelector{
					Loc:   wrapperLocation,
					Inner: pc,
				})
				break
			}

			s.Parts = append(s.Parts, pc)

		case lexer.LBracket:
			p.lexer.Next()

			attr := &ast.AttributeSelector{
				Loc:      p.lexer.Location(),
				Property: p.lexer.CurrentString,
			}
			p.lexer.Expect(lexer.Ident)
			if p.lexer.Current == lexer.RBracket {
				s.Parts = append(s.Parts, attr)
				p.lexer.Next()
				break
			}

			if p.lexer.Current == lexer.Delim {
				switch p.lexer.CurrentString {
				case "^", "~", "$", "*":
					attr.PreOperator = p.lexer.CurrentString
					p.lexer.Next()

					if p.lexer.CurrentString != "=" {
						p.lexer.Errorf("expected =, got %s: ", p.lexer.CurrentString)
					}
					p.lexer.Expect(lexer.Delim)
				case "=":
					p.lexer.Expect(lexer.Delim)
				}

				attr.Value = p.parseValue(false)
				s.Parts = append(s.Parts, attr)
			}

			p.lexer.Expect(lexer.RBracket)

		default:
			p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
		}
	}
}

func (p *parser) parseANPlusB() *ast.ANPlusB {
	prev := p.lexer.RetainWhitespace
	p.lexer.RetainWhitespace = false
	defer func() {
		p.lexer.RetainWhitespace = prev
	}()
	v := &ast.ANPlusB{
		Loc: p.lexer.Location(),
	}
	if p.lexer.Current == lexer.Number {
		v.A = p.lexer.CurrentNumeral
		p.lexer.Next()
	}

	if p.lexer.CurrentString != "n" {
		p.lexer.Errorf("expected literal n as part of An+B")
	}
	p.lexer.Expect(lexer.Ident)

	if p.lexer.Current == lexer.Delim && (p.lexer.CurrentString == "+" || p.lexer.CurrentString == "-") {
		v.Operator = p.lexer.CurrentString
		p.lexer.Next()

		v.B = p.lexer.CurrentNumeral
		p.lexer.Expect(lexer.Number)
	}
	return v
}
