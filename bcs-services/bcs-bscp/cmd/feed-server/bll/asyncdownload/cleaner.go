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

// Package asyncdownload NOTES
package asyncdownload

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// GB is the size of 1GB in bytes
const GB = 1024 * 1024 * 1024

// CacheCleaner scheduled task to clean source file cache dir
type CacheCleaner struct {
	// source file max cache size in GB
	cacheSizeGB int
	// source file cache retention rate, when cache size is greater than max size,
	// clean up oldest files but retain cache size * retention rate
	cacheRetentionRate float64
	ctx                context.Context
	cancel             context.CancelFunc
	metric             *metric
}

// NewCacheCleaner create a CacheCleaner
func NewCacheCleaner(cacaheSizeGB int, cacheRetentionRate float64, mc *metric) *CacheCleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &CacheCleaner{
		cacheSizeGB:        cacaheSizeGB,
		cacheRetentionRate: cacheRetentionRate,
		ctx:                ctx,
		cancel:             cancel,
		metric:             mc,
	}
}

// Run run a scheduled task
func (sfc *CacheCleaner) Run() {

	go func() {
		ticker := time.NewTicker(60 * time.Second)

		for {
			select {
			case <-ticker.C:
				sfc.do()
			case <-sfc.ctx.Done():
				logs.Infof("async download source files cache cleaner stopped")
				return
			}
		}
	}()
}

func (sfc *CacheCleaner) do() {
	sourceDir := cc.FeedServer().GSE.CacheDir
	if err := os.MkdirAll(sourceDir, os.ModePerm); err != nil {
		logs.Errorf("create source file cache dir %s failed, err %s", sourceDir, err.Error())
		return
	}
	size, count, err := sfc.getDirSize(sourceDir)
	if err != nil {
		logs.Errorf("get source file cache size failed, err %s", err.Error())
		return
	}
	sfc.metric.sourceFilesSizeBytes.Set(float64(size))
	sfc.metric.sourceFilesCounter.Set(float64(count))

	if size < int64(sfc.cacheSizeGB*GB) {
		return
	}
	logs.Infof("source file cache size %d MB bytes is greater than %d MB, start clean up oldest files",
		size/1024/1024, sfc.cacheSizeGB*1024)
	spaceToFree := size - int64(float64(sfc.cacheSizeGB*GB)*sfc.cacheRetentionRate)
	sfc.cleanupOldestFiles(sourceDir, spaceToFree)
	cleanedSize, cleanedCount, err := sfc.getDirSize(sourceDir)
	if err != nil {
		logs.Errorf("get source file cache size failed, err %s", err.Error())
		return
	}
	sfc.metric.sourceFilesSizeBytes.Set(float64(cleanedSize))
	sfc.metric.sourceFilesCounter.Set(float64(cleanedCount))
	logs.Infof("source file cache size after clean up %d MB", cleanedSize/1024/1024)
}

// Stop stop scheduled task
func (sfc *CacheCleaner) Stop() {
	sfc.cancel()
}

// getDirSize recursively traverses the directory to calculate its size, returning the size in bytes.
func (sfc *CacheCleaner) getDirSize(path string) (int64, int64, error) {
	var size int64
	var count int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return err
	})
	return size, count, err
}

func (sfc *CacheCleaner) cleanupOldestFiles(dir string, spaceToFree int64) {
	filePaths, filesMap := listFilesByModTime(dir)

	for _, filePath := range filePaths {
		if err := os.Remove(filePath); err != nil {
			logs.Errorf("remove source file %s failed, err: %s", filePath, err.Error())
		} else {
			logs.Infof("deleted source file %s", filePath)
			spaceToFree -= filesMap[filePath].Size()
		}

		if spaceToFree <= 0 {
			break
		}
	}
}

func listFilesByModTime(dir string) ([]string, map[string]os.FileInfo) {
	fileMap := make(map[string]os.FileInfo)
	filePaths := make([]string, 0)

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			filePaths = append(filePaths, path)
			fileMap[path] = info

		}
		return err
	}); err != nil {
		logs.Errorf("walk source files failed, err: %s", err.Error())
	}

	sort.Slice(filePaths, func(i, j int) bool {
		return fileMap[filePaths[i]].ModTime().Before(fileMap[filePaths[j]].ModTime())
	})

	return filePaths, fileMap
}
