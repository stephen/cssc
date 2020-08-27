package api_test

import (
	"log"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/api"
	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	result := api.Compile(api.Options{
		Entry: []string{
			"testdata/index.css",
		},
	})

	log.Println(spew.Config.Sdump(result.Files))
	assert.Len(t, result.Files, 0)
	assert.Len(t, result.Errors, 0)
}
