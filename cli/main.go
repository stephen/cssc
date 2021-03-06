package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
)

func main() {
	source := &sources.Source{
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
		color: rgb(255, 255, calc(2 + 2));
	}

	section {
		float: left;
		margin: 1em; border: solid 1px;
		width: calc(100%/3 - 2*1em - 2*1px);
	}

	/*
		here we are:
	*/
	section .child {}
	section.self {}

	[test="hello"] {}
	[test=hello] {}
	[test*=hello] {}
	[test^=2.5] {}
	[test] {}

	@media (width: 600px), (200px < width < 600px), (200px < width), (width < 600px) {
		.c {height: 100%;}
	}

	@media not screen {
		.c {height: 100%;}
	}

	@media screen and (color), projection and (color) {
		.c {height: 100%;}
	}

	@media not (width <= -100px) {
		body { background: green; }
	}
	@media (min-width: 30em) and (orientation: landscape) {
		body { background: green; }
	}

	@import url('landscape.css') screen and (orientation: landscape);
	`,
	}

	sheet, err := parser.Parse(source)
	if err != nil {
		panic(err)
	}

	log.Println(spew.Sdump(sheet))
	log.Println(printer.Print(sheet, printer.Options{
		OriginalSource: source,
	}))
}
