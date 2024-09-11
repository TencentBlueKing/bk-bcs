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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
)

// ZipArchive 实现了 Archive 接口，用于处理 zip 文件
type ZipArchive struct {
	destPath      string
	limitFileSize int64
}

// NewZipArchive xxx
func NewZipArchive(destPath string, limitFileSize int64) ZipArchive {
	return ZipArchive{
		destPath:      destPath,
		limitFileSize: limitFileSize,
	}
}

// 检测源字符集 转换成utf-8
func checkCharacterSets(src string) (string, error) {

	// 创建字符集检测器
	detector := chardet.NewTextDetector()

	// 检测字符集
	result, err := detector.DetectBest([]byte(src))
	if err != nil {
		return "", err
	}

	decoder := mahonia.NewDecoder(result.Charset)

	return decoder.ConvertString(src), nil
}

// UnZipPack decompresses the zip package and receives the parameter io.Reader
func (z ZipArchive) UnZipPack(reader io.Reader) error {
	// 将请求体内容写入临时文件
	tempZipFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return err
	}
	// 函数结束后删除临时文件
	defer func() {
		_ = os.Remove(tempZipFile.Name())
	}()

	if _, err = io.Copy(tempZipFile, reader); err != nil {
		return err
	}

	if err = tempZipFile.Close(); err != nil {
		return err
	}

	// 打开临时文件以进行解压缩
	zipFile, err := os.Open(tempZipFile.Name())
	if err != nil {
		return err
	}
	defer zipFile.Close()

	return z.Unzip(zipFile)
}

// Unzip decompresses the zip package and receives the parameter os.File
func (z ZipArchive) Unzip(zipFile *os.File) error {
	// 获取临时文件的信息
	fileInfo, err := zipFile.Stat()
	if err != nil {
		return err
	}
	// 创建zip读取器
	zr, err := zip.NewReader(zipFile, fileInfo.Size())
	if err != nil {
		return err
	}

	return z.unpackZip(zr)
}

func (z ZipArchive) unpackZip(zr *zip.Reader) error {
	for _, f := range zr.File {
		if f.UncompressedSize64 > uint64(z.limitFileSize) {
			return errf.New(int32(FileTooLarge), f.Name)
		}
		err := z.unzipFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (z ZipArchive) unzipFile(f *zip.File) error {
	fileName, err := checkCharacterSets(f.Name)
	if err != nil {
		return err
	}
	if f.FileInfo().IsDir() {
		if err = os.MkdirAll(filepath.Join(z.destPath, fileName), f.Mode().Perm()); err != nil {
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
	filePath := sanitize(fileName)
	z.destPath = filepath.Join(z.destPath, filePath)

	fileDir := filepath.Dir(z.destPath)
	_, err = os.Lstat(fileDir)
	if err != nil {
		if err = os.MkdirAll(fileDir, 0o700); err != nil {
			return err
		}
	}

	file, err := os.Create(z.destPath)
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
