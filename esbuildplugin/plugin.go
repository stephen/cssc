package esbuildplugin

import (
	"context"
	"runtime/pprof"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/samsarahq/go/oops"
	"github.com/stephen/cssc"
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/transforms"
)

type csscReporter []error

func (r *csscReporter) AddError(err error) {
	*r = append(*r, err)
}

type locationError interface {
	Location() (*sources.Source, ast.Span)
}

func (r *csscReporter) toEsbuild() []api.Message {
	var errs []api.Message
	for _, err := range *r {

		if lErr, ok := err.(locationError); ok {
			source, span := lErr.Location()
			_, colNumber := source.LineAndCol(span)
			lineSpan := source.FullLine(span)

			errs = append(errs, api.Message{
				Text: err.Error(),
				Location: &api.Location{
					File:     source.Path,
					Column:   int(colNumber),
					Line:     lineSpan.Start,
					Length:   lineSpan.End - lineSpan.Start,
					LineText: source.Content[lineSpan.Start:lineSpan.End],
				},
			})
		}

		errs = append(errs, api.Message{
			Text: err.Error(),
		})
	}

	return errs
}

// Option is an optional argument to the plugin.
type Option func(cssc.Options) cssc.Options

// WithTransforms sets the transform options for the plugin.
func WithTransforms(transforms transforms.Options) Option {
	return func(opts cssc.Options) cssc.Options {
		opts.Transforms = transforms
		return opts
	}
}

// WithResolver sets the import resolver for the plugin.
func WithResolver(resolver cssc.Resolver) Option {
	return func(opts cssc.Options) cssc.Options {
		opts.Resolver = resolver
		return opts
	}
}

// Plugin is an esbuild plugin for importing .css files.
func Plugin(opts ...Option) api.Plugin {
	return api.Plugin{
		Name: "cssc",
		Setup: func(build api.PluginBuild) {
			build.OnLoad(
				api.OnLoadOptions{Filter: `\.css$`},
				func(args api.OnLoadArgs) (res api.OnLoadResult, err error) {
					res.Loader = api.LoaderCSS

					var errors csscReporter
					options := cssc.Options{
						Entry:    []string{args.Path},
						Reporter: &errors,
					}
					for _, opt := range opts {
						options = opt(options)
					}

					pprof.SetGoroutineLabels(pprof.WithLabels(context.TODO(), pprof.Labels("cssc-path", args.Path)))
					result := cssc.Compile(options)

					if len(errors) > 0 {
						res.Errors = errors.toEsbuild()
						return
					}

					f, ok := result.Files[args.Path]
					if !ok {
						err = oops.Errorf("cssc output did not contain %s", args.Path)
						return
					}

					res.Contents = &f
					return res, nil
				},
			)
		},
	}
}
