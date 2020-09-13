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
	out, err := printer.Print(transformer.Transform(ast, transformer.Options{
		OriginalSource: source,
		Reporter:       &reporter{},
	}), printer.Options{})
	require.NoError(t, err)
	return out
}

type reporter struct{}

func (r *reporter) AddError(err error) {
	panic(err)
}
