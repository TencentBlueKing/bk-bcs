/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pack

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Packer provides some compressed file actions
// It is not goroutine-safe, one Packer should only do the Pack once.
type Packer struct {
	tw *tar.Writer
}

// New return a new Packer instance
func New() *Packer {
	return &Packer{}
}

// Pack add all the files in baseDir into tgz file
func (p *Packer) Pack(baseDir string) ([]byte, error) {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	buf := bytes.NewBuffer(nil)
	if err := p.add(baseDir, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *Packer) add(currentDir string, buf io.Writer) error {
	gw := gzip.NewWriter(buf)
	defer func() {
		_ = gw.Close()
	}()
	p.tw = tar.NewWriter(gw)
	defer func() {
		_ = p.tw.Close()
	}()

	return filepath.WalkDir(currentDir, p.walkDirFunc)
}

func (p *Packer) walkDirFunc(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	// dir handler, go into the dir and do recursion
	if d.IsDir() {
		if isInBlackListOfDirs(path) {
			return fs.SkipDir
		}

		return nil
	}

	if !d.Type().IsRegular() {
		return fs.SkipDir
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}
	header.Name = path

	// Write file header to the tar archive
	err = p.tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(p.tw, f)
	if err != nil {
		return err
	}

	return nil
}

var (
	blackListOfDirs = map[string]bool{
		".git": true,
	}
)

func isInBlackListOfDirs(name string) bool {
	_, ok := blackListOfDirs[filepath.Base(name)]
	return ok
}
