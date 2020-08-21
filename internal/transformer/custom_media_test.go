package transformer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomMedia(t *testing.T) {
	assert.Equal(t, "@media (max-width:30em){.a{color:green}}@media (max-width:30em) and (script){.c{color:red}}", Transform(`
	@custom-media --narrow-window (max-width: 30em);

	@media (--narrow-window) {
		.a { color: green; }
	}

	@media (--narrow-window) and (script) {
		.c { color: red; }
	}`))
}
