package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func Transform(s string) string {
	return printer.Print(transformer.Transform(parser.Parse(&lexer.Source{
		Path:    "main.css",
		Content: s,
	})))
}

func TestCustomProperties(t *testing.T) {
	assert.Equal(t, ".class{margin:0rem 1rem 3rem 5rem}", Transform(`:root {
	--var-width: 1rem 3rem 5rem;
}

.class {
	margin: 0rem var(--var-width);
}`))

	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif,other}`, Transform(`:root {
		--font: "Helvetica", sans-serif, other;
}

.class {
	font-family: var(--font);
}`))

}

func TestCustomProperties_Fallback(t *testing.T) {
	assert.Equal(t, ".class{margin:0rem 2rem}", Transform(`.class {
	margin: 0rem var(--var-width, 2rem);
}`))

	assert.Equal(t, ".class{margin:0rem 2rem 1rem 3rem}", Transform(`.class {
	margin: 0rem var(--var-width, 2rem 1rem 3rem);
}`))

	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif}`, Transform(`.class {
		font-family: var(--font, "Helvetica", sans-serif);
	}`))
}
