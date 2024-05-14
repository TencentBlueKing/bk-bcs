/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ZipHandler zip xxx
type ZipHandler struct {
}

// NewZipHandler xxx
func NewZipHandler() *ZipHandler {
	return &ZipHandler{}
}

// UnZip unpacks a ZIP stream. When given a os.File reader it will get its size without
// reading the entire zip file in memory.
func (z *ZipHandler) UnZip(r io.Reader, destPath string) error {
	var (
		zr        *zip.Reader
		readerErr error
	)
	if f, ok := r.(*os.File); ok {
		fstat, err := f.Stat()
		if err != nil {
			return err
		}
		zr, readerErr = zip.NewReader(f, fstat.Size())
	} else {
		var fileBuffer bytes.Buffer
		_, err := io.Copy(&fileBuffer, r)
		if err != nil {
			return err
		}
		memReader := bytes.NewReader(fileBuffer.Bytes())
		zr, readerErr = zip.NewReader(memReader, memReader.Size())
	}

	if readerErr != nil {
		return readerErr
	}
	return z.unpackZip(zr, destPath)
}

func (z *ZipHandler) unpackZip(zr *zip.Reader, destPath string) error {
	for _, f := range zr.File {
		err := z.unzipFile(f, destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *ZipHandler) unzipFile(f *zip.File, destPath string) error {
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filepath.Join(destPath, f.Name), f.Mode().Perm()); err != nil {
			return err
		}
		return nil
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		// NOCC:gas/error(ignore)
		_ = rc.Close()
	}()

	filePath := sanitize(f.Name)
	destPath = filepath.Join(destPath, filePath)

	fileDir := filepath.Dir(destPath)
	_, err = os.Lstat(fileDir)
	if err != nil {
		if err = os.MkdirAll(fileDir, 0o700); err != nil {
			return err
		}
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	if err = file.Chmod(f.Mode()); err != nil {
		return fmt.Errorf("warn: failed setting file permissions for %q: %#v", file.Name(), err)
	}

	if err = os.Chtimes(file.Name(), time.Now(), f.Modified); err != nil {
		return fmt.Errorf("warn: failed setting file atime and mtime for %q: %#v", file.Name(), err)
	}

	if _, err = io.CopyN(file, rc, int64(f.UncompressedSize64)); err != nil {
		return err
	}

	return nil
}
