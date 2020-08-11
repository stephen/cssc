package lexer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer_URL(t *testing.T) {
	h := NewHarness(t, lexer.NewLexer("url(http://test.com/image.jpg)"))

	h.ExpectAndNext(lexer.URL, "http://test.com/image.jpg", "")
	h.ExpectAndNext(lexer.EOF, "", "")

	assert.Panics(t, func() {
		lexer.NewLexer("url(http://test.com/image.jpg").Next()
	})

	assert.Panics(t, func() {
		lexer.NewLexer("url(())").Next()
	})
}

func TestLexer_Function(t *testing.T) {
	h := NewHarness(t, lexer.NewLexer(`url("http://test.com/image.jpg")`))

	h.ExpectAndNext(lexer.FunctionStart, "url", "")
	h.ExpectAndNext(lexer.String, "http://test.com/image.jpg", "")
	h.ExpectAndNext(lexer.RParen, "", "")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_AtRule(t *testing.T) {
	h := NewHarness(t, lexer.NewLexer(`@import "test.css"`))

	h.ExpectAndNext(lexer.AtKeyword, "import", "")
	h.ExpectAndNext(lexer.String, "test.css", "")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_SimpleBlocks(t *testing.T) {
	h := NewHarness(t, lexer.NewLexer(`.class {
	width: 5px;
}

/** this is the root
	container id */
#id {
	margin: -2.75rem;
	content: "text";
}`))

	h.ExpectAndNext(lexer.Delim, ".", "")
	h.ExpectAndNext(lexer.Ident, "class", "")
	h.ExpectAndNext(lexer.LCurly, "", "")
	h.ExpectAndNext(lexer.Ident, "width", "")
	h.ExpectAndNext(lexer.Colon, "", "")
	h.ExpectAndNext(lexer.Dimension, "px", "5")
	h.ExpectAndNext(lexer.Semicolon, "", "")
	h.ExpectAndNext(lexer.RCurly, "", "")

	h.ExpectAndNext(lexer.Comment, "* this is the root\n\tcontainer id ", "")

	h.ExpectAndNext(lexer.Hash, "id", "")
	h.ExpectAndNext(lexer.LCurly, "", "")
	h.ExpectAndNext(lexer.Ident, "margin", "")
	h.ExpectAndNext(lexer.Colon, "", "")
	h.ExpectAndNext(lexer.Dimension, "rem", "-2.75")
	h.ExpectAndNext(lexer.Semicolon, "", "")
	h.ExpectAndNext(lexer.Ident, "content", "")
	h.ExpectAndNext(lexer.Colon, "", "")
	h.ExpectAndNext(lexer.String, "text", "")
	h.ExpectAndNext(lexer.Semicolon, "", "")
	h.ExpectAndNext(lexer.RCurly, "", "")
}
