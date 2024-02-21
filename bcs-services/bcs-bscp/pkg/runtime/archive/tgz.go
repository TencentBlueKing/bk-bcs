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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TgzHandler tgz xxx
type TgzHandler struct {
}

// NewTgzHandler xxx
func NewTgzHandler() *TgzHandler {
	return &TgzHandler{}
}

// UnTar unarchives a TAR archive and returns the final destination path or an error
func (t *TgzHandler) UnTar(r io.Reader, destPath string) error {
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

		fp := filepath.Join(destPath, sanitize(hdr.Name))

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

func (t *TgzHandler) unTarFile(hdr *tar.Header, tr *tar.Reader, fp string) error {
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
