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

// Package crontab NOTES
package crontab

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// MaxSourceFileCacheSizeMB source file dir total max size
var MaxSourceFileCacheSizeMB int64 = 1024 * 10

// SourceFileCacheCleaner scheduled task to clean source file cache dir
type SourceFileCacheCleaner struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSourceFileCacheCleaner create a SourceFileCacheCleaner
func NewSourceFileCacheCleaner() *SourceFileCacheCleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &SourceFileCacheCleaner{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run run a scheduled task
func (sfc *SourceFileCacheCleaner) Run() {

	go func() {
		ticker := time.NewTicker(5 * time.Minute)

		for {
			select {
			case <-ticker.C:
				// 1. check if source file cache is expired
				source := cc.FeedServer().GSE.SourceDir
				size, err := getDirSize(source)
				if err != nil {
					logs.Errorf("get source file cache size failed, err %s", err.Error())
					continue
				}
				// 2. if expired, delete source file cache
				sizeMB := size / 1024 / 1024
				if sizeMB > MaxSourceFileCacheSizeMB {
					logs.Infof("source file cache size %f MB is greater than %d MB, start to clean",
						sizeMB, MaxSourceFileCacheSizeMB)
					if err := os.RemoveAll(source); err != nil {
						logs.Errorf("clean source file cache failed, err %s", err.Error())
						continue
					}
					logs.Infof("clean source file cache success")
				}
			case <-sfc.ctx.Done():
				logs.Infof("source file cache cleaner stopped")
				return
			}
		}
	}()
}

// Stop stop scheduled task
func (sfc *SourceFileCacheCleaner) Stop() {
	sfc.cancel()
}

// getDirSize recursively traverses the directory to calculate its size, returning the size in bytes.
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
