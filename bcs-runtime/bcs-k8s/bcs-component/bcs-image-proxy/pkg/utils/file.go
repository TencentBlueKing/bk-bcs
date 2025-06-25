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

// Package utils xxx
package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

// IsSparseFile check linux file is sparse file
func IsSparseFile(filePath string) (int64, int64, bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, false, errors.Wrap(err, "os.stat file '%s' failed")
	}

	// 获取系统底层文件信息（仅支持类Unix系统）
	sysStat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false, fmt.Errorf("unsupported file system")
	}

	// 计算实际磁盘占用（字节）
	physicalSize := sysStat.Blocks * 512
	logicalSize := fileInfo.Size()

	// 判断是否为稀疏文件（逻辑大小 >> 物理占用）
	threshold := int64(10) // 阈值可调整
	return logicalSize, physicalSize, logicalSize > physicalSize*threshold, nil
}

// CopyFile copy source file to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CreateTarGz create tar.gz file
func CreateTarGz(srcDir, dstFile string) error {
	_ = os.RemoveAll(dstFile)
	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()
	gw := gzip.NewWriter(dst)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk through the source directory
	err = filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// NOTE: ignore tmp dir
		if fi.IsDir() && fi.Name() == "tmp" {
			return nil
		}

		var link string
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(file); err != nil {
				return err
			}
		}
		// Get the header for the current file
		header, err := tar.FileInfoHeader(fi, link)
		if err != nil {
			return err
		}
		// Set the correct path in the header
		relFilePath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		header.Name = relFilePath

		// If it's not a directory, set the header size
		if !fi.Mode().IsDir() {
			header.Size = fi.Size()
		}
		// Write the header to the tar writer
		if err = tw.WriteHeader(header); err != nil {
			return err
		}
		// nothing more to do for non-regular
		if !fi.Mode().IsRegular() {
			return nil
		}
		// If it's a directory, we don't need to write its content
		if fi.Mode().IsDir() {
			return nil
		}

		// Open and copy the file's content
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})

	return err
}

// ConcurrentDirSize concurrent calculate dir size
func ConcurrentDirSize(root string) (int64, error) {
	var total int64
	sizes := make(chan int64)
	var wg sync.WaitGroup

	wg.Add(1)
	go walk(root, &wg, sizes)

	go func() {
		wg.Wait()
		close(sizes)
	}()

	for s := range sizes {
		total += s
	}
	return total, nil
}

func walk(dir string, wg *sync.WaitGroup, sizes chan<- int64) {
	defer wg.Done()
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go walk(subdir, wg, sizes)
		} else {
			info, _ := entry.Info()
			sizes <- info.Size()
		}
	}
}
