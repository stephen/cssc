package printer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stretchr/testify/assert"
)

func Print(s string) string {
	return printer.Print(parser.Parse(&lexer.Source{
		Path:    "main.css",
		Content: s,
	}))
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
