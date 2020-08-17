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

	.class {
		width: 2rem;
		margin: 2em 1px;
		height: 20%;
		padding: 0;
	}
	`,
	}

	log.Println(spew.Sdump(parser.Parse(source)))
}
