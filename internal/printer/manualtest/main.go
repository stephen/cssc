package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
)

func main() {
	source := &sources.Source{
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

	sheet, err := parser.Parse(source)
	if err != nil {
		panic(err)
	}
	log.Println(spew.Sdump(sheet))

	out, err := printer.Print(sheet, printer.Options{
		OriginalSource: source,
	})
	if err != nil {
		panic(err)
	}
	log.Println(out)

	fp := filepath.Join("internal/printer/manualtest/", "index.css")
	if err := ioutil.WriteFile(fp, []byte(out), 0644); err != nil {
		panic(err)
	}
}
