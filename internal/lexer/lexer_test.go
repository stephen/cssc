package lexer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer_URL(t *testing.T) {
	h := NewHarness(t, "url(http://test.com/image.jpg)")

	h.ExpectAndNext(lexer.URL, "http://test.com/image.jpg", "")
	h.ExpectAndNext(lexer.EOF, "", "")

	assert.Panics(t, func() {
		NewHarness(t, "url(http://test.com/image.jpg").Next()
	})

	assert.Panics(t, func() {
		NewHarness(t, "url(())").Next()
	})
}

func TestLexer_Function(t *testing.T) {
	h := NewHarness(t, `url("http://test.com/image.jpg")`)

	h.ExpectAndNext(lexer.FunctionStart, "url", "")
	h.ExpectAndNext(lexer.String, "http://test.com/image.jpg", "")
	h.ExpectAndNext(lexer.RParen, "", "")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_Function_RGB(t *testing.T) {
	h := NewHarness(t, `rgb(255, 254, 253)`)

	h.ExpectAndNext(lexer.FunctionStart, "rgb", "")
	h.ExpectAndNext(lexer.Number, "", "255")
	h.ExpectAndNext(lexer.Comma, "", "")
	h.ExpectAndNext(lexer.Number, "", "254")
	h.ExpectAndNext(lexer.Comma, "", "")
	h.ExpectAndNext(lexer.Number, "", "253")
	h.ExpectAndNext(lexer.RParen, "", "")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_Number(t *testing.T) {
	h := NewHarness(t, `20%`)

	h.ExpectAndNext(lexer.Percentage, "", "20")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_Dimension(t *testing.T) {
	h := NewHarness(t, `0.15s`)

	h.ExpectAndNext(lexer.Dimension, "s", "0.15")
	h.ExpectAndNext(lexer.EOF, "", "")

	h = NewHarness(t, `2rem`)

	h.ExpectAndNext(lexer.Dimension, "rem", "2")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_AtRule(t *testing.T) {
	h := NewHarness(t, `@import "test.css"`)

	h.ExpectAndNext(lexer.At, "import", "")
	h.ExpectAndNext(lexer.String, "test.css", "")
	h.ExpectAndNext(lexer.EOF, "", "")
}

func TestLexer_SimpleBlocks(t *testing.T) {
	h := NewHarness(t, `.class {
	width: 5px;
}

/** this is the root
	container id */
#id {
	margin: -2.75rem;
	content: "text";
}`)

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

func TestLexer_BrowserPrefix(t *testing.T) {
	h := NewHarness(t, `[list]::-webkit-calendar-picker-indicator`)

	h.ExpectAndNext(lexer.LBracket, "", "")
	h.ExpectAndNext(lexer.Ident, "list", "")
	h.ExpectAndNext(lexer.RBracket, "", "")
	h.ExpectAndNext(lexer.Colon, "", "")
	h.ExpectAndNext(lexer.Colon, "", "")
	h.ExpectAndNext(lexer.Ident, "-webkit-calendar-picker-indicator", "")
}

func TestLexer_Errorf(t *testing.T) {
	assert.PanicsWithValue(t, "main.css:3:6\nunclosed string: unexpected newline:\n\t  bad: \"no good;\n\t       ~~~~~~~~~", func() {
		NewHarness(t, `.class {
	something: "ok";
	bad: "no good;
}`).RunUntil(lexer.EOF)
	})
}
