package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
)

func main() {
	source := &lexer.Source{
		Content: `@import "test.css";
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
	`,
	}

	log.Println(spew.Sdump(parser.Parse(source)))
}
