package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

func (p *parser) parseSelectorList() *ast.SelectorList {
	l := &ast.SelectorList{
		Span: p.lexer.StartSpan(),
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
		Span: p.lexer.StartSpan(),
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
			s.Parts = append(s.Parts, &ast.Whitespace{Span: p.lexer.StartSpan()})
			p.lexer.Next()

		case lexer.Ident:
			s.Parts = append(s.Parts, &ast.TypeSelector{
				Span: p.lexer.StartSpan(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Hash:
			s.Parts = append(s.Parts, &ast.IDSelector{
				Span: p.lexer.StartSpan(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Delim:
			switch p.lexer.CurrentString {
			case ".":
				p.lexer.Next()
				s.Parts = append(s.Parts, &ast.ClassSelector{
					Span: p.lexer.StartSpan(),
					Name: p.lexer.CurrentString,
				})
				p.lexer.Expect(lexer.Ident)

			case "+", ">", "~", "|":
				s.Parts = append(s.Parts, &ast.CombinatorSelector{
					Span:     p.lexer.StartSpan(),
					Operator: p.lexer.CurrentString,
				})
				p.lexer.Next()

			case "*":
				s.Parts = append(s.Parts, &ast.TypeSelector{
					Span: p.lexer.StartSpan(),
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
			var wrapperLocation ast.Span
			if p.lexer.Current == lexer.Colon {
				wrapper = true
				wrapperLocation = p.lexer.StartSpan()
				p.lexer.Next()
			}

			pc := &ast.PseudoClassSelector{
				Span: p.lexer.StartSpan(),
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
							Span:  p.lexer.StartSpan(),
							Value: p.lexer.CurrentString,
						}
						p.lexer.Next()
					}
					p.lexer.Expect(lexer.RParen)

				} else {
					pc.Arguments = p.parseSelectorList()
					p.lexer.Expect(lexer.RParen)
				}

			default:
				p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
			}

			if wrapper {
				s.Parts = append(s.Parts, &ast.PseudoElementSelector{
					Span:  wrapperLocation,
					Inner: pc,
				})
				break
			}

			s.Parts = append(s.Parts, pc)

		case lexer.LBracket:
			p.lexer.Next()

			attr := &ast.AttributeSelector{
				Span:     p.lexer.StartSpan(),
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

				attr.Value = p.parseValue()
				s.Parts = append(s.Parts, attr)
			}

			p.lexer.Expect(lexer.RBracket)

		default:
			if len(s.Parts) == 0 {
				p.lexer.Errorf("expected selector")
			}
			return s
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
		Span: p.lexer.StartSpan(),
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
