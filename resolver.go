package cssc

// Resolver implements a method of resolving an import spec (e.g. @import "test.css")
// into a path on the filesystem.
type Resolver interface {
	// Resolve spec relative to from.
	Resolve(spec, fromDir string) (path string, err error)
}
