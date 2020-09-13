package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/internal/transformer"
	"github.com/stretchr/testify/require"
)

func Transform(t testing.TB, modifier func(o *transformer.Options), s string) string {
	source := &sources.Source{
		Path:    "main.css",
		Content: s,
	}
	ast, err := parser.Parse(source)

	require.NoError(t, err)
	o := &transformer.Options{
		OriginalSource: source,
		Reporter:       &reporter{},
	}

	if modifier != nil {
		modifier(o)
	}

	out, err := printer.Print(transformer.Transform(ast, *o), printer.Options{})
	require.NoError(t, err)
	return out
}

type reporter struct{}

func (r *reporter) AddError(err error) {
	panic(err)
}
