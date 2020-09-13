# cssc
[![PkgGoDev](https://pkg.go.dev/badge/github.com/stephen/cssc?tab=doc)](https://pkg.go.dev/github.com/stephen/cssc?tab=doc)

A fast, friendly css compiler in go.

This repo is the start of a css compiler (parser, ast, and printer) and set of transforms to support new CSS syntax in current browsers. To
start with, it aims to be able to replace projects like [postcss-preset-env](https://github.com/csstools/postcss-preset-env) and [cssnext](https://github.com/MoOx/postcss-cssnext).

It's approach is inspired from experimenting with [esbuild](https://github.com/evanw/esbuild) (see [here](https://github.com/evanw/esbuild/issues/111#issuecomment-673115702)).

## Status
The package can currently parse and print most standard CSS. There are likely bugs in both.

Some transforms are supported:

| Transform  | Support | Notes |
| ------------- | ------------- | ------------- |
| [`@import` rules](https://www.w3.org/TR/css-cascade-4) | Partial | Only non-conditional imports can be inlined. Import conditions will be ignored. |
| [Custom Properties](https://www.w3.org/TR/css-variables-1/) | Partial | Only variables defined on `:root` will be substituted. The compiler will ignore any non-`:root` variables. [See #3](https://github.com/stephen/cssc/issues/3). |
| [Custom Media Queries](https://www.w3.org/TR/mediaqueries-5/#custom-mq) | Complete | |
| [Media Feature Ranges](https://www.w3.org/TR/mediaqueries-4/#mq-min-max) | Complete | |
| [`:any-link`](https://www.w3.org/TR/selectors-4/#the-any-link-pseudo) | Complete | |

## API
For now, there is only a go API.

```golang
package main

import (
  "log"

  "github.com/stephen/cssc"
)

func main() {
  result := cssc.Compile(cssc.Options{
    Entry: []string{"css/index.css"},
  })

  // result.Files is a map of all output files.
  for path, content := range result.Files {
    log.Println(path, content)
  }
}
```

### Transforms
Transforms can be specified via options:
```golang
package main

import (
  "github.com/stephen/cssc"
  "github.com/stephen/cssc/transforms"
)

func main() {
  result := cssc.Compile(cssc.Options{
    Entry: []string{"css/index.css"},
    Transforms: transforms.Options{
      // Transform :any-link into :link and :visited equivalents.
      AnyLink: transforms.AnyLinkTransform,
      // Keep @import rules without transforming them or inlining their content.
      ImportRules: transforms.ImportRulesPassthrough,
    },
  })

  // result.Files...
}
```

By default, all features are in passthrough mode and will not get transformed.

### Error reporting
By default, errors and warnings are printed to stderr. You can control this behavior by providing a [Reporter](https://pkg.go.dev/github.com/stephen/cssc?tab=doc#Reporter):
```golang
package main

import (
  "log"

  "github.com/stephen/cssc"
)

type TestReporter []error

func (r *TestReporter) AddError(err error) {
  *r = append(*r, err)
}

func main() {
  var errors TestReporter
  result := cssc.Compile(cssc.Options{
    Entry:    []string{"css/index.css"},
    Reporter: &errors,
  })

  for _, err := range errors {
    log.Println(err)
  }
}
```


## Benchmarks
To keep track of performance, I've been benchmarking performance on (partially) [parsing bootstrap.css](https://github.com/postcss/benchmark).

```bash
$ go test -bench=. internal/parser/*.go
goos: darwin
goarch: amd64
BenchmarkParser-12    	     296	   3934884 ns/op	 1548281 B/op	   45916 allocs/op
PASS
```

I expect this to be a moving target as I complete the parser implementation.
