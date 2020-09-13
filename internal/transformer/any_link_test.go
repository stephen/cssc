package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/transformer"
	"github.com/stephen/cssc/transforms"
	"github.com/stretchr/testify/assert"
)

func compileAnyLink(o *transformer.Options) {
	o.AnyLink = transforms.AnyLinkTransform
}

func TestAnyLink(t *testing.T) {
	assert.Equal(t, ".test:visited,.test:link{color:red}", Transform(t, compileAnyLink, `
.test:any-link {
	color: red;
}`))

	assert.Equal(t, "complex .test:visited:not(.thing),complex .test:link:not(.thing){color:red}", Transform(t, compileAnyLink, `
complex .test:any-link:not(.thing) {
	color: red;
}`))

	assert.Equal(t, "a:visited,a:link,section,.Something{color:red}", Transform(t, compileAnyLink, `
a:any-link, section, .Something {
	color: red;
}`))

	assert.Equal(t, ".test:any-link{color:red}", Transform(t, nil, `
.test:any-link {
	color: red;
}`))
}
