package cssc

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/samsarahq/go/oops"
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/logging"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/internal/transformer"
	"github.com/stephen/cssc/resolver"
	"github.com/stephen/cssc/transforms"
	"golang.org/x/sync/errgroup"
)

// Options is the set of options to pass to Compile.
type Options struct {
	// Entry is the set of files to start parsing.
	Entry []string

	// Reporter is the error and warning reporter. If not specified, the default
	// reporter prints to stderr.
	Reporter Reporter

	Transforms transforms.Options

	// Resolver is a path resolver. If not specified, the default node-style
	// resolver will be used.
	Resolver Resolver
}

func newCompilation(opts Options) *compilation {
	c := &compilation{
		sources:        make(map[string]int),
		sourcesByIndex: make(map[int]*sources.Source),
		outputsByIndex: make(map[int]struct{}),
		astsByIndex:    make(map[int]*ast.Stylesheet),
		result:         newResult(),
		reporter:       logging.DefaultReporter,
		transforms:     opts.Transforms,
		resolver:       &resolver.NodeResolver{},
	}

	if opts.Reporter != nil {
		c.reporter = opts.Reporter
	}

	if opts.Resolver != nil {
		c.resolver = opts.Resolver
	}

	return c
}

type compilation struct {
	// mu synchronizes other globals like the error reporter.
	mu sync.Mutex

	// sourcesMu synchronizes assignment for new source indices.
	sourcesMu sync.RWMutex
	sources   map[string]int
	nextIndex int

	sourcesByIndexMu sync.RWMutex
	sourcesByIndex   map[int]*sources.Source

	astsByIndexMu sync.RWMutex
	astsByIndex   map[int]*ast.Stylesheet

	// outputsByIndex is the set of sources to write outputs for.
	outputsByIndex map[int]struct{}

	result *Result

	reporter Reporter

	transforms transforms.Options

	resolver Resolver
}

// addSource will read in a path and assign it a source index. If
// it's already been loaded, the cached source is returned.
func (c *compilation) addSource(path string) (int, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, oops.Wrapf(err, "failed to make path absolute: %s", path)
	}

	c.sourcesMu.RLock()
	if _, ok := c.sources[abs]; ok {
		defer c.sourcesMu.RUnlock()
		return c.sources[abs], nil
	}
	c.sourcesMu.RUnlock()

	in, err := ioutil.ReadFile(abs)
	if err != nil {
		return 0, oops.Wrapf(err, "failed to read file: %s", path)
	}

	source := &sources.Source{
		Content: string(in),
		Path:    abs,
	}

	c.sourcesMu.Lock()
	i := c.nextIndex
	c.sources[abs] = i
	c.sourcesMu.Unlock()

	c.sourcesByIndexMu.Lock()
	c.sourcesByIndex[i] = source
	c.sourcesByIndexMu.Unlock()

	c.nextIndex++
	return i, nil
}

func newResult() *Result {
	return &Result{
		Files: make(map[string]string),
	}
}

// Result is the results of a compilation.
type Result struct {
	mu    sync.Mutex
	Files map[string]string
}

func (c *compilation) addError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reporter.AddError(err)
}

// parseFile assigns the file a source index and parses the source. It also
// looks at imported files and adds them to the compilation. hasOutput should
// be called if the file should be included in compilation output.
//
// parseFile also runs the last transformation pass on the output. Note that we
// don't make this function print the output as well so that we can make the current
// file available to any callers as a dependency.
func (c *compilation) parseFile(file string, hasOutput bool) *ast.Stylesheet {
	// Assign the file a source index.
	idx, err := c.addSource(file)
	if err != nil {
		c.addError(err)
		return nil
	}

	if hasOutput {
		c.outputsByIndex[idx] = struct{}{}
	}

	c.astsByIndexMu.RLock()
	if ss, ok := c.astsByIndex[idx]; ok {
		c.astsByIndexMu.RUnlock()
		return ss
	}
	c.astsByIndexMu.RUnlock()

	c.sourcesByIndexMu.RLock()
	source := c.sourcesByIndex[idx]
	c.sourcesByIndexMu.RUnlock()
	ss, err := parser.Parse(source)
	if err != nil {
		c.addError(err)
		return nil
	}

	// Immediately look at the imports from the file and feed those dependencies
	// into parseFile as well. If we're set to inline imports, then we'll use
	// collect those dependency ASTs to let the transformer replace them.
	var mu sync.Mutex
	replacements := make(map[*ast.AtRule]*ast.Stylesheet)
	var wg errgroup.Group
	for _, imp := range ss.Imports {
		wg.Go(func() error {
			rel, err := c.resolver.Resolve(imp.Value, filepath.Dir(source.Path))
			if err != nil {
				c.addError(err)
				return nil
			}

			// If import follow is on, then every referenced file makes it to the output.
			imported := c.parseFile(rel, c.transforms.ImportRules == transforms.ImportRulesFollow)

			mu.Lock()
			defer mu.Unlock()
			if imported != nil {
				replacements[imp.AtRule] = imported
			}
			return nil
		})
	}
	wg.Wait()

	opts := transformer.Options{
		Options:        c.transforms,
		OriginalSource: source,
		Reporter:       c.reporter,
	}

	if c.transforms.ImportRules == transforms.ImportRulesInline {
		opts.ImportReplacements = replacements
	}

	ss = transformer.Transform(ss, opts)
	c.astsByIndexMu.Lock()
	c.astsByIndex[idx] = ss
	c.astsByIndexMu.Unlock()
	return ss
}

// Compile runs a compilation with the specified Options.
func Compile(opts Options) *Result {
	c := newCompilation(opts)

	var wg errgroup.Group

	for _, e := range opts.Entry {
		wg.Go(func() error {
			c.parseFile(e, true)
			return nil
		})
	}
	wg.Wait()

	wg = errgroup.Group{}
	for i := range c.outputsByIndex {
		idx := i
		wg.Go(func() error {
			// XXX: this is the wrong file name
			source := c.sourcesByIndex[idx]
			ast := c.astsByIndex[idx]
			if source == nil || ast == nil {
				// Skip attempting to print if there was a problem reading/parsing this file
				// in the first place.
				return nil
			}

			out, err := printer.Print(ast, printer.Options{
				OriginalSource: source,
			})
			if err != nil {
				c.addError(err)
				return nil
			}

			c.result.mu.Lock()
			defer c.result.mu.Unlock()
			c.result.Files[source.Path] = out
			return nil
		})
	}
	wg.Wait()

	return c.result
}

// Reporter is an error and warning reporter.
//
// Note that it is the same type as logging.Reporter, which is
// an internal-only interface.
type Reporter interface {
	AddError(err error)
}
