package transformer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomProperties(t *testing.T) {
	assert.Equal(t, ".class{margin:0rem 1rem 3rem 5rem}", Transform(t, nil, `:root {
	--var-width: 1rem 3rem 5rem;
}

.class {
	margin: 0rem var(--var-width);
}`))

	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif,other}`, Transform(t, nil, `:root {
		--font: "Helvetica", sans-serif, other;
}

.class {
	font-family: var(--font);
}`))

}

func TestCustomProperties_Fallback(t *testing.T) {
	assert.Equal(t, ".class{margin:0rem 2rem}", Transform(t, nil, `.class {
	margin: 0rem var(--var-width, 2rem);
}`))

	assert.Equal(t, ".class{margin:0rem 2rem 1rem 3rem}", Transform(t, nil, `.class {
	margin: 0rem var(--var-width, 2rem 1rem 3rem);
}`))

	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif}`, Transform(t, nil, `.class {
		font-family: var(--font, "Helvetica", sans-serif);
	}`))
}
