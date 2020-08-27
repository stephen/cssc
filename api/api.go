package api

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/samsarahq/go/oops"
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/parser"
	"github.com/stephen/cssc/internal/printer"
	"golang.org/x/sync/errgroup"
)

// Options is the set of options to pass to Compile.
type Options struct {
	// Entry is the set of files to start parsing.
	Entry []string
}

func newCompilation() *compilation {
	return &compilation{
		sources:        make(map[string]int),
		sourcesByIndex: make(map[int]*lexer.Source),
		outputsByIndex: make(map[int]struct{}),
		astsByIndex:    make(map[int]*ast.Stylesheet),
		lockersByIndex: make(map[int]*sync.Mutex),
		result:         newResult(),
	}
}

type compilation struct {
	mu      sync.RWMutex
	sources map[string]int

	nextIndex int

	sourcesByIndex map[int]*lexer.Source
	astsByIndex    map[int]*ast.Stylesheet
	outputsByIndex map[int]struct{}
	lockersByIndex map[int]*sync.Mutex

	result *Result
}

// addSource will read in a path and assign it a source index. If
// it's already been loaded, the cached source is returned.
func (c *compilation) addSource(path string) (int, error) {
	c.mu.RLock()
	if _, ok := c.sources[path]; ok {
		defer c.mu.RUnlock()
		return c.sources[path], nil
	}
	c.mu.RUnlock()

	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, oops.Wrapf(err, "failed to make path absolute: %s", path)
	}

	in, err := ioutil.ReadFile(abs)
	if err != nil {
		return 0, oops.Wrapf(err, "failed to read file: %s", path)
	}

	source := &lexer.Source{
		Content: string(in),
		Path:    abs,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	i := c.nextIndex
	c.sources[abs] = i
	c.sourcesByIndex[i] = source
	c.lockersByIndex[i] = &sync.Mutex{}

	c.nextIndex++
	return i, nil
}

// parse -> imports -> transform -> print

func newResult() *Result {
	return &Result{
		Files: make(map[string]string),
	}
}

// Result is the results of a compilation.
type Result struct {
	Files map[string]string

	Errors []error
}

func (c *compilation) parseFile(file string, hasOutput bool) {
	idx, err := c.addSource(file)
	if err != nil {
		c.result.Errors = append(c.result.Errors, err)
	}

	locker := c.lockersByIndex[idx]
	locker.Lock()
	if _, ok := c.astsByIndex[idx]; ok {
		locker.Unlock()
		return
	}

	ast := parser.Parse(c.sourcesByIndex[idx])
	c.astsByIndex[idx] = ast
	c.outputsByIndex[idx] = struct{}{}
	locker.Unlock()

	wg := errgroup.Group{}
	for _, imp := range ast.Imports {
		wg.Go(func() error {
			// XXX: if @import passthrough is on, then this is true.
			c.parseFile(imp, false)
			return nil
		})
	}
	wg.Wait()

}

// Compile runs a compilation with the specified Options.
func Compile(opts Options) *Result {
	c := newCompilation()
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
			c.result.Files[source.Path] = printer.Print(c.astsByIndex[idx], printer.Options{
				OriginalSource: source,
			})
			return nil
		})
	}
	wg.Wait()

	return c.result
}
