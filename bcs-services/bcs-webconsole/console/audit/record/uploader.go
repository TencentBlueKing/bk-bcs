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

// Package record xxx
package record

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/repository"
)

var (
	// singletonUploader 单例 Uploader
	singletonUploader *Uploader
	once              sync.Once
	castPattern       = regexp.MustCompile(`/(\d{4}-\d{2}-\d{2})/.*.cast`)
)

// GetGlobalUploader : get global Uploader
func GetGlobalUploader() *Uploader {
	if singletonUploader == nil {
		once.Do(func() {
			var err error
			singletonUploader, err = newUploader()

			if err != nil {
				panic(err)
			}
		})
	}
	return singletonUploader
}

// Uploader 文件上传
type Uploader struct {
	lock          sync.Mutex
	castFileState map[string]state
	storage       repository.Provider
	dataDir       string
}

func newUploader() (*Uploader, error) {
	u := &Uploader{
		dataDir:       config.G.Audit.DataDir,
		castFileState: map[string]state{},
	}

	// StorageType 不设置表示不上传, 只使用本地模式, 文件需要自己清理
	if config.G.Repository.StorageType != "" {
		storage, err := repository.NewProvider(config.G.Repository.StorageType)
		if err != nil {
			return nil, err
		}
		u.storage = storage
	}

	if err := ensureDirExist(config.G.Audit.DataDir); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *Uploader) setState(filePath string, s state) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.castFileState[filePath] = s
}

// IntervalUpload 定时上传
func (u *Uploader) IntervalUpload(ctx context.Context) error {
	if !config.G.Audit.Enabled || u.storage == nil {
		blog.Info("audit not enabled or storage type not set, uploader just ignore")
		return nil
	}

	timer := time.NewTicker(time.Second * 10)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			// 获取退出信号, 全部上传
			result, err := u.batchUpload(true)
			if err != nil {
				blog.Errorf("interval remain upload failed, err: %s", err)
				return err
			}

			blog.Infof("interval remain upload done, result: %s", result)
			return nil

		case <-timer.C:
			result, err := u.batchUpload(false)
			if err != nil {
				blog.Errorf("interval upload failed, err: %s", err)
				continue
			}

			blog.Infof("interval upload done, result: %s", result)
		}
	}
}

// IntervalDelete 定时删除
func (u *Uploader) IntervalDelete(ctx context.Context) error {
	// RetentionDays 为0的情况下为永久的，不删除
	if u.storage == nil || config.G.Audit.RetentionDays == 0 {
		blog.Info("storage type not set or retention days is 0, delete just ignore")
		return nil
	}

	timer := time.NewTicker(time.Hour)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			blog.Infof("interval delete folders done by ctx")
			return nil

		case <-timer.C:
			err := u.deleteFolders()
			if err != nil {
				blog.Errorf("interval delete folders failed, err: %s", err)
				continue
			}

			blog.Infof("interval delete folders done")
		}
	}
}

func (u *Uploader) checkFileCanUpload(path string) bool {
	u.lock.Lock()
	defer u.lock.Unlock()

	s, ok := u.castFileState[path]
	if !ok {
		return true
	}

	if s == terminationState {
		return true
	}

	return false
}

type batchUploadResult struct {
	success []string
	failed  []string
	ignore  []string
}

func (r *batchUploadResult) String() string {
	return fmt.Sprintf("success=%s, failed=%s, ignore=%s", r.success, r.failed, r.ignore)
}

// batchUpload 批量上传
// 规则如下: 时间超过 slienceTimeout 未更新的, 且远端不存在的上传
func (u *Uploader) batchUpload(ignoreState bool) (*batchUploadResult, error) {
	result := &batchUploadResult{}

	err := filepath.Walk(u.dataDir, func(path string, info fs.FileInfo, err error) error {
		// 目录/文件不存在等
		if err != nil {
			return err
		}

		filePath := path[len(u.dataDir):]

		if info.IsDir() {
			return nil
		}

		if !castPattern.MatchString(filePath) {
			result.ignore = append(result.ignore, filePath)
			return nil
		}

		if !ignoreState && !u.checkFileCanUpload(filePath) {
			result.ignore = append(result.ignore, filePath)
			return nil
		}

		cast := &castFile{
			absPath:  path,
			filePath: filePath,
			name:     info.Name(),
			dir:      filepath.Dir(filePath),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ok, e := u.storage.IsExist(ctx, filePath)
		if e != nil {
			blog.Errorf("check file exist err: %s", e)
			result.failed = append(result.failed, cast.filePath)
			return nil
		}

		if ok {
			// 已经上传的清理本地文件， 节省空间
			cast.clean()
			result.success = append(result.success, cast.filePath)
			return nil
		}

		if err := u.upload(cast); err != nil {
			result.failed = append(result.failed, cast.filePath)
			blog.Errorf("upload file err: %s", err)
			return nil
		}

		result.success = append(result.success, cast.filePath)
		return nil
	})

	return result, err
}

// Upload 单个文件上传, filePath格式: {date}/{fileName}
func (u *Uploader) upload(cast *castFile) error {

	// 10 分钟上传时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	u.setState(cast.filePath, uploadingState)
	if err := u.storage.UploadFile(ctx, cast.absPath, cast.filePath); err != nil {
		blog.Errorf("upload file %s failed, err: %s", cast.filePath, err)
		return err
	}

	u.setState(cast.filePath, uploadedState)
	// 清理文件
	cast.clean()

	blog.Infof("upload file %s success", cast.filePath)

	return nil
}

// 删除文件夹
func (u *Uploader) deleteFolders() error {

	// 1 分钟删除时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	folderNames, err := u.storage.ListFolders(ctx, "")
	if err != nil {
		return fmt.Errorf("list folders failed, err: %w", err)
	}

	var expireFolders []string
	var invalidFolders []string
	reserveDay := time.Duration(config.G.Audit.RetentionDays)
	for _, v := range folderNames {
		v = strings.Trim(v, "/")
		date, err := time.Parse(dateTimeFormat, v)
		if err != nil {
			invalidFolders = append(invalidFolders, v)
			continue
		}
		// 删除 RetentionDays 天前的文件夹
		if date.Before(time.Now().Add(-time.Hour * 24 * reserveDay)) {
			expireFolders = append(expireFolders, v)
			if err := u.storage.DeleteFolders(ctx, v); err != nil {
				return fmt.Errorf("delete folders %s failed, err: %w", v, err)
			}
		}
	}

	blog.Infof("delete folders success: %v, invalid folders: %v", expireFolders, invalidFolders)

	return nil
}
