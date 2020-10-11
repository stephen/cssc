package cssc_test

import (
	"testing"

	"github.com/stephen/cssc"
	"github.com/stephen/cssc/transforms"
	"github.com/stretchr/testify/assert"
)

func TestImports(t *testing.T) {
	t.Run("inline", func(t *testing.T) {
		var errors TestReporter
		result := cssc.Compile(cssc.Options{
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
		result := cssc.Compile(cssc.Options{
			Entry: []string{
				"testdata/imports/index.css",
			},
			Reporter: &errors,
			Transforms: transforms.Options{
				ImportRules: transforms.ImportRulesPassthrough,
			},
		})

		assert.Len(t, result.Files, 1)
		assert.Len(t, errors, 0)
	})

	t.Run("follow", func(t *testing.T) {
		var errors TestReporter
		result := cssc.Compile(cssc.Options{
			Entry: []string{
				"testdata/imports/index.css",
			},
			Reporter: &errors,
			Transforms: transforms.Options{
				ImportRules: transforms.ImportRulesFollow,
			},
		})

		assert.Len(t, result.Files, 3)
		assert.Len(t, errors, 0)
	})
}
