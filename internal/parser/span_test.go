package parser_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/logging"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Parse(t testing.TB, source *sources.Source) *ast.Stylesheet {
	ss, err := parser.Parse(source)
	require.NoError(t, err)
	return ss
}

func TestSpans(t *testing.T) {
	source := &sources.Source{
		Path: "main.css",
		Content: `
.simple { margin: 1px 2px; }
.more, .complex { margin: 1px 2px; }

.multiline {
	width: 50%;
}

@media (screen) or (print) {
	#id {
		line-height: 2rem;
	}
}

/* what is this even targetting?? */
.big#THINGS + div, span::after, a:href, a:visited, not.a:problem + a[link="thing"] {
	color: purple;
	background-color: rgba(calc(0 + 1), 2, 3);
}

@custom-media test (800px < width < 1000px) or (print);

@keyframes custom {
	0% { opacity: 0%; }
	100% { opacity: 100%; }
}

`,
	}

	var b bytes.Buffer
	ast.Walk(Parse(t, source), func(n ast.Node) {
		line, col := source.LineAndCol(*n.Location())
		fmt.Fprintln(&b, fmt.Sprintf("%s:%d:%d", reflect.TypeOf(n).String(), line, col))
		// XXX: does Location need to be a pointer still?
		fmt.Fprintln(&b, logging.AnnotateSourceSpan(source, *n.Location()))
		fmt.Fprintln(&b)
	})

	if os.Getenv("WRITE_SNAPSHOTS") != "" {
		require.NoError(t, os.MkdirAll("testdata/", 0644))
		require.NoError(t, ioutil.WriteFile("testdata/spans.txt", b.Bytes(), 0644))
		return
	}

	expected, err := ioutil.ReadFile("testdata/spans.txt")
	require.NoError(t, err, "if you are trying to generate the snapshot, use WRITE_SNAPSHOTS=1")

	assert.Equal(t, string(expected), b.String())
}
