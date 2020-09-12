package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/internal/transformer"
	"github.com/stretchr/testify/require"
)

func Transform(t testing.TB, s string) string {
	ast, err := parser.Parse(&sources.Source{
		Path:    "main.css",
		Content: s,
	})

	require.NoError(t, err)
	return printer.Print(transformer.Transform(ast), printer.Options{})
}
