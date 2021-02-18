package cssc_test

import (
	"testing"

	"github.com/stephen/cssc"
	"github.com/stephen/cssc/transforms"
	"github.com/stretchr/testify/assert"
)

type TestReporter []error

func (r *TestReporter) AddError(err error) {
	*r = append(*r, err)
}

func TestApi_Error(t *testing.T) {
	var errors TestReporter
	result := cssc.Compile(cssc.Options{
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
	result := cssc.Compile(cssc.Options{
		Entry: []string{
			"testdata/simple/index.css",
		},
		Reporter: &errors,
	})

	assert.Len(t, result.Files, 1)
	assert.Len(t, errors, 0)
}

func TestApi_BrokenImport(t *testing.T) {
	var errors TestReporter
	result := cssc.Compile(cssc.Options{
		Entry: []string{
			"testdata/brokenimports/index.css",
		},
		Transforms: transforms.Options{
			ImportRules: transforms.ImportRulesInline,
		},
		Reporter: &errors,
	})

	assert.Len(t, result.Files, 1)
	assert.Len(t, errors, 1)
}

func TestApi_Crlf(t *testing.T) {
	var errors TestReporter
	result := cssc.Compile(cssc.Options{
		Entry: []string{
			"testdata/crlf/monaco.css",
		},
		Transforms: transforms.Options{
			ImportRules: transforms.ImportRulesInline,
		},
		Reporter: &errors,
	})

	assert.Len(t, result.Files, 1)
	assert.Len(t, errors, 0)
}
