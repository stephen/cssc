package parser

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stretchr/testify/require"
)

func BenchmarkParser(b *testing.B) {
	b.ReportAllocs()
	// source := `@import "test.css";
	// @import url("./testing.css");
	// @import url(tester.css);
	// /* some notes about the next line
	// are here */

	// .class {}
	// #id {}
	// body#id {}
	// body::after {}
	// a:hover {}
	// :not(a, b, c) {}
	// .one, .two {}
	// `

	by, err := ioutil.ReadFile("testdata/bootstrap.css")
	require.NoError(b, err)
	source := &lexer.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Parse(source)
	}
}
