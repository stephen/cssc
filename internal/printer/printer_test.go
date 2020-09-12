package printer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/assert"
)

func Print(s string) string {
	return printer.Print(parser.Parse(&sources.Source{
		Path:    "main.css",
		Content: s,
	}), printer.Options{})
}

func TestClass(t *testing.T) {
	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif}`,
		Print(`.class {
		font-family: "Helvetica", sans-serif;
	}`))
}

func TestClass_MultipleDeclarations(t *testing.T) {
	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif;width:2rem}`,
		Print(`.class {
		font-family: "Helvetica", sans-serif;
		width: 2rem;
	}`))
}

func TestClass_ComplexSelector(t *testing.T) {
	assert.Equal(t, `div.test #thing,div.test#thing,div .test#thing{}`,
		Print(`div.test #thing, div.test#thing, div .test#thing { }`))
}

func TestMediaQueryRanges(t *testing.T) {
	assert.Equal(t, `@media (200px<width<600px),(200px<width),(width<600px){}`,
		Print(`@media (200px < width < 600px), (200px < width), (width < 600px) {}`))
}

func TestKeyframes(t *testing.T) {
	assert.Equal(t, `@keyframes x{from{opacity:0}to{opacity:1}}`,
		Print(`@keyframes x { from { opacity: 0 } to { opacity: 1 } }`))
}

func TestRule_NoSemicolon(t *testing.T) {
	assert.Equal(t, `.class{width:2rem}`,
		Print(`.class { width: 2rem }`))
}
