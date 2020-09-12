package parser

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/require"
)

func BenchmarkParser(b *testing.B) {
	b.ReportAllocs()

	by, err := ioutil.ReadFile("../testdata/bootstrap.css")
	require.NoError(b, err)
	source := &sources.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Parse(source)
	}
}
