package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/internal/transformer"
	"github.com/stephen/cssc/transforms"
	"github.com/stretchr/testify/assert"
)

func compileMediaQueryRanges(o *transformer.Options) {
	o.MediaFeatureRanges = transforms.MediaFeatureRangesTransform
}

func TestMediaQueryRanges(t *testing.T) {

	assert.Equal(t, `@media (min-width:200px) and (max-width:600px),(min-width:200px),(max-width:600px){}`,
		Transform(t, compileMediaQueryRanges, `@media (200px <= width <= 600px), (200px <= width), (width <= 600px) {}`))

	assert.Equal(t, `@media (max-width:200px) and (min-width:600px),(max-width:200px),(min-width:600px){}`,
		Transform(t, compileMediaQueryRanges, `@media (200px >= width >= 600px), (200px >= width), (width >= 600px) {}`))

	assert.Equal(t, `@media (min-width:25.001%) and (max-width:74.999%){}`, Transform(t, compileMediaQueryRanges, `@media (25% < width < 75%) {}`))
	assert.Equal(t, `@media (min-width:200.001px) and (max-width:599.999px){}`, Transform(t, compileMediaQueryRanges, `@media (200px < width < 600px) {}`))
	assert.Equal(t, `@media (max-width:599.999px){}`, Transform(t, compileMediaQueryRanges, `@media (width < 600px) {}`))
	assert.Equal(t, `@media (min-width:200.001px){}`, Transform(t, compileMediaQueryRanges, `@media (200px < width) {}`))
	assert.Equal(t, `@media (max-width:199.999px) and (min-width:600.001px){}`, Transform(t, compileMediaQueryRanges, `@media (200px > width > 600px) {}`))
	assert.Equal(t, `@media (min-width:600.001px){}`, Transform(t, compileMediaQueryRanges, `@media (width > 600px) {}`))
	assert.Equal(t, `@media (max-width:199.999px){}`, Transform(t, compileMediaQueryRanges, `@media (200px > width) {}`))
}

func TestMediaQueryRanges_Passthrough(t *testing.T) {
	assert.Equal(t, `@media (200px>=width>=600px),(200px>=width),(width>=600px){}`,
		Transform(t, nil, `@media (200px >= width >= 600px), (200px >= width), (width >= 600px) {}`))
}
