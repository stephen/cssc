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
	@import url(tester.css);`
	// /* some notes about the next line
	// are here */
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
}

func (p *parser) parse() {
	for p.lexer.Current != lexer.EOF {
		switch p.lexer.Current {
		case lexer.AtKeyword:
			p.parseAtRule()
		default:
			log.Printf("current token: %s (%s%s)", p.lexer.Current, p.lexer.CurrentNumeral, p.lexer.CurrentString)
			p.lexer.CurrentNumeral, p.lexer.CurrentString = "", ""
			p.lexer.Next()
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
		Loc: &ast.Loc{Position: p.lexer.Location()},
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

	p.ss.Rules = append(p.ss.Rules, imp)
}
