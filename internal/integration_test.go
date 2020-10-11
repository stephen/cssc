package cssc

import (
	"io/ioutil"
	"testing"

	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	for _, c := range []string{
		"testdata/bootstrap.css",
		"testdata/comments.css",
	} {
		t.Run(c, func(t *testing.T) {

			by, err := ioutil.ReadFile(c)
			require.NoError(t, err)
			source := &sources.Source{
				Path:    c,
				Content: string(by),
			}

			ast, err := parser.Parse(source)
			require.NoError(t, err)

			printer.Print(ast, printer.Options{})
		})
	}
}
