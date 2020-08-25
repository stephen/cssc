package cssc

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stretchr/testify/require"
)

func TestBootstrap(t *testing.T) {
	by, err := ioutil.ReadFile("testdata/bootstrap.css")
	require.NoError(t, err)
	source := &lexer.Source{
		Path:    "bootstrap.css",
		Content: string(by),
	}

	printer.Print(parser.Parse(source), printer.Options{})
}
