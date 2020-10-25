package cssc

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/samsarahq/go/oops"
)

// ErrNotFound is returned when the resolver cannot resolve a path.
var ErrNotFound = errors.New("could not resolve css module")

// Resolver implements a method of resolving an import spec (e.g. @import "test.css")
// into a
type Resolver interface {
	// Resolve spec relative to from.
	Resolve(spec, fromDir string) (path string, err error)
}

// NodeResolver implements the default node import resolution strategy. See
// https://www.typescriptlang.org/docs/handbook/module-resolution.html.
//
// When resolving node_modules, the resolver will use the style attribute in
// package.json for resolution.
type NodeResolver struct {
	// BaseURL is the root directory of the project. It serves
	// the same purpose as baseUrl in tsconfig.json. If the value is relative,
	// it will be resolved against the current working directory.
	BaseURL string
}

// Resolve implements Resolver.
func (r *NodeResolver) Resolve(spec, fromDir string) (string, error) {
	if isRelative := strings.HasPrefix(spec, "../") || strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "/"); isRelative {
		path := filepath.Join(fromDir, spec)
		if res, err := r.resolve(path); err != nil {
			return "", oops.Wrapf(err, "could not resolve %s relative to %s", spec, fromDir)
		} else {
			return res, nil
		}
	}

	// For non-relative imports, first try resolving against baseUrl.
	if r.BaseURL != "" {
		if res, err := r.resolve(filepath.Join(r.BaseURL, spec)); err == nil {
			return res, nil
		}
	}

	// Lastly, try looking through node_modules.
	res, err := r.resolveFromNodeModules(spec, fromDir)
	if err != nil {
		return "", oops.Wrapf(ErrNotFound, "could not resolve absolute path %s from %s", spec, fromDir)
	}

	return res, nil
}

type packageJSON struct {
	Style string `json:"style"`
}

// resolve attempts to resolve given absolute path as a file, then
// as a package folder, then as a folder with an index.
func (r *NodeResolver) resolve(absPath string) (string, error) {
	// Attempt to resolve first as a file.
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If it doesn't exist, try to resolve with an extension.
			withExtension := absPath + ".css"
			info, err := os.Stat(withExtension)
			if err != nil {
				if os.IsNotExist(err) {
					return "", oops.Wrapf(ErrNotFound, "could not resolve as file: %s or %s", absPath, withExtension)
				}
				return "", oops.Wrapf(err, "failure during resolution")
			}

			if info.IsDir() {
				return "", oops.Wrapf(ErrNotFound, "%s exists, but is directory", withExtension)
			}

			return withExtension, nil
		}

		return "", oops.Wrapf(err, "failure during resolution")
	}

	if !info.IsDir() {
		return absPath, nil
	}

	// Otherwise, try to resolve as a directory.
	path, err := r.resolveAsDir(absPath)
	if err == nil {
		return path, nil
	}

	return "", oops.Wrapf(err, "could not resolve path: %s", absPath)
}

// resolveAsDir takes a directory path and resolves its css entry point.
func (r *NodeResolver) resolveAsDir(path string) (string, error) {
	pkgPath := filepath.Join(path, "package.json")
	pkg, err := ioutil.ReadFile(pkgPath)
	if err != nil {
		// If there is no package.json, then try resolving an index.css file.
		indexPath := filepath.Join(path, "index.css")
		if info, err := os.Stat(indexPath); err != nil || info.IsDir() {
			if os.IsNotExist(err) {
				return "", oops.Wrapf(ErrNotFound, "could not resolve as directory: %s", path)
			}
			return "", oops.Wrapf(err, "failure during resolution")
		}

		return indexPath, nil
	}

	// Look for the style attribute in the package.json.
	var pkgContent packageJSON
	if err := json.Unmarshal(pkg, &pkgContent); err != nil {
		return "", oops.Wrapf(err, "failed to read package.json: %s", pkgPath)
	}

	if pkgContent.Style == "" {
		return "", oops.Wrapf(ErrNotFound, "package.json exists, but has no style attribute: %s", pkgPath)
	}

	stylePath := filepath.Join(path, pkgContent.Style)
	if info, err := os.Stat(stylePath); err != nil || info.IsDir() {
		if os.IsNotExist(err) {
			return "", oops.Wrapf(ErrNotFound, "package.json has style attribute, but it cannot be resolved: %s (to %s)", pkgContent.Style, stylePath)
		}
		return "", oops.Wrapf(err, "failure during resolution")
	}

	return stylePath, nil
}

// resolveAsNodeModule walks directories from fromDir to find node_modules paths.
func (r *NodeResolver) resolveFromNodeModules(module, fromDir string) (string, error) {
	currentDir := fromDir
	for currentDir != "/" {
		modulePkgPath := filepath.Join(currentDir, "node_modules", module)

		if res, err := r.resolve(modulePkgPath); err == nil {
			return res, nil
		}

		currentDir = filepath.Dir(currentDir)
	}

	return "", oops.Wrapf(ErrNotFound, "could not find absolute path in node_modules")
}
