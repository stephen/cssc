package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/transformer"
	"github.com/stephen/cssc/transforms"
	"github.com/stretchr/testify/assert"
)

func compileMath(o *transformer.Options) {
	o.CalcReduction = transforms.CalcReductionReduce
}

func TestMath(t *testing.T) {
	assert.Equal(t, `.class{width:3px}`, Transform(t, compileMath, `.class { width: calc(1px + 2px) }`))
	assert.Equal(t, `.class{width:-1px}`, Transform(t, compileMath, `.class { width: calc(1px - 2px) }`))
	assert.Equal(t, `.class{width:calc(1px+2rem)}`, Transform(t, compileMath, `.class { width: calc(1px + 2rem) }`))
	assert.Equal(t, `.class{width:17%}`, Transform(t, compileMath, `.class { width: calc(22% - 5%) }`))
	assert.Panics(t, func() { Transform(t, compileMath, `.class { width: calc(2 + 25%) }`) })

	assert.Equal(t, `.class{width:35%}`, Transform(t, compileMath, `.class { width: calc(10% + 25%) }`))
	assert.Equal(t, `.class{width:50%}`, Transform(t, compileMath, `.class { width: calc(2 * 25%) }`))
	assert.Panics(t, func() { Transform(t, compileMath, `.class { width: calc(2% * 25%) }`) })

	assert.Equal(t, `.class{width:5%}`, Transform(t, compileMath, `.class { width: calc(10% / 2) }`))

	// XXX: fix precision in below
	assert.Equal(t, `.class{width:3.3333333333333335%}`, Transform(t, compileMath, `.class { width: calc(10% / 3) }`))
	assert.Equal(t, `.class{width:4px}`, Transform(t, compileMath, `.class { width: calc(20px / 5) }`))
	assert.Equal(t, `.class{width:20}`, Transform(t, compileMath, `.class { width: calc(20 / 1) }`))
	assert.Panics(t, func() { Transform(t, compileMath, `.class { width: calc(2% / 25%) }`) })
	assert.Panics(t, func() { Transform(t, compileMath, `.class { width: calc(2% / 0) }`) })

	assert.Equal(t, `.class{width:3px}`, Transform(t, compileMath, `.class { width: calc(1px + 4px / 2) }`))

	assert.Equal(t, `.class{width:calc(1px+2px)}`, Transform(t, nil, `.class { width: calc(1px + 2px) }`))

	// XXX: this should work but doesn't. because we treat math as left-associative and binary but
	// 22% - 1rem cannot be reduced into something that can be added to 8rem.
	// assert.Equal(t, `.class{width:calc(22%+7rem)}`, Transform(t, compileMath, `.class { width: calc(22% - 1rem + 8rem) }`))
}
