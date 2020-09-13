package transformer

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkTransformer(b *testing.B) {
	b.ReportAllocs()

	by, err := ioutil.ReadFile("../testdata/bootstrap.css")
	require.NoError(b, err)
	source := &sources.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}
	s, err := parser.Parse(source)
	assert.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Transform(s, Options{
			OriginalSource: source,
		})
	}
}
