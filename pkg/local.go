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
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/otiai10/copy"

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

func (p *LocalPackage) Install(ctx context.Context, dependencyName, dependencyDir, version string) (lockVersion string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current working directory: %w")
	}

	// fmt.Printf("name: %v\n", p.Source)
	// fmt.Printf("name: %v\n", dependencyName)
	// fmt.Printf("dir: %v\n", dependencyDir)
	// fmt.Printf("version: %v\n", version)
	// fmt.Printf("p.Source.Directory: %v\n", p.Source.Directory)
	// fmt.Printf("p.Source.TargetPath: %v\n", p.Source.TargetPath)
	// fmt.Printf("p.Source.HardCopy: %v\n", p.Source.HardCopy)

	pathToLocalSource := filepath.Join(wd, p.Source.Directory)
	desiredLocationOfImport := filepath.Join(dependencyDir, p.Source.TargetPath, dependencyName)

	// fmt.Printf("pathToLocalSource: %v\n", pathToLocalSource)
	// fmt.Printf("desiredLocationOfImport: %v\n", desiredLocationOfImport)

	_, err = os.Stat(pathToLocalSource)
	if os.IsNotExist(err) {
		return "", errors.Wrap(err, "symlink destination path does not exist: %w")
	}

	// remove all existing items in desired import location
	err = os.RemoveAll(desiredLocationOfImport)
	if err != nil {
		return "", errors.Wrap(err, "failed to clean previous destination path: %w")
	}

	if p.Source.HardCopy == true {
		err := copy.Copy(p.Source.Directory, desiredLocationOfImport)
		if err != nil {
			return "", errors.Wrap(err, "failed to copy in local dependency: %w")
		}
		color.Magenta("LOCAL COPIED %s -> %s", pathToLocalSource, desiredLocationOfImport)
	}
	if p.Source.HardCopy == false {
		dirOfSymlink, _ := filepath.Split(desiredLocationOfImport)
		symlink, err := filepath.Rel(dirOfSymlink, pathToLocalSource)

		dir, _ := filepath.Split(desiredLocationOfImport)
		os.MkdirAll(dir, os.ModePerm)

		err = os.Symlink(symlink, desiredLocationOfImport)
		if err != nil {
			return "", errors.Wrap(err, "failed to create symlink for local dependency: %w")
		}
		color.Magenta("LOCAL SYMLINK'D %s -> %s", desiredLocationOfImport, symlink)
	}

	return "", nil
}
