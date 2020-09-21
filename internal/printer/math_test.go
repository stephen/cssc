package printer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMath(t *testing.T) {
	assert.Equal(t, `.class{width:calc(1px+2px)}`, Print(t, `.class { width: calc(1px + 2px) }`))
	assert.Equal(t, `.class{width:calc(1px+2px/2)}`, Print(t, `.class { width: calc(1px + 2px / 2) }`))
	assert.Equal(t, `.class{width:calc(22%+1rem)}`, Print(t, `.class { width: calc(22% + 1rem) }`))
	assert.Equal(t, `.class{width:calc(22%-5%)}`, Print(t, `.class { width: calc(22% - 5%) }`))
}
