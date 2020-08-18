package printer

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stretchr/testify/require"
)

func BenchmarkPrinter(b *testing.B) {
	b.ReportAllocs()

	by, err := ioutil.ReadFile("../testdata/bootstrap.css")
	require.NoError(b, err)
	source := &lexer.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}
	ast := parser.Parse(source)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Print(ast)
	}
}
