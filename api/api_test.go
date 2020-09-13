package api_test

import (
	"testing"

	"github.com/stephen/cssc/api"
	"github.com/stretchr/testify/assert"
)

type TestReporter []error

func (r *TestReporter) AddError(err error) {
	*r = append(*r, err)
}

func TestApi_Error(t *testing.T) {
	var errors TestReporter
	result := api.Compile(api.Options{
		Entry: []string{
			"testdata/nonexistent/index.css",
		},
		Reporter: &errors,
	})

	assert.Len(t, result.Files, 0)
	assert.Len(t, errors, 1)
}

func TestApi_Simple(t *testing.T) {
	var errors TestReporter
	result := api.Compile(api.Options{
		Entry: []string{
			"testdata/simple/index.css",
		},
		Reporter: &errors,
	})

	assert.Len(t, result.Files, 1)
	assert.Len(t, errors, 0)
}
