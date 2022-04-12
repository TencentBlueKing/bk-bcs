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
)

// BufferedFile stores the file in memory
type BufferedFile struct {
	Name    string `json:"name" yaml:"name"`
	Content []byte `json:"content" yaml:"content"`
}

// UnpackFromTgz receive a tgz file then uncompressed and return all the files in it.
func UnpackFromTgz(data []byte) ([]*BufferedFile, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	files := make([]*BufferedFile, 0)
	tr := tar.NewReader(gz)
	for {
		b := bytes.NewBuffer(nil)
		hd, err := tr.Next()
		if err == io.EOF {
			break
		}

		if hd.FileInfo().IsDir() {
			// Use this instead of hd.Typeflag because we don't have to do any
			// inference chasing.
			continue
		}

		switch hd.Typeflag {
		// We don't want to process these extension header files.
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			continue
		}

		if _, err := io.Copy(b, tr); err != nil {
			return nil, err
		}

		files = append(files, &BufferedFile{
			Name:    hd.Name,
			Content: b.Bytes(),
		})

		b.Reset()
	}

	return files, nil
}
