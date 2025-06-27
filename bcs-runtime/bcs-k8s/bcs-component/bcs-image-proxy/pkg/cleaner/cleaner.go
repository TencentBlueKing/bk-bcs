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

// Package cleaner xxx
package cleaner

import (
	"bufio"
	"context"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/recorder"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

// ImageCleaner defines the cleander of image
type ImageCleaner struct {
	op      *options.ImageProxyOption
	cronObj *cron.Cron
}

var (
	globalCleaner *ImageCleaner
)

// GlobalCleaner return the global cleaner object
func GlobalCleaner() *ImageCleaner {
	if globalCleaner != nil {
		return globalCleaner
	}
	globalCleaner = &ImageCleaner{
		op: options.GlobalOptions(),
	}
	return globalCleaner
}

// Init the image cleaner
func (r *ImageCleaner) Init() error {
	if r.op.CleanConfig.Cron == "" {
		return nil
	}

	r.cronObj = cron.New()
	_, err := r.cronObj.AddFunc(r.op.CleanConfig.Cron, func() {
		blog.Infof("[clean] auto-clean is started")
		needClean, err := r.cleanCheck()
		if err != nil {
			blog.Errorf("[clean] clean check failed: %s", err.Error())
			return
		}
		blog.Infof("[clean] clean check success")
		if !needClean {
			return
		}
		if err = r.handleClean(); err != nil {
			blog.Errorf("[clean] clean handle failed: %s", err.Error())
			return
		}
		blog.Infof("[clean] clean handle completed")
	})
	if err != nil {
		return errors.Wrapf(err, "init image-cleaner failed")
	}
	return nil
}

// Run the image cleaner
func (r *ImageCleaner) Run(ctx context.Context) {
	r.cronObj.Start()
	defer r.cronObj.Stop()
	select {
	case <-ctx.Done():
		return
	}
}

func (r *ImageCleaner) calcPathSize(path string) (int64, error) {
	blog.Infof("[clean] calculating path '%s' size...", path)
	storageSize, err := utils.ConcurrentDirSize(path)
	if err != nil {
		return 0, errors.Wrapf(err, "get path '%s' size failed", path)
	}
	blog.Infof("[clean] path '%s' size: %d", path, storageSize)
	return storageSize, nil
}

func (r *ImageCleaner) cleanCheck() (bool, error) {
	storageSize, err := r.calcPathSize(r.op.StoragePath)
	if err != nil {
		return false, errors.Wrapf(err, "calc storage path size failed")
	}
	torrentSize, err := r.calcPathSize(r.op.TorrentPath)
	if err != nil {
		return false, errors.Wrapf(err, "calc torrent path size failed")
	}
	transferSize, err := r.calcPathSize(r.op.TransferPath)
	if err != nil {
		return false, errors.Wrapf(err, "calc transfer path size failed")
	}
	smallSize, err := r.calcPathSize(r.op.SmallFilePath)
	if err != nil {
		return false, errors.Wrapf(err, "calc smallfile path size failed")
	}
	ociSize, err := r.calcPathSize(r.op.OCIPath)
	if err != nil {
		return false, errors.Wrapf(err, "calc oci path size failed")
	}
	allSize := storageSize + torrentSize + transferSize + smallSize + ociSize
	threshold := r.op.CleanConfig.Threshold * 1000 * 1000 * 1000
	if allSize < threshold {
		blog.Infof("[clean] current-size: %s, threshold: %s, no-need clean", humanize.Bytes(uint64(allSize)),
			humanize.Bytes(uint64(threshold)))
		return false, nil
	}
	return true, nil
}

func (r *ImageCleaner) handleClean() error {
	if r.op.CleanConfig.RetainDays == 0 {
		blog.Infof("[clean] clean.retainDays is 0, no need clean")
		return nil
	}
	blog.Infof("[clean] start counting the layer used in the last %d days", r.op.CleanConfig.RetainDays)
	fi, err := os.Open(r.op.EventFile)
	if err != nil {
		return errors.Wrapf(err, "open event file failed")
	}
	defer fi.Close()

	retainTime := time.Now().Add(-time.Duration(r.op.CleanConfig.RetainDays) * 24 * time.Hour)
	retainLayers := make(map[string]struct{})

	scanner := bufio.NewScanner(fi)
	var lines int64 = 0
	var currentBatch int64 = 0
	var batchSize int64 = 200
	for scanner.Scan() {
		lines++
		batch := lines / batchSize
		if batch < currentBatch {
			continue
		}
		line := scanner.Text()
		if line == "" {
			continue
		}
		event := new(recorder.Event)
		if err = jsoniter.Unmarshal([]byte(line), event); err != nil {
			continue
		}
		currentBatch++

		// 检测时间是否超过
		if event.CreatedAt.After(retainTime) {
			if event.Digest != "" {
				retainLayers[event.Digest] = struct{}{}
			}
			break
		}
	}
	blog.Infof("[clean] check that the current line number is the retention time: %d", lines)

	for scanner.Scan() {
		line := scanner.Text()
		event := new(recorder.Event)
		if err = jsoniter.Unmarshal([]byte(line), event); err != nil {
			continue
		}
		if event.Digest != "" {
			retainLayers[event.Digest] = struct{}{}
		}
	}
	blog.Infof("[clean] detect the number of layers that need to be retained: %d", len(retainLayers))
	r.doClean(r.op.StoragePath, retainLayers)
	r.doClean(r.op.TransferPath, retainLayers)
	r.doClean(r.op.OCIPath, retainLayers)
	r.doClean(r.op.SmallFilePath, retainLayers)

	needClean, err := r.cleanCheck()
	if err != nil {
		return errors.Wrapf(err, "second check failed")
	}
	if !needClean {
		return nil
	}
	blog.Warnf("[clean] need deep clean, but not implement")
	return nil
}

func (r *ImageCleaner) doClean(storeDir string, retainLayers map[string]struct{}) {
	blog.Infof("[clean] start clen dir: %s", storeDir)
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		blog.Errorf("[clean] read dir %s failed: %v", storeDir, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".tar.gzip") {
			continue
		}
		digest := strings.TrimSuffix(fileName, ".tar.gzip")
		if _, ok := retainLayers[digest]; ok {
			continue
		}
		fullPath := path.Join(storeDir, entry.Name())
		info, _ := entry.Info()
		size := humanize.Bytes(uint64(info.Size()))
		if err = os.RemoveAll(fullPath); err != nil {
			blog.Errorf("[clean] remove file '%s' failed: %s", fullPath, err.Error())
		} else {
			blog.Infof("[clean] remove file '%s' success, size: %s", fullPath, size)
		}
	}
	blog.Infof("[clean] clean dir %s finished", storeDir)
}
