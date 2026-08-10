// Package rt carries the source of the SafeGo intrinsic packages inside the safego binary.
//
// The intrinsics are part of the compiler, not a dependency it happens to be built against:
// what rt/volatile declares and what the emitter lowers are two halves of one contract, and
// they are versioned together (Spec 012 D6). A project imports pkg.safego.dev/rt through the
// module graph like any other dependency — that is what gives it a go.sum entry and a
// transparency-log record — but the compiler must also be able to answer "which intrinsic
// source is this binary's contract" from the binary alone, without a network or a module
// cache. That is what these bytes are for, and it is the same property runtime/embed.go
// holds for the C runtime.
//
// The embed directives live here because //go:embed cannot reach outside its own package
// directory, and this is that directory.
//
// This package has no other purpose. It declares no API, and nothing in the subset imports
// it: the intrinsics are the subdirectories, and a SafeGo program imports those.
package rt

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

// Source is the intrinsic tree: one directory per intrinsic package, plus the module file
// that names the tree's canonical import path.
//
//go:embed go.mod
//go:embed errors/*.go sync/*.go time/*.go volatile/*.go
var Source embed.FS

// File is one intrinsic source, named relative to the root of the rt module.
type File struct {
	Name    string
	Content []byte
}

// Collect returns every embedded source, in a deterministic order.
//
// Names keep their directory, so a written-out tree mirrors this module: a defect reported
// against an artifact is found in the source by the same path. Test files are embedded — the
// published module's tree hash covers them, and the two are meant to compare equal — but they
// are filtered out here, because they are not part of the contract the compiler lowers.
func Collect() ([]File, error) {
	var out []File

	err := fs.WalkDir(Source, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		content, readErr := Source.ReadFile(name)
		if readErr != nil {
			return readErr
		}

		out = append(out, File{Name: path.Clean(name), Content: content})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
