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
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-ini/ini"
	"github.com/gofrs/flock"
	"github.com/spf13/viper"

	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// ReleaseMetadata is release metadata struct.
type ReleaseMetadata struct {
	// configset id.
	Cfgsetid string

	// config set name.
	CfgsetName string

	// config set fpath.
	CfgsetFpath string

	// serial num of release.
	Serialno uint64

	// release id.
	Releaseid string

	// content id.
	Cid string

	// content link.
	CfgLink string

	// release effect time.
	EffectTime string

	// release name.
	ReleaseName string

	// multi release id.
	MultiReleaseid string

	isRollback bool
}

// ConfigSet is struct for one config set which marked the current effected release.
type ConfigSet struct {
	// configset id.
	cfgsetid string

	// current release metadata.
	current *ReleaseMetadata
}

// NewConfigSet creates new ConfigSet.
func NewConfigSet(cfgsetid string) *ConfigSet {
	return &ConfigSet{cfgsetid: cfgsetid}
}

// Effect saves new effected release metadata.
func (set *ConfigSet) Effect(metadata *ReleaseMetadata) error {
	if metadata == nil {
		return errors.New("invalid metadata: nil")
	}
	set.current = metadata
	return nil
}

// Current returns newest release which is effected at the moment.
func (set *ConfigSet) Current() *ReleaseMetadata {
	return set.current
}

const (
	// release details file name.
	releaseDetailsFileName = "release.details"

	// release config lock file name.
	releaseConfigLockFileName = "release.lock"

	// cfgsetid in details.
	detailsCfgsetid = "cfgsetid"

	// cfgset name in details.
	detailsCfgsetName = "cfgsetname"

	// cfgset fpath in details.
	detailsCfgsetFpath = "cfgsetpath"

	// serialno in details.
	detailsSerialno = "serialno"

	// releaseid in details.
	detailsReleaseid = "releaseid"

	// content id in details.
	detailsCid = "cid"

	// content link in details.
	detailsCfgLink = "cfglink"

	// release effected time.
	detailsEffectTime = "effecttime"

	// release name.
	detailsReleaseName = "releasename"

	// multi releaseid.
	detailsMultiReleaseid = "multireleaseid"

	// release event type.
	detailsIsRollback = "rollback"
)

// EffectCache is config release effect cache.
type EffectCache struct {
	businessName string
	appName      string
	path         string

	// config sets metadatas, cfgsetid -> ConfigSet.
	configSets map[string]*ConfigSet

	// mu make ops on configSets safely.
	mu sync.RWMutex

	// path of file cache.
	fileCachePath string
}

// NewEffectCache creates new EffectCache.
func NewEffectCache(fileCachePath, businessName, appName, path string) *EffectCache {
	return &EffectCache{
		fileCachePath: fileCachePath,
		businessName:  businessName,
		appName:       appName,
		path:          path,
		configSets:    make(map[string]*ConfigSet),
	}
}

func (c *EffectCache) configSetPath(cfgsetid string) string {
	return fmt.Sprintf("%s/%s", c.fileCachePath, cfgsetid)
}

func (c *EffectCache) detailsFile(cfgsetid string) string {
	return fmt.Sprintf("%s/%s/%s", c.fileCachePath, cfgsetid, releaseDetailsFileName)
}

func (c *EffectCache) lockFile(cfgsetid string) string {
	return fmt.Sprintf("%s/%s/.%s", c.fileCachePath, cfgsetid, releaseConfigLockFileName)
}

// writeDetails writes relese details to file cache.
func (c *EffectCache) writeDetails(metadata *ReleaseMetadata) error {
	if err := os.MkdirAll(c.configSetPath(metadata.Cfgsetid), os.ModePerm); err != nil {
		return err
	}

	fl := flock.New(c.lockFile(metadata.Cfgsetid))
	locked, err := fl.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("can't get flock, try again later")
	}
	defer fl.Unlock()

	details, err := ini.LooseLoad(c.detailsFile(metadata.Cfgsetid))
	if err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsCfgsetid, metadata.Cfgsetid); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsCfgsetName, metadata.CfgsetName); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsCfgsetFpath, metadata.CfgsetFpath); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsReleaseid, metadata.Releaseid); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsCid, metadata.Cid); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsCfgLink, metadata.CfgLink); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsSerialno, fmt.Sprintf("%d", metadata.Serialno)); err != nil {
		return err
	}

	if _, err := details.Section("").NewKey(detailsEffectTime, metadata.EffectTime); err != nil {
		return err
	}

	details.Section("").NewKey(detailsReleaseName, metadata.ReleaseName)
	details.Section("").NewKey(detailsMultiReleaseid, metadata.MultiReleaseid)
	details.Section("").NewKey(detailsIsRollback, fmt.Sprintf("%v", metadata.isRollback))

	if err := details.SaveTo(c.detailsFile(metadata.Cfgsetid)); err != nil {
		return err
	}
	return nil
}

// readDetails reads relese details from file cache.
func (c *EffectCache) readDetails(cfgsetid string) (*ReleaseMetadata, error) {
	if err := os.MkdirAll(c.configSetPath(cfgsetid), os.ModePerm); err != nil {
		return nil, err
	}

	fl := flock.New(c.lockFile(cfgsetid))
	locked, err := fl.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("can't get flock, try again later")
	}
	defer fl.Unlock()

	details, err := ini.LooseLoad(c.detailsFile(cfgsetid))
	if err != nil {
		return nil, err
	}

	dCfgsetid := details.Section("").Key(detailsCfgsetid).String()
	if dCfgsetid == "" {
		return nil, errors.New("invalid detail:cfgsetid")
	}

	cfgsetName := details.Section("").Key(detailsCfgsetName).String()
	if cfgsetName == "" {
		return nil, errors.New("invalid detail:cfgsetname")
	}
	cfgsetFpath := details.Section("").Key(detailsCfgsetFpath).String()

	releaseid := details.Section("").Key(detailsReleaseid).String()
	if releaseid == "" {
		return nil, errors.New("invalid detail:releaseid")
	}

	cid := details.Section("").Key(detailsCid).String()
	if cid == "" {
		return nil, errors.New("invalid detail:cid")
	}
	cfgLink := details.Section("").Key(detailsCfgLink).String()

	effectTime := details.Section("").Key(detailsEffectTime).String()
	if effectTime == "" {
		return nil, errors.New("invalid detail:effecttime")
	}

	serialno, err := details.Section("").Key(detailsSerialno).Uint64()
	if err != nil {
		return nil, err
	}
	releaseName := details.Section("").Key(detailsReleaseName).String()
	multiReleaseid := details.Section("").Key(detailsMultiReleaseid).String()
	isRollback, _ := details.Section("").Key(detailsIsRollback).Bool()

	metadata := &ReleaseMetadata{
		Cfgsetid:       dCfgsetid,
		CfgsetName:     cfgsetName,
		CfgsetFpath:    cfgsetFpath,
		Serialno:       serialno,
		Releaseid:      releaseid,
		Cid:            cid,
		CfgLink:        cfgLink,
		EffectTime:     effectTime,
		ReleaseName:    releaseName,
		MultiReleaseid: multiReleaseid,
		isRollback:     isRollback,
	}

	return metadata, nil
}

// Effect adds effected release.
func (c *EffectCache) Effect(metadata *ReleaseMetadata) error {
	if metadata == nil {
		return errors.New("invalid metadata: nil")
	}

	if err := c.writeDetails(metadata); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.configSets[metadata.Cfgsetid]; !ok {
		c.configSets[metadata.Cfgsetid] = NewConfigSet(metadata.Cfgsetid)
	}

	configSet := c.configSets[metadata.Cfgsetid]
	return configSet.Effect(metadata)
}

// LocalRelease returns local effected release information of target configset.
func (c *EffectCache) LocalRelease(cfgsetid string) (*ReleaseMetadata, error) {
	c.mu.RLock()
	configSet, ok := c.configSets[cfgsetid]
	if ok && configSet != nil {
		c.mu.RUnlock()
		return configSet.Current(), nil
	}
	c.mu.RUnlock()

	md, err := c.readDetails(cfgsetid)
	if err != nil {
		logger.Warn("EffectCache[%s %s %s]| suppose no effected release of cfgsetid[%+v], [suppose base on detail info[%+v]]",
			c.businessName, c.appName, c.path, cfgsetid, err)
		return nil, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.configSets[md.Cfgsetid]; !ok {
		c.configSets[md.Cfgsetid] = NewConfigSet(md.Cfgsetid)
	}

	configSet = c.configSets[md.Cfgsetid]
	configSet.Effect(md)

	return md, nil
}

// NeedEffected checks whether need to effect the release.
func (c *EffectCache) NeedEffected(cfgsetid string, serialno uint64) (bool, error) {
	md, err := c.LocalRelease(cfgsetid)
	if err != nil {
		return false, err
	}

	if md == nil {
		return true, nil
	}

	if serialno <= md.Serialno {
		return false, nil
	}

	return true, nil
}

// Debug returns debug information.
func (c *EffectCache) Debug(cfgsetid string) string {
	md, err := c.LocalRelease(cfgsetid)
	if err != nil {
		return fmt.Sprintf("can't get cfgsetid[%s] debug information, %s", cfgsetid, err.Error())
	}

	if md == nil {
		return fmt.Sprintf("cfgsetid[%+v] no effected release", cfgsetid)
	}

	return fmt.Sprintf("cfgsetid[%+v] localRelease[%+v] localSerialno[%+v] effecttime[%+v]",
		cfgsetid, md.Releaseid, md.Serialno, md.EffectTime)
}

var (
	// bufferSize is file writer buffer.
	bufferSize = 1024 * 1024 * 8 // 8M
)

// Content is configs content.
type Content struct {
	// content id.
	Cid string

	// content link.
	CfgLink string

	// config content metadata.
	Metadata []byte
}

const (
	// content cache info.
	contentCacheInfoFileName = "content.info"

	// content file cache lock file.
	contentLockFile = "content.lock"

	// configs metadata cache file.
	configsFileName = "content.metadata"

	// content cached time information.
	contentCacheInfoCachedTime = "cachedtime"

	// content cached size information.
	contentCacheInfoSize = "size"
)

// ContentCache is release config content cache.
type ContentCache struct {
	viper *viper.Viper

	businessName string
	appName      string
	path         string

	// content file cache path.
	contentFilePath string

	// memory LRU content information cache.
	mcache gcache.Cache

	// expired cache path.
	expiredPath string

	// expiration of memory information cache.
	mcacheExpiration time.Duration

	// expiration of content cache.
	contentCacheExpiration time.Duration

	// purge interval.
	purgeInterval time.Duration
}

// NewContentCache creates a new ContentCache.
func NewContentCache(viper *viper.Viper, contentFilePath, businessName, appName, path string, mcacheSize int, expiredPath string,
	mcacheExpiration, contentCacheExpiration, purgeInterval time.Duration) *ContentCache {
	return &ContentCache{
		viper:                  viper,
		businessName:           businessName,
		appName:                appName,
		path:                   path,
		contentFilePath:        contentFilePath,
		expiredPath:            expiredPath,
		mcacheExpiration:       mcacheExpiration,
		contentCacheExpiration: contentCacheExpiration,
		purgeInterval:          purgeInterval,
		mcache:                 gcache.New(mcacheSize).EvictType(gcache.TYPE_LRU).Build(),
	}
}

func (c *ContentCache) contentPath(cid string) string {
	return fmt.Sprintf("%s/%s", c.contentFilePath, cid)
}

func (c *ContentCache) expiredFile(cid string) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s.%d%d%d-%d%d%d",
		c.expiredPath, cid, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}

func (c *ContentCache) contentInfoFile(cid string) string {
	return fmt.Sprintf("%s/%s/%s", c.contentFilePath, cid, contentCacheInfoFileName)
}

func (c *ContentCache) lockFile(cid string) string {
	return fmt.Sprintf("%s/%s/.%s", c.contentFilePath, cid, contentLockFile)
}

func (c *ContentCache) contentFile(cid string) string {
	return fmt.Sprintf("%s/%s/%s", c.contentFilePath, cid, configsFileName)
}

func (c *ContentCache) contentPreFile(path, cid string) string {
	return fmt.Sprintf("%s/.%s.pre", path, cid)
}

func (c *ContentCache) contentTempFile(cid string) string {
	return fmt.Sprintf("%s/%s/%s.tmp", c.contentFilePath, cid, configsFileName)
}

// Add adds a new config effected release content to cache.
func (c *ContentCache) Add(content *Content) error {
	if content == nil {
		return errors.New("invalid content: nil")
	}

	if err := os.MkdirAll(c.contentPath(content.Cid), os.ModePerm); err != nil {
		return err
	}

	fl := flock.New(c.lockFile(content.Cid))
	locked, err := fl.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("can't get flock, try again later")
	}
	defer fl.Unlock()

	// write temp content data.
	fConfig, err := os.OpenFile(c.contentTempFile(content.Cid), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fConfig.Close()

	if _, err := fConfig.WriteString(string(content.Metadata)); err != nil {
		return err
	}

	// content temp file sign.
	fileCid, err := common.FileSHA256(c.contentTempFile(content.Cid))
	if err != nil {
		return err
	}
	if fileCid != content.Cid {
		return fmt.Errorf("inconsistent cid[%+v][%+v]", content.Cid, fileCid)
	}

	// rename content temp file to real cache file.
	if err := os.Rename(c.contentTempFile(content.Cid), c.contentFile(content.Cid)); err != nil {
		return err
	}

	// write content information.
	info, err := ini.LooseLoad(c.contentInfoFile(content.Cid))
	if err != nil {
		return err
	}

	if _, err := info.Section("").NewKey(contentCacheInfoSize,
		common.ToStr(len(content.Metadata))); err != nil {
		return err
	}

	if _, err := info.Section("").NewKey(contentCacheInfoCachedTime,
		time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return err
	}

	// save local content cache details.
	if err := info.SaveTo(c.contentInfoFile(content.Cid)); err != nil {
		return err
	}

	// set memory LRU information cache.
	c.mcache.SetWithExpire(content.Cid, true, c.mcacheExpiration)
	logger.Info("ContentCache[%s %s %s]| add new content cache success, %+v", c.businessName, c.appName, c.path, content.Cid)

	return nil
}

// Has checks whether the target cid content exists or not.
func (c *ContentCache) Has(cid string) (bool, error) {
	if c.mcache.Has(cid) {
		return true, nil
	}

	// final result base on local file cache.
	if err := os.MkdirAll(c.contentPath(cid), os.ModePerm); err != nil {
		return false, err
	}

	fl := flock.New(c.lockFile(cid))
	locked, err := fl.TryLock()
	if err != nil {
		return false, err
	}
	if !locked {
		return false, errors.New("can't get flock, try again later")
	}
	defer fl.Unlock()

	return c.has(cid)
}

// has checks whether the target cid content exists in local file cache.
func (c *ContentCache) has(cid string) (bool, error) {
	info, err := ini.LooseLoad(c.contentInfoFile(cid))
	if err != nil {
		return false, err
	}

	// cache time.
	if info.Section("").Key(contentCacheInfoCachedTime).String() == "" {
		return false, nil
	}

	// cache size.
	size, err := info.Section("").Key(contentCacheInfoSize).Uint64()
	if err != nil {
		return false, err
	}
	if size == 0 {
		return false, nil
	}

	// check content file sign.
	fileCid, err := common.FileSHA256(c.contentFile(cid))
	if err != nil {
		return false, err
	}

	if fileCid != cid {
		logger.Warn("ContentCache[%s %s %s]| has, inconsistent cid[%+v][%+v]", c.businessName, c.appName, c.path, cid, fileCid)
		return false, nil
	}

	// already exists.
	c.mcache.SetWithExpire(cid, true, c.mcacheExpiration)

	return true, nil
}

// realConfigName returns real config name.
func (c *ContentCache) realConfigName(path, name string) (string, error) {
	return fmt.Sprintf("%s/%s", path, name), nil
}

// Effect effects a release by cid in content cache.
func (c *ContentCache) Effect(cid, name, path string) error {
	if err := os.MkdirAll(c.contentPath(cid), os.ModePerm); err != nil {
		return err
	}

	fl := flock.New(c.lockFile(cid))
	locked, err := fl.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("can't get flock, try again later")
	}
	defer fl.Unlock()

	// content cache file md5 sign.
	fileCid, err := common.FileSHA256(c.contentFile(cid))
	if err != nil {
		return err
	}
	if cid != fileCid {
		return errors.New("invalid content, can't effect this cache")
	}

	// write app real config content.
	fCache, err := os.OpenFile(c.contentFile(cid), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer fCache.Close()

	// app real config pre-file.
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	preFile := c.contentPreFile(path, cid)

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
		fConfig.Write(buf[:n])
	}

	// content cache pre-file md5 sign.
	preFileCid, err := common.FileSHA256(preFile)
	if err != nil {
		return err
	}
	if cid != preFileCid {
		return errors.New("invalid cid of pre-file")
	}

	configName, err := c.realConfigName(path, name)
	if err != nil {
		return err
	}
	logger.Warn("ContentCache[%s %s %s]| Effect the real configs now, configName[%s] preFile[%s]",
		c.businessName, c.appName, c.path, configName, preFile)

	// rename pre-file to real app config file.
	if err := os.Rename(preFile, configName); err != nil {
		return err
	}
	return nil
}

// purge cleans up the local file content cache.
func (c *ContentCache) purge() error {
	if err := os.MkdirAll(c.expiredPath, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(c.contentFilePath, os.ModePerm); err != nil {
		return err
	}

	// range all content cache.
	root, err := ioutil.ReadDir(c.contentFilePath)
	if err != nil {
		return err
	}

	for _, fContent := range root {
		if !fContent.IsDir() {
			continue
		}

		fl := flock.New(c.lockFile(fContent.Name()))
		locked, err := fl.TryLock()
		if err != nil {
			logger.Warn("ContentCache[%s %s %s]| content cache purge, flock %+v", c.businessName, c.appName, c.path, err)
			continue
		}

		if !locked {
			continue
		}

		// get file stat information.
		fInfo, err := os.Stat(c.contentPath(fContent.Name()))
		if err != nil {
			logger.Warn("ContentCache[%s %s %s]| content cache purge, fstat %+v", c.businessName, c.appName, c.path, err)
			fl.Unlock()
			continue
		}

		// if need to purge the content cache which is expired or with invalid cid.
		var isNeedPurge bool

		// expiration checking.
		if time.Now().Unix()-fInfo.ModTime().Unix() >= int64(c.contentCacheExpiration/time.Second) {
			isNeedPurge = true
		} else {
			// content checking.
			fileCid, err := common.FileSHA256(c.contentFile(fContent.Name()))
			if err != nil {
				logger.Warn("ContentCache[%s %s %s]| can't cal content cid in purge, %+v", c.businessName, c.appName, c.path, err)
				fl.Unlock()
				continue
			}

			if fileCid != fContent.Name() {
				logger.Warn("ContentCache[%s %s %s]| find invalid cid[%+v]-[%+v] in purge.", c.businessName, c.appName, c.path, fContent.Name, fileCid)
				isNeedPurge = true
			}
		}

		if !isNeedPurge {
			fl.Unlock()
			continue
		}

		// remove memory information cache.
		c.mcache.Remove(fContent.Name())

		// rename the invalid cache to tmp.
		if err := os.Rename(c.contentPath(fContent.Name()), c.expiredFile(fContent.Name())); err != nil {
			logger.Warn("ContentCache[%s %s %s]| content cache purge, remove invalid cid, %+v", c.businessName, c.appName, c.path, err)
			fl.Unlock()
			continue
		}

		logger.Warn("ContentCache[%s %s %s]| content cache purge, purge cid[%+v] success", c.businessName, c.appName, c.path, fContent.Name())
		fl.Unlock()
	}
	return nil
}

// Setup setups the content cache.
func (c *ContentCache) Setup() {
	// start purging content cache.
	ticker := time.NewTicker(c.purgeInterval)
	defer ticker.Stop()

	for {
		if c.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(c.businessName, c.appName, c.path))) {
			return
		}

		<-ticker.C
		c.purge()
	}
}
