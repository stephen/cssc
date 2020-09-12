package cssc

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/require"
)

func TestBootstrap(t *testing.T) {
	by, err := ioutil.ReadFile("testdata/bootstrap.css")
	require.NoError(t, err)
	source := &sources.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}

	ast, err := parser.Parse(source)
	require.NoError(t, err)

	printer.Print(ast, printer.Options{})
}
