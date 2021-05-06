/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-ini/ini"

	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

var (
	// bufferSize is file writer buffer.
	bufferSize = 1024 * 4

	// threshold of oldest cache clean.
	oldestCacheTimeThreshold = int64(24 * time.Hour / time.Second)
)

// Content is configs content.
type Content struct {
	// content id.
	ContentID string

	// content size.
	ContentSize uint64

	// config content metadata.
	Metadata []byte
}

const (
	// content cache info.
	contentCacheInfoFileName = "content.info"

	// content file cache lock file.
	contentLockFile = "content.lock"

	// content metadata cache file.
	contentFileName = "content.metadata"

	// content cached time information.
	contentCacheInfoCachedTime = "cachedTime"

	// content cached size information.
	contentCacheInfoSize = "size"

	// linkDownloadCost is download cost.
	linkDownloadCost = "cost"

	// link file cache lock file.
	linkLockFile = "link.lock"

	// link file metadata.
	linkFileName = "link.metadata"
)

func contentCacheContentPath(contentCachePath, contentID string) string {
	return fmt.Sprintf("%s/%s", contentCachePath, contentID)
}

func contentCacheContentInfoFile(contentCachePath, contentID string) string {
	return fmt.Sprintf("%s/%s/%s", contentCachePath, contentID, contentCacheInfoFileName)
}

func contentCacheLockFile(contentCachePath, contentID string) string {
	return fmt.Sprintf("%s/.%s.%s", contentCachePath, contentID, contentLockFile)
}

func contentCacheContentFile(contentCachePath, contentID string) string {
	return fmt.Sprintf("%s/%s/%s", contentCachePath, contentID, contentFileName)
}

func contentCacheContentPreFile(path, contentID string) string {
	return fmt.Sprintf("%s/.%s.pre", path, contentID)
}

func contentCacheLinkLockFile(linkContentCachePath, contentID string) string {
	return fmt.Sprintf("%s/.%s.%s", linkContentCachePath, contentID, linkLockFile)
}

func contentCacheLinkFile(linkContentCachePath, contentID string) string {
	return fmt.Sprintf("%s/%s.%s", linkContentCachePath, contentID, linkFileName)
}

// ContentCache is release config content cache.
type ContentCache struct {
	viper *safeviper.SafeViper

	bizID string
	appID string
	path  string

	// content file cache path.
	contentCachePath string

	// configs link download cache path.
	linkContentCachePath string
}

// NewContentCache creates a new ContentCache.
func NewContentCache(viper *safeviper.SafeViper, bizID, appID, path,
	contentCachePath, linkContentCachePath string) *ContentCache {

	os.MkdirAll(contentCachePath, os.ModePerm)
	os.MkdirAll(linkContentCachePath, os.ModePerm)

	return &ContentCache{
		viper:                viper,
		bizID:                bizID,
		appID:                appID,
		path:                 path,
		contentCachePath:     contentCachePath,
		linkContentCachePath: linkContentCachePath,
	}
}

// Add adds a new config effected release content to cache by source link.
func (c *ContentCache) Add(content *Content) error {
	if content == nil {
		return errors.New("invalid content: nil")
	}
	if err := os.MkdirAll(c.linkContentCachePath, os.ModePerm); err != nil {
		return err
	}

	fl, err := LockFile(contentCacheLinkLockFile(c.linkContentCachePath, content.ContentID), true)
	if err != nil {
		return err
	}
	defer UnlockFile(fl)

	isCached, err := c.Has(content.ContentID)
	if err != nil {
		return err
	}
	if isCached {
		return nil
	}

	// download and cache now.
	newLinkFile := contentCacheLinkFile(c.linkContentCachePath, content.ContentID)

	// TODO: dynamic download concurrent and limits.

	contentLinkURL := fmt.Sprintf("http://%s:%d/%s/%s",
		c.viper.GetString("gateway.hostName"), c.viper.GetInt("gateway.port"),
		c.viper.GetString("gateway.fileContetAPIPath"), c.bizID)

	option := &DownloadConfigsOption{
		URL:                contentLinkURL,
		ContentID:          content.ContentID,
		NewFile:            newLinkFile,
		Concurrent:         c.viper.GetInt("cache.downloadPerFileConcurrent"),
		LimitBytesInSecond: c.viper.GetInt64("cache.downloadPerFileLimitBytesInSecond"),
	}

	downloadCost, err := DownloadConfigs(option, c.viper.GetDuration("cache.downloadTimeout"))
	if err != nil {
		return err
	}
	logger.Warn("ContentCache[%s %s %s]| download link file[%s] success, cost: %+v",
		c.bizID, c.appID, c.path, contentLinkURL, downloadCost)

	// rename to target content cache.
	if err := os.MkdirAll(contentCacheContentPath(c.contentCachePath, content.ContentID), os.ModePerm); err != nil {
		return err
	}

	flContent, err := LockFile(contentCacheLockFile(c.contentCachePath, content.ContentID), true)
	if err != nil {
		return err
	}
	defer UnlockFile(flContent)

	// content temp file sign.
	fileCid, err := common.FileSHA256(newLinkFile)
	if err != nil {
		return err
	}
	if fileCid != content.ContentID {
		return fmt.Errorf("inconsistent cid[%+v][%+v], %s", content.ContentID, fileCid, newLinkFile)
	}
	newLinkFileInfo, err := os.Stat(newLinkFile)
	if err != nil {
		return err
	}

	// rename content link file to real cache file.
	if err := os.Rename(newLinkFile, contentCacheContentFile(c.contentCachePath, content.ContentID)); err != nil {
		return err
	}

	// write content information.
	contentInfoFile := contentCacheContentInfoFile(c.contentCachePath, content.ContentID)
	info, err := ini.LooseLoad(contentInfoFile)
	if err != nil {
		return err
	}
	if _, err := info.Section("").NewKey(contentCacheInfoSize,
		common.ToStr(int(newLinkFileInfo.Size()))); err != nil {
		return err
	}
	if _, err := info.Section("").NewKey(contentCacheInfoCachedTime,
		time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return err
	}
	if _, err := info.Section("").NewKey(linkDownloadCost,
		fmt.Sprintf("%+v", downloadCost)); err != nil {
		return err
	}
	// save local content cache details.
	if err := info.SaveTo(contentInfoFile); err != nil {
		return err
	}
	return nil
}

// Has checks whether the target cid content exists or not.
func (c *ContentCache) Has(contentID string) (bool, error) {
	fl, err := LockFile(contentCacheLockFile(c.contentCachePath, contentID), true)
	if err != nil {
		return false, err
	}
	defer UnlockFile(fl)

	return c.has(contentID)
}

// has checks whether the target cid content exists in local file cache.
func (c *ContentCache) has(contentID string) (bool, error) {
	info, err := ini.LooseLoad(contentCacheContentInfoFile(c.contentCachePath, contentID))
	if err != nil {
		return false, err
	}

	// cache time.
	if info.Section("").Key(contentCacheInfoCachedTime).String() == "" {
		return false, nil
	}

	// cache size.
	if _, err := info.Section("").Key(contentCacheInfoSize).Uint64(); err != nil {
		return false, err
	}

	// check content file sign.
	fileCid, err := common.FileSHA256(contentCacheContentFile(c.contentCachePath, contentID))
	if err != nil {
		return false, err
	}

	if fileCid != contentID {
		logger.Warn("ContentCache[%s %s %s]| has, inconsistent cid[%+v][%+v]",
			c.bizID, c.appID, c.path, contentID, fileCid)
		return false, nil
	}
	return true, nil
}

// realConfigName returns real config name.
func (c *ContentCache) realConfigName(path, name string) string {
	return fmt.Sprintf("%s/%s", path, name)
}

// Effect effects a release by cid in content cache.
func (c *ContentCache) Effect(contentID, name, path string, option *PermissionOption) error {
	fl, err := LockFile(contentCacheLockFile(c.contentCachePath, contentID), true)
	if err != nil {
		return err
	}
	defer UnlockFile(fl)

	// content cache file md5 sign.
	contentFile := contentCacheContentFile(c.contentCachePath, contentID)
	fileCid, err := common.FileSHA256(contentFile)
	if err != nil {
		return err
	}
	if contentID != fileCid {
		return errors.New("invalid content, can't effect this cache")
	}

	fCache, err := os.OpenFile(contentFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer fCache.Close()

	// app real config pre-file.
	preFile := contentCacheContentPreFile(path, contentID)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	fConfig, err := os.OpenFile(preFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fConfig.Close()

	// copy content cache with buffer.
	buf := make([]byte, bufferSize)
	for {
		n, err := fCache.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fConfig.Write(buf[:n])
	}

	// content cache pre-file md5 sign.
	preFileCid, err := common.FileSHA256(preFile)
	if err != nil {
		return err
	}
	if contentID != preFileCid {
		return errors.New("invalid cid of pre-file")
	}
	configName := c.realConfigName(path, name)

	logger.Warn("ContentCache[%s %s %s]| Effect the real configs now, configName[%s] preFile[%s]",
		c.bizID, c.appID, c.path, configName, preFile)

	// flush config file permissions.
	if err := FlushFilePermission(preFile, option); err != nil {
		return fmt.Errorf("flush file permission options, %+v", err)
	}

	// rename pre-file to real app config file.
	if err := os.Rename(preFile, configName); err != nil {
		return fmt.Errorf("rename target file, %+v", err)
	}
	return nil
}

// ContentCacheCleaner is content cache cleaner.
type ContentCacheCleaner struct {
	viper *safeviper.SafeViper

	// content file cache path.
	contentCachePath string

	// expired cache path.
	expiredPath string

	// max disk usage rate of content cache.
	contentCacheMaxDiskUsageRate int

	// expiration of content cache.
	contentCacheExpiration time.Duration

	// disk usage status check interval.
	diskUsageCheckInterval time.Duration
}

// NewContentCacheCleaner creates a new ContentCacheCleaner instance.
func NewContentCacheCleaner(viper *safeviper.SafeViper, contentCachePath, expiredPath string, contentCacheMaxDiskUsageRate int,
	contentCacheExpiration, diskUsageCheckInterval time.Duration) *ContentCacheCleaner {

	return &ContentCacheCleaner{
		viper:                        viper,
		contentCachePath:             contentCachePath,
		expiredPath:                  expiredPath,
		contentCacheMaxDiskUsageRate: contentCacheMaxDiskUsageRate,
		contentCacheExpiration:       contentCacheExpiration,
		diskUsageCheckInterval:       diskUsageCheckInterval,
	}
}

func (c *ContentCacheCleaner) expiredFile(contentID string) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s.%d%d%d-%d%d%d",
		c.expiredPath, contentID, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}

// purge cleans expired cache, invalid content id cache, and force cleans when it exceed
// the threshold of max disk usage.
func (c *ContentCacheCleaner) purge(isExceedThreshold bool) error {
	root, err := ioutil.ReadDir(c.contentCachePath)
	if err != nil {
		return err
	}

	// oldest cache used to clean dist usage.
	var oldestCacheSize int64
	var oldestCacheTime int64
	var oldestCacheContentID string

	// range the content cache path.
	for _, fContent := range root {
		if !fContent.IsDir() {
			continue
		}
		contentID := fContent.Name()
		needPurge := false

		// lock target content cache first.
		fl, err := LockFile(contentCacheLockFile(c.contentCachePath, contentID), true)
		if err != nil {
			logger.Warn("ContentCacheCleaner| clean content cache[%s], flock %+v", contentID, err)
			continue
		}

		// check file info of the cache metadata.
		contentFile := contentCacheContentFile(c.contentCachePath, contentID)
		fInfo, err := os.Stat(contentFile)
		if err != nil {
			UnlockFile(fl)
			continue
		}
		cacheSize := fInfo.Size()
		cachedTime := fInfo.ModTime().Unix()

		// compare the expired time.
		if time.Now().Unix()-cachedTime >= int64(c.contentCacheExpiration/time.Second) {
			needPurge = true
		} else {
			// compare the sha256 id of cache file.
			fileCid, err := common.FileSHA256(contentFile)
			if err != nil {
				UnlockFile(fl)
				logger.Warn("ContentCacheCleaner| clean content cache[%s], can't cal content id, %+v", contentID, err)
				continue
			}
			if fileCid != contentID {
				logger.Warn("ContentCacheCleaner| clean content cache[%s], need purge invalid content id %s",
					contentID, fileCid)
				needPurge = true
			}
		}

		if !needPurge {
			UnlockFile(fl)
			logger.V(4).Infof("ContentCacheCleaner| content cache[%s] size[%d] cached time[%d], not need to purge",
				contentID, cacheSize, cachedTime)

			// compare the oldest cache.
			if (oldestCacheTime == 0 || cachedTime < oldestCacheTime) ||
				(cachedTime == oldestCacheTime && cacheSize > oldestCacheSize) {

				oldestCacheTime = cachedTime
				oldestCacheSize = cacheSize
				oldestCacheContentID = contentID
			}
			continue
		}

		// rename the cache to expired path.
		if err := os.Rename(contentCacheContentPath(c.contentCachePath, contentID),
			c.expiredFile(contentID)); err != nil {
			UnlockFile(fl)
			logger.Warn("ContentCacheCleaner| clean content cache[%s], remove failed, %+v", contentID, err)
			continue
		}
		UnlockFile(fl)
		logger.Warn("ContentCacheCleaner| clean content cache[%s] success", contentID)
	}

	// clean disk usage.
	logger.V(2).Infof("ContentCacheCleaner| oldest and largest cache[%s] size[%d] cached time[%d], threshold[%+v]",
		oldestCacheContentID, oldestCacheSize, oldestCacheTime, isExceedThreshold)

	if isExceedThreshold {
		if (time.Now().Unix() - oldestCacheTime) < oldestCacheTimeThreshold {
			// the cache is not oldest enough.
			logger.V(2).Infof("ContentCacheCleaner| oldest and largest cache[%s] size[%d] cached time[%d], "+
				"exceed threshold[%+v] not old enough",
				oldestCacheContentID, oldestCacheSize, oldestCacheTime, isExceedThreshold)
			return nil
		}

		// oldest enough, clean now, and lock first.
		fl, err := LockFile(contentCacheLockFile(c.contentCachePath, oldestCacheContentID), true)
		if err != nil {
			return fmt.Errorf("clean oldest and largest cache[%s], flock %+v", oldestCacheContentID, err)
		}
		defer UnlockFile(fl)

		// rename the cache to expired path.
		if err := os.Rename(contentCacheContentPath(c.contentCachePath, oldestCacheContentID),
			c.expiredFile(oldestCacheContentID)); err != nil {
			return fmt.Errorf("clean oldest and largest cache[%s] to expired path, flock %+v",
				oldestCacheContentID, err)
		}
	}

	return nil
}

func (c *ContentCacheCleaner) diskUsageCheck() error {
	// get disk usage status of content cache path.
	status, err := common.DiskUsage(c.contentCachePath)
	if err != nil {
		return err
	}
	if status.All == 0 {
		return fmt.Errorf("can't get disk usage status, %+v", status)
	}

	// get content cache dir size.
	cacheSize, err := common.StatDirectoryFileSize(c.contentCachePath)
	if err != nil {
		return fmt.Errorf("can't get content cache dir size, %+v", err)
	}

	usedRate := int((float64(cacheSize) / float64(status.All)) * 100)
	logger.V(2).Infof("ContentCacheCleaner| current content cache disk usage rate: %d", usedRate)

	return c.purge(usedRate > c.contentCacheMaxDiskUsageRate)
}

// Run setups and runs the content cache cleaner.
func (c *ContentCacheCleaner) Run() {
	os.MkdirAll(c.contentCachePath, os.ModePerm)
	os.MkdirAll(c.expiredPath, os.ModePerm)

	// disk usage status check.
	ticker := time.NewTicker(c.diskUsageCheckInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := c.diskUsageCheck(); err != nil {
			logger.Errorf("check content cache disk usage status failed, %+v", err)
		}
	}
}
