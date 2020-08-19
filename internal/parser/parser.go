package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

// Parse parses an input stylesheet.
func Parse(source *lexer.Source) *ast.Stylesheet {
	p := newParser(source)
	p.parse()
	return p.ss
}

func newParser(source *lexer.Source) *parser {
	return &parser{
		lexer: lexer.NewLexer(source),
		ss:    &ast.Stylesheet{},
	}
}

type parser struct {
	lexer *lexer.Lexer
	ss    *ast.Stylesheet

	// Flags. These flags are only on the parser because they must
	// travel across function boundaries. Other single-depth flags
	// should be passed as arguments locally.

	// innerSelectorList is true if we're currently parsing a nested
	// selector list, e.g. :not(a, b, c).
	innerSelectorList bool
}

func (p *parser) parse() {
	for p.lexer.Current != lexer.EOF {
		switch p.lexer.Current {
		case lexer.At:
			p.parseAtRule()

		case lexer.CDO, lexer.CDC:
			// From https://www.w3.org/TR/css-syntax-3/#parser-entry-points,
			// we'll always assume we're parsing from the top-level, so we can discard CDO/CDC.
			p.lexer.Next()

		case lexer.Comment:
			p.ss.Nodes = append(p.ss.Nodes, &ast.Comment{
				Loc:  p.lexer.Location(),
				Text: p.lexer.CurrentString,
			})
			p.lexer.Next()

		default:
			p.ss.Nodes = append(p.ss.Nodes, p.parseQualifiedRule(false))
		}

	}
}

func isImportantString(in string) bool {
	return len(in) == 9 &&
		(in[0] == 'i' || in[0] == 'I') &&
		(in[1] == 'm' || in[1] == 'M') &&
		(in[2] == 'p' || in[2] == 'P') &&
		(in[3] == 'o' || in[3] == 'O') &&
		(in[4] == 'r' || in[4] == 'R') &&
		(in[5] == 't' || in[5] == 'T') &&
		(in[6] == 'a' || in[6] == 'A') &&
		(in[7] == 'n' || in[7] == 'N') &&
		(in[8] == 't' || in[8] == 'T')
}

// parseQualifiedRule parses a rule. If isKeyframes is set, the parser will assume
// all preludes are keyframes percentage selectors. Otherwise, it will assume
// the preludes are selector lists.
func (p *parser) parseQualifiedRule(isKeyframes bool) *ast.QualifiedRule {
	r := &ast.QualifiedRule{
		Loc: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")
		case lexer.LCurly:
			block := &ast.DeclarationBlock{
				Loc: p.lexer.Location(),
			}

			r.Block = block
			p.lexer.Next()

			for p.lexer.Current != lexer.RCurly {
				decl := &ast.Declaration{
					Loc:      p.lexer.Location(),
					Property: p.lexer.CurrentString,
				}
				p.lexer.Expect(lexer.Ident)
				p.lexer.Expect(lexer.Colon)
			values:
				for {
					switch p.lexer.Current {
					case lexer.EOF:
						p.lexer.Errorf("unexpected EOF")
					case lexer.Semicolon:
						// XXX: if no values, get upset.
						p.lexer.Next()
						block.Declarations = append(block.Declarations, decl)

						break values
					case lexer.Delim:
						if p.lexer.CurrentString != "!" {
							p.lexer.Errorf("unexpected token: %s", p.lexer.CurrentString)
						}
						p.lexer.Next()

						if !isImportantString(p.lexer.CurrentString) {
							p.lexer.Errorf("expected !important, unexpected token: %s", p.lexer.CurrentString)
						}
						p.lexer.Next()
						decl.Important = true

					case lexer.Comma:
						decl.Values = append(decl.Values, &ast.Comma{Loc: p.lexer.Location()})
						p.lexer.Next()

					default:
						decl.Values = append(decl.Values, p.parseValue(false))
					}
				}
			}
			p.lexer.Next()

			return r
		default:
			if isKeyframes {
				r.Prelude = p.parseKeyframeSelectorList()
				continue
			}

			r.Prelude = p.parseSelectorList()
		}
	}
}

func (p *parser) parseKeyframeSelectorList() *ast.KeyframeSelectorList {
	l := &ast.KeyframeSelectorList{
		Loc: p.lexer.Location(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
		}

		switch p.lexer.Current {
		case lexer.Percentage:
			l.Selectors = append(l.Selectors, &ast.Percentage{
				Loc:   p.lexer.Location(),
				Value: p.lexer.CurrentNumeral,
			})

		case lexer.Ident:
			if p.lexer.CurrentString != "from" && p.lexer.CurrentString != "to" {
				p.lexer.Errorf("unexpected string: %s. keyframe selector can only be from, to, or a percentage", p.lexer.CurrentString)
			}
			l.Selectors = append(l.Selectors, &ast.Identifier{
				Loc:   p.lexer.Location(),
				Value: p.lexer.CurrentString,
			})

		default:
			p.lexer.Errorf("unexepected token: %s. keyframe selector can only be from, to, or a percentage", p.lexer.Current.String())
		}
		p.lexer.Next()

		if p.lexer.Current == lexer.Comma {
			p.lexer.Next()
			continue
		}

		break
	}

	return l
}

// parseValue parses a possible ast value at the current position. Callers
// can set allowMathOperators if the enclosing context allows math expressions.
// See: https://www.w3.org/TR/css-values-4/#math-function.
func (p *parser) parseValue(allowMathOperators bool) ast.Value {
	switch p.lexer.Current {
	case lexer.Dimension:
		defer p.lexer.Next()
		return &ast.Dimension{
			Loc: p.lexer.Location(),

			Unit:  p.lexer.CurrentString,
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Percentage:
		defer p.lexer.Next()
		return &ast.Percentage{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Number:
		defer p.lexer.Next()
		// XXX: should we make sure this is 0?
		return &ast.Number{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Ident:
		defer p.lexer.Next()
		return &ast.Identifier{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Hash:
		defer p.lexer.Next()
		return &ast.HexColor{
			Loc:  p.lexer.Location(),
			RGBA: p.lexer.CurrentString,
		}

	case lexer.String:
		defer p.lexer.Next()
		return &ast.String{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Delim:
		switch p.lexer.CurrentString {
		case "*", "/", "+", "-":
			if !allowMathOperators {
				p.lexer.Errorf("math operations are only allowed within: calc, min, max, or clamp")
				return nil
			}
			defer p.lexer.Next()

			return &ast.MathOperator{
				Loc:      p.lexer.Location(),
				Operator: p.lexer.CurrentString,
			}

		default:
			p.lexer.Errorf("unexpected token: %s", p.lexer.CurrentString)
			return nil
		}

	case lexer.FunctionStart:
		fn := &ast.Function{
			Loc:  p.lexer.Location(),
			Name: p.lexer.CurrentString,
		}
		p.lexer.Next()

	arguments:
		for {
			switch p.lexer.Current {
			case lexer.RParen:
				p.lexer.Next()
				break arguments
			case lexer.Comma:
				p.lexer.Next()
			default:
				fn.Arguments = append(fn.Arguments, p.parseValue(fn.IsMath()))
			}
		}

		return fn
	default:
		p.lexer.Errorf("unknown token: %s|%s|%s", p.lexer.Current, p.lexer.CurrentString, p.lexer.CurrentNumeral)
		return nil
	}
}

func (p *parser) parseAtRule() {
	switch p.lexer.CurrentString {
	case "import":
		p.parseImportAtRule()

	case "media":
		p.parseMediaAtRule()

	case "keyframes", "-webkit-keyframes":
		p.parseKeyframes()

	default:
		p.lexer.Errorf("unsupported at rule: %s", p.lexer.CurrentString)
	}
}

// parseImportAtRule parses an import at rule. It roughly implements
// https://www.w3.org/TR/css-cascade-4/#at-import.
func (p *parser) parseImportAtRule() {
	prelude := &ast.String{
		Loc: p.lexer.Location(),
	}

	imp := &ast.AtRule{
		Loc:     p.lexer.Location(),
		Name:    p.lexer.CurrentString,
		Prelude: prelude,
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.URL:
		prelude.Value = p.lexer.CurrentString
		p.lexer.Next()

	case lexer.FunctionStart:
		if p.lexer.CurrentString != "url" {
			p.lexer.Errorf("@import target must be a url or string")
		}
		p.lexer.Next()

		prelude.Value = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)
		p.lexer.Expect(lexer.RParen)

	case lexer.String:
		prelude.Value = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)

	default:
		p.lexer.Errorf("unexpected import specifier")
	}

	p.lexer.Expect(lexer.Semicolon)

	// XXX: support conditional @import

	p.ss.Nodes = append(p.ss.Nodes, imp)
}

// parseKeyframes parses a keyframes at rule. It roughly implements
// https://www.w3.org/TR/css-animations-1/#keyframes
func (p *parser) parseKeyframes() {
	r := &ast.AtRule{
		Loc:  p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.String:
		r.Prelude = &ast.String{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Ident:
		r.Prelude = &ast.Identifier{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	default:
		p.lexer.Errorf("unexpected token %s, expected string or identifier for keyframes", p.lexer.Current.String())
	}
	p.lexer.Next()

	block := &ast.QualifiedRuleBlock{
		Loc: p.lexer.Location(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(true))
		}
	}
}

// parseMediaAtRule parses a media at rule. It roughly implements
// https://www.w3.org/TR/mediaqueries-4/#media.
func (p *parser) parseMediaAtRule() {
	r := &ast.AtRule{
		Loc:  p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	for p.lexer.Current != lexer.Semicolon && p.lexer.Current != lexer.LCurly {
		p.lexer.Next()
	}
	// XXX: actually parse media query.

	block := &ast.QualifiedRuleBlock{
		Loc: p.lexer.Location(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(false))
		}
	}
}
