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
			p.parseQualifiedRule()
		}

	}
}

func (p *parser) parseQualifiedRule() {
	r := &ast.QualifiedRule{
		Loc: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")
		case lexer.LCurly:
			// XXX: Consume a simple block
			r.Block = &ast.Block{
				Loc: p.lexer.Location(),
			}
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
						r.Block.Declarations = append(r.Block.Declarations, decl)

						break values
					default:
						decl.Values = append(decl.Values, p.parseValue(false))
					}
				}
			}
			p.lexer.Next()

			p.ss.Nodes = append(p.ss.Nodes, r)
			return
		default:
			r.Selectors = p.parseSelectorList()
		}
	}
}

var mathFunctions = map[string]struct{}{
	"calc":  struct{}{},
	"min":   struct{}{},
	"max":   struct{}{},
	"clamp": struct{}{},
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
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Number:
		defer p.lexer.Next()
		// XXX: should we make sure this is 0?
		return &ast.Number{
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

	case lexer.Delim:
		switch p.lexer.CurrentString {
		case "*", "/", "+", "-":
			if !allowMathOperators {
				p.lexer.Errorf("math operations are only allowed within: calc, min, max, or clamp")
				return nil
			}
			p.lexer.Next()

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
				_, allowMath := mathFunctions[fn.Name]
				fn.Arguments = append(fn.Arguments, p.parseValue(allowMath))
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
		p.lexer.Next()
		p.parseImportAtRule()
	case "media":
		p.lexer.Next()
		p.parseMediaAtRule()
	case "keyframes", "-webkit-keyframes":
		// XXX: maybe consolidate all at rule AST/parsing?
		p.lexer.Next()
		p.parseMediaAtRule()
	default:
		p.lexer.Errorf("unsupported at rule: %s", p.lexer.CurrentString)
	}
}

// parseImportAtRule parses an import at rule. It roughly implements
// https://www.w3.org/TR/css-cascade-4/#at-import.
func (p *parser) parseImportAtRule() {
	imp := &ast.AtRule{
		Loc:  p.lexer.Location(),
		Name: "import",
	}

	prelude := &ast.ImportPrelude{}

	switch p.lexer.Current {
	case lexer.URL:
		prelude.URL = p.lexer.CurrentString
		p.lexer.Next()

	case lexer.FunctionStart:
		if p.lexer.CurrentString != "url" {
			p.lexer.Errorf("@import target must be a url or string")
		}
		p.lexer.Next()

		prelude.URL = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)
		p.lexer.Expect(lexer.RParen)

	case lexer.String:
		prelude.URL = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)

	default:
		p.lexer.Errorf("unexpected import specifier")
	}

	p.lexer.Expect(lexer.Semicolon)

	// XXX: support conditional @import

	p.ss.Nodes = append(p.ss.Nodes, imp)
}

// parseMediaAtRule parses a media at rule. It roughly implements
// https://www.w3.org/TR/mediaqueries-4/#media.
func (p *parser) parseMediaAtRule() {
	imp := &ast.AtRule{
		Loc:  p.lexer.Location(),
		Name: "media",
	}

	p.lexer.Next()
	for p.lexer.Current != lexer.Semicolon && p.lexer.Current != lexer.LCurly {
		p.lexer.Next()
	}

	// XXX: actually parse media query and inner block.
	if p.lexer.Current == lexer.LCurly {
		p.lexer.Next()
		inner := 0
	skip:
		for {
			switch p.lexer.Current {
			case lexer.LCurly:
				inner++

			case lexer.RCurly:
				if inner == 0 {
					break skip
				}
				inner--
			}
			p.lexer.Next()
		}
	}
	p.lexer.Next()

	p.ss.Nodes = append(p.ss.Nodes, imp)
}
