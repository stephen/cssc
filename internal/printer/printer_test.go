package printer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Print(t testing.TB, s string) string {
	ss, err := parser.Parse(&sources.Source{
		Path:    "main.css",
		Content: s,
	})
	require.NoError(t, err)

	out, err := printer.Print(ss, printer.Options{})
	require.NoError(t, err)

	return out
}

func TestClass(t *testing.T) {
	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif}`,
		Print(t, `.class {
		font-family: "Helvetica", sans-serif;
	}`))
}

func TestClass_MultipleDeclarations(t *testing.T) {
	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif;width:2rem}`,
		Print(t, `.class {
		font-family: "Helvetica", sans-serif;
		width: 2rem;
	}`))
}

func TestClass_ComplexSelector(t *testing.T) {
	assert.Equal(t, `div.test #thing,div.test#thing,div .test#thing{}`,
		Print(t, `div.test #thing, div.test#thing, div .test#thing { }`))
}

func TestMediaQueryRanges(t *testing.T) {
	assert.Equal(t, `@media (200px<width<600px),(200px<width),(width<600px){}`,
		Print(t, `@media (200px < width < 600px), (200px < width), (width < 600px) {}`))
}

func TestKeyframes(t *testing.T) {
	assert.Equal(t, `@keyframes x{from{opacity:0}to{opacity:1}}`,
		Print(t, `@keyframes x { from { opacity: 0 } to { opacity: 1 } }`))
}

func TestRule_NoSemicolon(t *testing.T) {
	assert.Equal(t, `.class{width:2rem}`,
		Print(t, `.class { width: 2rem }`))
}

func TestANPlusB(t *testing.T) {
	assert.Equal(t, `:nth-child(n+3){}`, Print(t, `:nth-child(n + 3) {}`))
	assert.Equal(t, `:nth-child(n+3){}`, Print(t, `:nth-child(n+ 3) {}`))
	assert.Equal(t, `:nth-child(n+3){}`, Print(t, `:nth-child(n+3) {}`))

	assert.Equal(t, `:nth-child(2n+3){}`, Print(t, `:nth-child(2n + 3) {}`))
	assert.Equal(t, `:nth-child(2n+3){}`, Print(t, `:nth-child(2n +3) {}`))

	assert.Equal(t, `:nth-child(-n+6){}`, Print(t, `:nth-child(-n + 6) {}`))
	assert.Equal(t, `:nth-child(-2n+6){}`, Print(t, `:nth-child(-2n + 6) {}`))

	assert.Equal(t, `:nth-child(n){}`, Print(t, `:nth-child(n) {}`))
	assert.Equal(t, `:nth-child(-n){}`, Print(t, `:nth-child(-n) {}`))

	// From https://www.w3.org/TR/css-syntax-3/#anb-microsyntax.
	assert.Equal(t, `:nth-child(2n){}`, Print(t, `:nth-child(2n+0) {}`))
	assert.Equal(t, `:nth-child(even){}`, Print(t, `:nth-child(even) {}`))
	assert.Equal(t, `:nth-child(4n+1){}`, Print(t, `:nth-child(4n+1) {}`))
	assert.Equal(t, `:nth-child(-n+6){}`, Print(t, `:nth-child(-1n+6) {}`))
	assert.Equal(t, `:nth-child(-4n+10){}`, Print(t, `:nth-child(-4n+10) {}`))
	assert.Equal(t, `:nth-child(5){}`, Print(t, `:nth-child(0n+5) {}`))
	assert.Equal(t, `:nth-child(5){}`, Print(t, `:nth-child(5) {}`))
	assert.Equal(t, `:nth-child(n){}`, Print(t, `:nth-child(1n+0) {}`))
	assert.Equal(t, `:nth-child(n){}`, Print(t, `:nth-child(n+0) {}`))
	assert.Equal(t, `:nth-child(n){}`, Print(t, `:nth-child(n) {}`))
	assert.Equal(t, `:nth-child(2n){}`, Print(t, `:nth-child(2n+0) {}`))
	assert.Equal(t, `:nth-child(2n){}`, Print(t, `:nth-child(2n) {}`))
	assert.Equal(t, `:nth-child(3n-6){}`, Print(t, `:nth-child(3n-6) {}`))
	assert.Equal(t, `:nth-child(3n+1){}`, Print(t, `:nth-child(3n + 1) {}`))
	assert.Equal(t, `:nth-child(3n-2){}`, Print(t, `:nth-child(+3n - 2) {}`))
	assert.Equal(t, `:nth-child(-n+6){}`, Print(t, `:nth-child(-n+ 6) {}`))
	assert.Equal(t, `:nth-child(6){}`, Print(t, `:nth-child(+6) {}`))
}
