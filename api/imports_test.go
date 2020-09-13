package api_test

import (
	"testing"

	"github.com/stephen/cssc/api"
	"github.com/stephen/cssc/api/transforms"
	"github.com/stretchr/testify/assert"
)

func TestImports(t *testing.T) {
	t.Run("inline", func(t *testing.T) {
		var errors TestReporter
		result := api.Compile(api.Options{
			Entry: []string{
				"testdata/imports/index.css",
			},
			Reporter: &errors,
			Transforms: transforms.Options{
				ImportRules: transforms.ImportRulesInline,
			},
		})

		assert.Len(t, result.Files, 1)
		assert.Len(t, errors, 0)
	})

	t.Run("passthrough", func(t *testing.T) {
		var errors TestReporter
		result := api.Compile(api.Options{
			Entry: []string{
				"testdata/imports/index.css",
			},
			Reporter: &errors,
			Transforms: transforms.Options{
				ImportRules: transforms.ImportRulesPassthrough,
			},
		})

		assert.Len(t, result.Files, 3)
		assert.Len(t, errors, 0)
	})
}
