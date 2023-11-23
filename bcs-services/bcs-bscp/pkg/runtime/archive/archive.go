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

// Package archive xxx
package archive

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	magicZIP = []byte{0x50, 0x4b, 0x03, 0x04} // ZIP文件的魔法数字
	magicGZ  = []byte{0x1f, 0x8b}             // GZIP文件的魔法数字
	magicTAR = []byte{0x75, 0x73, 0x74, 0x61} // TAR文件的魔法数字
)

func magicNumber(reader *bufio.Reader, offset int) (string, error) {
	headerBytes, err := reader.Peek(offset + 6)
	if err != nil {
		return "", err
	}

	magic := headerBytes[offset : offset+6]

	if bytes.Equal(magicTAR, magic[0:5]) {
		return "tar", nil
	}

	if bytes.Equal(magicZIP, magic[0:4]) {
		return "zip", nil
	}

	if bytes.Equal(magicGZ, magic[0:2]) {
		return "gzip", nil
	}

	return "", nil
}

// Unpack unpacks a compressed stream. Magic numbers are used to determine what
// decompressor and/or unarchiver to use.
func Unpack(reader io.Reader) (string, error) {
	r := bufio.NewReader(reader)

	var (
		gzr *gzip.Reader
		err error
	)
	// Reads magic number from the stream so we can better determine how to proceed
	fType, err := magicNumber(r, 0)
	if err != nil {
		return "", err
	}

	// Create a temporary folder
	tmpDir, err := generateTempDir()
	if err != nil {
		return "", err
	}
	switch fType {
	case "zip":
		return tmpDir, NewZipHandler().UnZip(r, tmpDir)
	case "gzip":
		gzr, err = gzip.NewReader(r)
		if err != nil {
			return "", err
		}
		defer func() {
			_ = gzr.Close()
		}()
		return tmpDir, NewTgzHandler().UnTar(gzr, tmpDir)
	case "tar":
		return tmpDir, NewTgzHandler().UnTar(gzr, tmpDir)
	}

	return "", fmt.Errorf("this package is not supported")
}

// Generate temporary directory
func generateTempDir() (string, error) {
	return os.MkdirTemp("", "template-config-*")
}

// Sanitizes name to avoid overwriting sensitive system files when unarchiving
func sanitize(name string) string {
	// Gets rid of volume drive label in Windows
	if len(name) > 1 && name[1] == ':' && runtime.GOOS == "windows" {
		name = name[2:]
	}

	name = filepath.Clean(name)
	name = filepath.ToSlash(name)
	for strings.HasPrefix(name, "../") {
		name = name[3:]
	}

	return name
}

// gbkToUTF8 将 GBK 编码转为 UTF-8 编码
func gbkToUTF8(s string) string {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	result, _ := io.ReadAll(reader)
	return string(result)
}
