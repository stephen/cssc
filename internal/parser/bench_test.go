package parser

import "testing"

func BenchmarkParser(b *testing.B) {
	b.ReportAllocs()
	source := `@import "test.css";
	@import url("./testing.css");
	@import url(tester.css);
	/* some notes about the next line
	are here */

	.class {}
	#id {}
	body#id {}
	body::after {}
	a:hover {}
	:not(a, b, c) {}
	.one, .two {}
	`

	for i := 0; i < b.N; i++ {
		Parse(source)
	}
}
