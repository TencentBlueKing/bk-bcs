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
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// GZIP files start with 1F 8B
	gzipMagic = []byte{0x1F, 0x8B}
	// ZIP files start with 50 4B 03 04 or 50 4B 05 06 or 50 4B 07 08
	zipMagic = []byte{0x50, 0x4B}
	// Check if the buffer contains a valid tar header
	ustarMagic  = []byte("ustar\x00")
	gnuTarMagic = []byte("ustar  \x00")
)

// ArchiveType 文件类型
type ArchiveType string // nolint

const (
	// GZIP gzip
	GZIP ArchiveType = "GZIP"
	// ZIP zip
	ZIP ArchiveType = "ZIP"
	// TAR tar
	TAR ArchiveType = "TAR"
	// Unknown unknown
	Unknown ArchiveType = "Unknown"
)

// Unpack 解压zip、gzip、tar
func Unpack(reader io.Reader, archiveType ArchiveType) (string, error) {

	tempDir, err := os.MkdirTemp("", "configItem-")
	if err != nil {
		return "", err
	}
	switch archiveType {
	case ZIP:
		if err := NewZipArchive(tempDir).UnZipPack(reader); err != nil {
			return "", err
		}
	case GZIP:
		if err := NewTgzArchive(tempDir).UnTgzPack(reader); err != nil {
			return "", err
		}
	case TAR:
		if err := NewTgzArchive(tempDir).UnTar(reader); err != nil {
			return "", err
		}
	default:
		return "", errors.New("file type detection failed")
	}

	return tempDir, nil
}

// IdentifyFileType 检测文件类型：zip、zip、tar
func IdentifyFileType(buf []byte) ArchiveType {
	if isGzip(buf) {
		return GZIP
	} else if isZip(buf) {
		return ZIP
	} else if isTar(buf) {
		return TAR
	}
	return Unknown
}

func isGzip(buf []byte) bool {
	return len(buf) >= len(gzipMagic) && bytes.HasPrefix(buf, gzipMagic)
}

func isZip(buf []byte) bool {
	return len(buf) >= len(zipMagic) && bytes.HasPrefix(buf, zipMagic)
}

func isTar(buf []byte) bool {
	return bytes.Equal(buf[257:263], ustarMagic) || bytes.Equal(buf[257:265], gnuTarMagic)
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
