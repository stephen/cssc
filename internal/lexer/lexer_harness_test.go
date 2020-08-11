package lexer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func NewHarness(t testing.TB, l *lexer.Lexer) *Harness {
	return &Harness{t, l}
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
