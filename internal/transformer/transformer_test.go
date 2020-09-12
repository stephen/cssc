package transformer_test

import (
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/internal/transformer"
)

func Transform(s string) string {
	return printer.Print(transformer.Transform(parser.Parse(&sources.Source{
		Path:    "main.css",
		Content: s,
	})), printer.Options{})
}
