package main

import (
	"fmt"
	"io/ioutil"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
)

func main() {
	by, err := ioutil.ReadFile("internal/testdata/bootstrap.css")
	if err != nil {
		panic(err)
	}

	source := &lexer.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}

	fmt.Println(printer.Print(parser.Parse(source)))
}
