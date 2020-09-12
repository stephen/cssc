package printer

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkPrinter(b *testing.B) {
	b.ReportAllocs()

	by, err := ioutil.ReadFile("../testdata/bootstrap.css")
	require.NoError(b, err)
	source := &sources.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}
	ast, err := parser.Parse(source)
	assert.NoError(b, err)

	b.Run("no sourcemap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Print(ast, Options{})
		}
	})

	b.Run("with sourcemap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Print(ast, Options{OriginalSource: source})
		}
	})
}
