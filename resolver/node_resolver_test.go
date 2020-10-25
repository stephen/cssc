package resolver_test

import (
	"path/filepath"
	"testing"

	"github.com/stephen/cssc/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_Relative(t *testing.T) {
	testdata, err := filepath.Abs("testdata/")
	require.NoError(t, err)

	r := resolver.NodeResolver{}

	result, err := r.Resolve("./case-1.css", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-1.css"), result)

	result, err = r.Resolve("./case-1", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-1.css"), result)

	result, err = r.Resolve("./case-2", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-2/index.css"), result)

	result, err = r.Resolve("./case-3", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-3/whatever.css"), result)

	result, err = r.Resolve("./case-2/index.css", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-2/index.css"), result)

	result, err = r.Resolve("./case-3/whatever.css", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-3/whatever.css"), result)

	result, err = r.Resolve("./case-0", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	result, err = r.Resolve("./case-0.css", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	result, err = r.Resolve("./case-7", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	result, err = r.Resolve("./case-7.css", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	result, err = r.Resolve("./case-8.css", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-8.css/index.css"), result)
}

func TestResolver_Absolute_WithBaseURL(t *testing.T) {
	testdata, err := filepath.Abs("testdata/")
	require.NoError(t, err)

	r := resolver.NodeResolver{BaseURL: testdata}

	result, err := r.Resolve("case-1.css", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-1.css"), result)

	result, err = r.Resolve("case-2", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-2/index.css"), result)

	result, err = r.Resolve("case-3", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "case-3/whatever.css"), result)

	result, err = r.Resolve("case-0", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	result, err = r.Resolve("case-7", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	result, err = r.Resolve("case-7.css", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)
}

func TestResolver_Absolute(t *testing.T) {
	testdata, err := filepath.Abs("testdata/nested/1/2/")
	require.NoError(t, err)

	r := resolver.NodeResolver{}

	result, err := r.Resolve("case-4", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "./node_modules/case-4.css"), result)

	result, err = r.Resolve("case-5", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "../node_modules/case-5/index.css"), result)

	result, err = r.Resolve("case-6", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "../../node_modules/case-6/dist/whatever.css"), result)

	result, err = r.Resolve("case-6/dist/unreferenced.css", testdata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testdata, "../../node_modules/case-6/dist/unreferenced.css"), result)

	result, err = r.Resolve("case-0", testdata)
	assert.Error(t, err)
	assert.Equal(t, "", result)
}
