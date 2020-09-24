package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpanMutability(t *testing.T) {
	n := &Comma{}
	n.Location().Position = 10
	assert.Equal(t, 10, n.Span.Position)
}
