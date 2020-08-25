package lexer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func NewHarness(t testing.TB, content string) *Harness {
	return &Harness{t, lexer.NewLexer(&lexer.Source{Path: "main.css", Content: content})}
}

// Harness is a simple test harness to expect on the lexer.
type Harness struct {
	testing.TB
	*lexer.Lexer
}

// ExpectAndNext asserts against the current token and drives the lexer forward.
// It resets CurrentString and CurrentNumeral before calling Next().
func (h *Harness) ExpectAndNext(token lexer.Token, stringLiteral, numericLiteral string) {
	h.TB.Helper()
	assert.Equal(h.TB, token, h.Current, "expected %s, but got %s", token.String(), h.Current.String())
	assert.Equal(h.TB, stringLiteral, h.CurrentString)
	assert.Equal(h.TB, numericLiteral, h.CurrentNumeral)

	h.CurrentNumeral, h.CurrentString = "", ""

	h.Next()
}

func (h *Harness) RunUntil(token lexer.Token) {
	for h.Current != token {
		h.Next()
	}
}
