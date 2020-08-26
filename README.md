# cssc
A fast, friendly css compiler in go.

This repo is the start of a css compiler (parser, ast, and printer) and set of transforms to support new CSS syntax in current browsers. To
start with, it aims to be able to replace projects like [postcss-preset-env](https://github.com/csstools/postcss-preset-env) and [cssnext](https://github.com/MoOx/postcss-cssnext).

It's approach is inspired from experimenting with [esbuild](https://github.com/evanw/esbuild) (see [here](https://github.com/evanw/esbuild/issues/111#issuecomment-673115702)).

## Status
The package is currently **unusable**.

I have the start of a lexer, parser, printer, transformer and ast, but they are all incomplete.


| Transform  | Support | Notes |
| ------------- | ------------- | ------------- |
| [Custom Properties](https://www.w3.org/TR/css-variables-1/) | Complete | |
| [Custom Media Queries](https://www.w3.org/TR/mediaqueries-5/#custom-mq) | Complete | |
| [Media Feature Ranges](https://www.w3.org/TR/mediaqueries-4/#mq-min-max) | Partial | `<=` and `>=` are supported, but `<` and `>` are not yet. |

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
