package lexer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer_URLToken(t *testing.T) {
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
