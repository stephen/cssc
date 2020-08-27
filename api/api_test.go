package api_test

import (
	"testing"

	"github.com/stephen/cssc/api"
	"github.com/stretchr/testify/assert"
)

func TestApi_Error(t *testing.T) {
	result := api.Compile(api.Options{
		Entry: []string{
			"testdata/nonexistent/index.css",
		},
	})

	assert.Len(t, result.Files, 0)
	assert.Len(t, result.Errors, 1)
}

func TestApi_Simple(t *testing.T) {
	result := api.Compile(api.Options{
		Entry: []string{
			"testdata/simple/index.css",
		},
	})

	assert.Len(t, result.Files, 1)
	assert.Len(t, result.Errors, 0)
}

func TestApi_Imports(t *testing.T) {
	result := api.Compile(api.Options{
		Entry: []string{
			"testdata/imports/index.css",
		},
	})

	assert.Len(t, result.Files, 1)
	assert.Len(t, result.Errors, 0)
}
