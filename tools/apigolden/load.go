// The loader is stdlib-only, adapted from gogpu's tools/apigolden load.go
// (gogpu PR #575, commit a0d90ca). See the load doc comment for the reason.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// listedPackage is the subset of `go list -json` this tool reads.
type listedPackage struct {
	ImportPath string
	Dir        string
	Name       string
	GoFiles    []string
	CgoFiles   []string
	Export     string
	Error      *struct{ Err string }
}

// load type-checks the module's public packages for one build context.
//
// The loader deliberately uses only the standard library: `go list -export`
// produces export data with the running toolchain, and go/importer decodes it
// with that same toolchain, so the tool cannot drift out of step with the Go
// version that invokes it (golang.org/x/tools bundles its own export-data
// reader, which lags a new Go release — on a cold build cache it fails outright
// exactly when the gate has no other proof). The package under analysis is
// parsed from source and type-checked here, so parameter names survive into the
// golden; dependencies come from export data.
func load(root string, c config) ([]*types.Package, error) {
	list, err := goList(root, c)
	if err != nil {
		return nil, err
	}

	exports := make(map[string]string, len(list))
	for _, p := range list {
		if p.Export != "" {
			exports[p.ImportPath] = p.Export
		}
	}

	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		file, ok := exports[path]
		if !ok {
			return nil, fmt.Errorf("no export data for %q", path)
		}
		return os.Open(file)
	})

	var keep []*types.Package
	for _, p := range list {
		if !include(p.ImportPath) {
			continue
		}
		if len(p.GoFiles) == 0 && len(p.CgoFiles) == 0 {
			// The package has no files for this context (an OS-specific file
			// set). It is absent from this context's API, not broken.
			if p.Error != nil && strings.Contains(p.Error.Err, "build constraints exclude all Go files") {
				continue
			}
			return nil, fmt.Errorf("package %s has no files for %s", p.ImportPath, c.describe())
		}
		if p.Error != nil {
			return nil, fmt.Errorf("package %s did not list: %s", p.ImportPath, p.Error.Err)
		}

		files := make([]*ast.File, 0, len(p.GoFiles))
		for _, name := range p.GoFiles {
			file, err := parser.ParseFile(fset, filepath.Join(p.Dir, name), nil, parser.SkipObjectResolution)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", filepath.Join(p.Dir, name), err)
			}
			files = append(files, file)
		}

		conf := types.Config{
			Importer:         imp,
			Sizes:            types.SizesFor("gc", c.goarch),
			IgnoreFuncBodies: true,
		}
		pkg, err := conf.Check(p.ImportPath, fset, files, nil)
		if err != nil {
			return nil, fmt.Errorf("type-check %s: %w", p.ImportPath, err)
		}
		keep = append(keep, pkg)
	}
	sort.Slice(keep, func(i, j int) bool { return keep[i].Path() < keep[j].Path() })
	return keep, nil
}

// goList runs `go list -e -json -deps -export ./...` in the module root with
// the context forced, and decodes the stream of package objects.
func goList(root string, c config) ([]listedPackage, error) {
	cmd := exec.Command("go", "list", "-e", "-json", "-deps", "-export", "./...")
	cmd.Dir = root
	cmd.Env = contextEnv(c)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	dec := json.NewDecoder(bytes.NewReader(out))
	var list []listedPackage
	for {
		var p listedPackage
		err := dec.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode go list output: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}
