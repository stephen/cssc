package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/transformer"
	"github.com/stephen/cssc/transforms"
	"github.com/stretchr/testify/assert"
)

func compileCustomProperties(o *transformer.Options) {
	o.CustomProperties = transforms.CustomPropertiesTransformRoot
}

func TestCustomProperties(t *testing.T) {
	assert.Equal(t, ".class{margin:0rem 1rem 3rem 5rem}", Transform(t, compileCustomProperties, `:root {
	--var-width: 1rem 3rem 5rem;
}

.class {
	margin: 0rem var(--var-width);
}`))

	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif,other}`, Transform(t, compileCustomProperties, `:root {
		--font: "Helvetica", sans-serif, other;
}

.class {
	font-family: var(--font);
}`))

}

func TestCustomProperties_Fallback(t *testing.T) {
	assert.Equal(t, ".class{margin:0rem 2rem}", Transform(t, compileCustomProperties, `.class {
	margin: 0rem var(--var-width, 2rem);
}`))

	assert.Equal(t, ".class{margin:0rem 2rem 1rem 3rem}", Transform(t, compileCustomProperties, `.class {
	margin: 0rem var(--var-width, 2rem 1rem 3rem);
}`))

	assert.Equal(t, `.class{font-family:"Helvetica",sans-serif}`, Transform(t, compileCustomProperties, `.class {
		font-family: var(--font, "Helvetica", sans-serif);
	}`))
}
