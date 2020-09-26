package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

func (p *parser) parseSelectorList() *ast.SelectorList {
	l := &ast.SelectorList{
		Span: p.lexer.TokenSpan(),
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

	// parseSelector will always return a selector or error, so we should
	// have at least one Selector.
	l.End = l.Selectors[len(l.Selectors)-1].End

	return l
}

func (p *parser) parseSelector() *ast.Selector {
	s := &ast.Selector{
		Span: p.lexer.TokenSpan(),
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
			s.Parts = append(s.Parts, &ast.Whitespace{Span: p.lexer.TokenSpan()})
			p.lexer.Next()

		case lexer.Ident:
			s.Parts = append(s.Parts, &ast.TypeSelector{
				Span: p.lexer.TokenSpan(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Hash:
			s.Parts = append(s.Parts, &ast.IDSelector{
				Span: p.lexer.TokenSpan(),
				Name: p.lexer.CurrentString,
			})
			p.lexer.Next()

		case lexer.Delim:
			switch p.lexer.CurrentString {
			case ".":
				span := p.lexer.TokenSpan()
				p.lexer.Next()
				cls := &ast.ClassSelector{
					Span: span,
					Name: p.lexer.CurrentString,
				}
				s.Parts = append(s.Parts, cls)
				cls.End = p.lexer.TokenEnd()
				p.lexer.Expect(lexer.Ident)

			case "+", ">", "~", "|":
				s.Parts = append(s.Parts, &ast.CombinatorSelector{
					Span:     p.lexer.TokenSpan(),
					Operator: p.lexer.CurrentString,
				})
				p.lexer.Next()

			case "*":
				s.Parts = append(s.Parts, &ast.TypeSelector{
					Span: p.lexer.TokenSpan(),
					Name: p.lexer.CurrentString,
				})
				p.lexer.Next()

			default:
				p.lexer.Errorf("unexpected delimeter: %s", p.lexer.CurrentString)
			}

		case lexer.Colon:
			span := p.lexer.TokenSpan()
			p.lexer.Next()

			// Wrap it in a PseudoElementSelector if there are two colons.
			var wrapper bool
			var wrapperLocation ast.Span
			if p.lexer.Current == lexer.Colon {
				wrapper = true
				wrapperLocation = span
				span = p.lexer.TokenSpan()
				p.lexer.Next()
			}

			pc := &ast.PseudoClassSelector{
				Span: span,
				Name: p.lexer.CurrentString,
			}

			switch p.lexer.Current {
			case lexer.Ident:
				pc.End = p.lexer.TokenEnd()
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
							Span:  p.lexer.TokenSpan(),
							Value: p.lexer.CurrentString,
						}
						p.lexer.Next()
					}
				} else {
					pc.Arguments = p.parseSelectorList()
				}
				pc.End = p.lexer.TokenEnd()
				p.lexer.Expect(lexer.RParen)

			default:
				p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
			}

			if wrapper {
				wrapped := &ast.PseudoElementSelector{
					Span:  wrapperLocation,
					Inner: pc,
				}
				wrapped.End = pc.End
				s.Parts = append(s.Parts, wrapped)
				break
			}

			s.Parts = append(s.Parts, pc)

		case lexer.LBracket:
			startLoc := p.lexer.TokenSpan()
			p.lexer.Next()
			attr := &ast.AttributeSelector{
				Span:     startLoc,
				Property: p.lexer.CurrentString,
			}

			p.lexer.Expect(lexer.Ident)
			if p.lexer.Current == lexer.RBracket {
				attr.End = p.lexer.TokenEnd()
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
				if attr.Value == nil {
					p.lexer.Errorf("value must be specified")
				}
				s.Parts = append(s.Parts, attr)
			}

			attr.End = p.lexer.TokenEnd()
			p.lexer.Expect(lexer.RBracket)

		default:
			if len(s.Parts) == 0 {
				p.lexer.Errorf("expected selector")
			}
			s.End = s.Parts[len(s.Parts)-1].Location().End
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

	v := &ast.ANPlusB{Span: p.lexer.TokenSpan()}

	if p.lexer.Current == lexer.Number {
		v.A = p.lexer.CurrentNumeral
		p.lexer.Next()
	}

	if p.lexer.CurrentString != "n" {
		p.lexer.Errorf("expected literal n as part of An+B")
	}
	v.End = p.lexer.TokenEnd()
	p.lexer.Expect(lexer.Ident)

	if p.lexer.Current == lexer.Delim && (p.lexer.CurrentString == "+" || p.lexer.CurrentString == "-") {
		v.Operator = p.lexer.CurrentString
		p.lexer.Next()

		v.B = p.lexer.CurrentNumeral
		v.End = p.lexer.TokenEnd()
		p.lexer.Expect(lexer.Number)
	}
	return v
}
