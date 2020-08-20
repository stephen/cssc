package transformer

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stretchr/testify/require"
)

func BenchmarkTransformer(b *testing.B) {
	b.ReportAllocs()

	by, err := ioutil.ReadFile("../testdata/bootstrap.css")
	require.NoError(b, err)
	source := &lexer.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}
	s := parser.Parse(source)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Transform(s)
	}
}
