package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
)

func main() {
	source := &lexer.Source{
		Path: "index.css",
		Content: `
	body {
		background-color: purple;
	}

	@media not (width <= -100px) {
		body {
			background: red;
		}
	}
	@media (min-width: 30em) and (orientation: landscape) {
		body {

			background: green;}
	}

	@import url('landscape.css') screen and (orientation: landscape);

	.a {
		color: white;
	}

	.b {
		color: red;
	}
	`,
	}

	sheet := parser.Parse(source)
	log.Println(spew.Sdump(sheet))
	out := printer.Print(sheet, printer.Options{
		OriginalSource: source,
	})

	log.Println(out)

	fp := filepath.Join("internal/printer/manualtest/", "index.css")
	if err := ioutil.WriteFile(fp, []byte(out), 0644); err != nil {
		panic(err)
	}
}
