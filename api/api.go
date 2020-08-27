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
		astsByIndex:    make(map[int]*ast.Stylesheet),
		lockersByIndex: make(map[int]*sync.Mutex),
	}
}

type compilation struct {
	mu      sync.RWMutex
	sources map[string]int

	nextIndex int

	sourcesByIndex map[int]*lexer.Source
	astsByIndex    map[int]*ast.Stylesheet
	lockersByIndex map[int]*sync.Mutex
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

// Compile runs a compilation with the specified Options.
func Compile(opts Options) *Result {
	c := newCompilation()
	r := newResult()
	var wg errgroup.Group

	for _, e := range opts.Entry {
		entry := e
		wg.Go(func() error {
			idx, err := c.addSource(entry)
			if err != nil {
				r.Errors = append(r.Errors, err)
			}

			locker := c.lockersByIndex[idx]
			locker.Lock()
			ast := parser.Parse(c.sourcesByIndex[idx])
			c.astsByIndex[idx] = ast

			locker.Unlock()

			locker.Lock()
			r.Files[e] = printer.Print(c.astsByIndex[idx], printer.Options{
				OriginalSource: c.sourcesByIndex[idx],
			})
			locker.Unlock()
			return nil
		})
	}
	wg.Wait()

	return r
}
