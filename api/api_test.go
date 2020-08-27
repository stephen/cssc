package api_test

import (
	"log"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/api"
	"github.com/stretchr/testify/assert"
)

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

	log.Println(spew.Config.Sdump(result.Files))
	assert.Len(t, result.Files, 1)
	assert.Len(t, result.Errors, 0)
}
