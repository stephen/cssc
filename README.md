# cssc
A fast, friendly css compiler in go.

This repo is the start of a css compiler (parser, ast, and printer) and set of transforms to support new CSS syntax in current browsers. To
start with, it aims to be able to replace projects like [postcss-preset-env](https://github.com/csstools/postcss-preset-env) and [cssnext](https://github.com/MoOx/postcss-cssnext).

It's approach is inspired from experimenting with [esbuild](https://github.com/evanw/esbuild) (see [here](https://github.com/evanw/esbuild/issues/111#issuecomment-673115702)).

## Status
The package is currently **unusable**.

I have the start of a lexer, parser and ast, but they are all incomplete.

## Benchmarks
To keep track of performance, I've been benchmarking performance on (partially) [parsing bootstrap.css](https://github.com/postcss/benchmark).

```bash
$ go test -bench=. internal/parser/*.go
goos: darwin
goarch: amd64
BenchmarkParser-12    	     316	   3611143 ns/op	  826259 B/op	   23787 allocs/op
PASS
```

For the ~180Kb bootstrap.css file, the parser allocates ~800Kb and takes ~3.5ms.

I expect this to be a moving target as I complete the parser implementation.
