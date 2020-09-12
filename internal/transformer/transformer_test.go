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
	source := &sources.Source{
		Path:    "main.css",
		Content: s,
	}
	ast, err := parser.Parse(source)

	require.NoError(t, err)
	return printer.Print(transformer.Transform(ast, source, transformer.WithReporter(&reporter{})), printer.Options{})
}

type reporter struct{}

func (r *reporter) AddError(err error) {
	panic(err)
}
