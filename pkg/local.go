// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

type LocalPackage struct {
	Source *deps.Local
}

func NewLocalPackage(source *deps.Local) Interface {
	return &LocalPackage{
		Source: source,
	}
}

func (p *LocalPackage) Install(ctx context.Context, name, dir, version string) (lockVersion string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current working directory: %w")
	}

	fmt.Printf("name: %v\n", name)
	fmt.Printf("dir: %v\n", dir)
	fmt.Printf("version: %v\n", version)
	fmt.Printf("p.Source.Directory: %v\n", p.Source.Directory)

	oldname := filepath.Join(wd, p.Source.Directory)
	newname := filepath.Join(dir, name)
	linkname, err := filepath.Rel(dir, oldname)

	fmt.Printf("oldname: %v\n", oldname)
	fmt.Printf("newname: %v\n", newname)
	fmt.Printf("linkname: %v\n", linkname)

	localvendordir := filepath.Join(dir, "local", name)
	fmt.Printf("localvendordir: %v\n", localvendordir)

	if err != nil {
		linkname = oldname
	}

	err = os.RemoveAll(localvendordir)
	if err != nil {
		return "", errors.Wrap(err, "failed to clean previous destination path: %w")
	}

	err = CopyDir(p.Source.Directory, localvendordir)
	if err != nil {
		return "", errors.Wrap(err, "you fucked up: %w")
	}
	// copy in full directory (optional: only jsonnet and libsonnet files) oldname -> dir + 'local'
	// make symlink from newname -> dir + 'local'

	_, err = os.Stat(oldname)
	if os.IsNotExist(err) {
		return "", errors.Wrap(err, "symlink destination path does not exist: %w")
	}

	err = os.Symlink(localvendordir, newname)
	if err != nil {
		return "", errors.Wrap(err, "failed to create symlink for local dependency: %w")
	}

	color.Magenta("LOCAL %s -> %s", name, oldname)

	return "", nil
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
