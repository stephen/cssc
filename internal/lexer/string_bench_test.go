package lexer

import (
	"strings"
	"testing"
)

func BenchmarkStringComparison(b *testing.B) {
	in := "URL"
	b.Run("strings.ToLower", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = strings.ToLower(in) == "url"
		}
	})

	b.Run("isURLString", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			isURLString(in)
		}
	})
}
