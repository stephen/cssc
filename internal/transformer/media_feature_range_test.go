package transformer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaQueryRanges(t *testing.T) {
	assert.Equal(t, `@media (min-width:200px) (max-width:600px),(min-width:200px),(max-width:600px){}`,
		Transform(t, `@media (200px <= width <= 600px), (200px <= width), (width <= 600px) {}`))

	assert.Equal(t, `@media (max-width:200px) (min-width:600px),(max-width:200px),(min-width:600px){}`,
		Transform(t, `@media (200px >= width >= 600px), (200px >= width), (width >= 600px) {}`))
}

func TestMediaQueryRanges_Unsupported(t *testing.T) {
	assert.Panics(t, func() { Transform(t, `@media (200px < width < 600px) {}`) })
	assert.Panics(t, func() { Transform(t, `@media (width < 600px) {}`) })
	assert.Panics(t, func() { Transform(t, `@media (200px < width) {}`) })
	assert.Panics(t, func() { Transform(t, `@media (200px > width > 600px) {}`) })
	assert.Panics(t, func() { Transform(t, `@media (width > 600px) {}`) })
	assert.Panics(t, func() { Transform(t, `@media (200px > width) {}`) })
}
