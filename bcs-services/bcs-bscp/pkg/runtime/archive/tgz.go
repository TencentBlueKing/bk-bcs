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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TgzArchive 实现了 Archive 接口，用于处理 gzip 文件
type TgzArchive struct {
	destPath      string
	limitFileSize int64
}

// NewTgzArchive xxx
func NewTgzArchive(destPath string, limitFileSize int64) TgzArchive {
	return TgzArchive{
		destPath:      destPath,
		limitFileSize: limitFileSize,
	}
}

// UnTgzPack decompresses the gzip archive and returns an error
func (t TgzArchive) UnTgzPack(reader io.Reader) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer func() {
		// NOCC:gas/error(ignore)
		_ = gzr.Close()
	}()

	return t.UnTar(gzr)
}

// UnTar decompresses a TAR archive and returns an error
func (t TgzArchive) UnTar(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}

		if err != nil {
			return err
		}

		if hdr.Size > t.limitFileSize {
			return fmt.Errorf("file %s exceeds size", hdr.Name)
		}

		fp := filepath.Join(t.destPath, sanitize(hdr.Name))

		if hdr.FileInfo().IsDir() {
			if err = os.MkdirAll(fp, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
			continue
		}

		unTarErr := t.unTarFile(hdr, tr, fp)
		if unTarErr != nil {
			return unTarErr
		}
	}

	return nil
}

func (t TgzArchive) unTarFile(hdr *tar.Header, tr *tar.Reader, fp string) error {
	parentDir, _ := filepath.Split(fp)

	// NOCC:gas/permission(ignore)
	if err := os.MkdirAll(parentDir, 0o740); err != nil {
		return err
	}

	file, err := os.Create(fp)
	if err != nil {
		return err
	}

	defer func() {
		// NOCC:gas/error(ignore)
		_ = file.Close()
	}()

	if err = file.Chmod(os.FileMode(hdr.Mode)); err != nil {
		return fmt.Errorf("warn: failed setting file permissions for %q: %#v", file.Name(), err)
	}

	if err = os.Chtimes(file.Name(), time.Now(), hdr.ModTime); err != nil {
		return fmt.Errorf("warn: failed setting file atime and mtime for %q: %#v", file.Name(), err)
	}

	if _, err = io.Copy(file, tr); err != nil {
		return err
	}

	return nil
}
