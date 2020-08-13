package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

func main() {
	source := `@import "test.css";
@import url("./testing.css");
	@import url(tester.css);
	/* some notes about the next line
	are here */

	.class {}
	#id {}
	body#id {}
	body::after {}
	a:hover {}
	:not(a, b, c) {}
	.one, .two {}
	`
	// .n {
	// 	width: yes;
	// 	height: 2.3px;
	// 	border: -2em;
	// 	content: "test\u005ctest";
	// }

	// #test {
	// 	uhoh: hello;
	// 	img: url(test.com)
	// 	other: url("test.net")
	// }

	log.Println(spew.Sdump(Parse(source)))
}

// Parse parses an input stylesheet.
func Parse(source string) *ast.Stylesheet {
	p := newParser(source)
	p.parse()
	return p.ss
}

func newParser(source string) *parser {
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
		case lexer.AtKeyword:
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
			panic("unexpected EOF")
		case lexer.LCurly:
			// Consume a simple block
			p.lexer.Next()
			for p.lexer.Current != lexer.RCurly {
				p.lexer.Next()
			}
			p.lexer.Next()

			p.ss.Nodes = append(p.ss.Nodes, r)
			return
		default:
			r.Selectors = p.parseSelectorList()
		}
	}
}

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

			default:
				panic("unknown delimiter")
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
			// attribute selector
		default:
			panic("unknown " + p.lexer.Current.String())
		}
	}
}

func (p *parser) parseAtRule() {
	switch p.lexer.CurrentString {
	case "import":
		p.lexer.Next()
		p.parseImportAtRule()
	default:
		panic("unsuppported at rule")
	}
}

// parseImportAtRule parses an at rule. It roughly implements
// https://www.w3.org/TR/css-cascade-4/#at-import.
func (p *parser) parseImportAtRule() {
	imp := &ast.ImportAtRule{
		Loc: p.lexer.Location(),
	}

	switch p.lexer.Current {
	case lexer.URL:
		imp.URL = p.lexer.CurrentString
		p.lexer.Next()

	case lexer.FunctionStart:
		if p.lexer.CurrentString != "url" {
			panic("@import target must be a url or string")
		}
		p.lexer.Next()

		imp.URL = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)
		p.lexer.Expect(lexer.RParen)

	case lexer.String:
		imp.URL = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)

	default:
		panic("unexpected import specifier")
	}

	p.lexer.Expect(lexer.Semicolon)

	// XXX: support conditional @import

	p.ss.Nodes = append(p.ss.Nodes, imp)
}
