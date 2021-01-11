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
	"os"
	"sync"

	"github.com/go-ini/ini"

	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// ReleaseMetadata is release metadata struct.
type ReleaseMetadata struct {
	// config id.
	CfgID string

	// config name.
	CfgName string

	// config fpath.
	CfgFpath string

	// User config file user.
	User string

	// UserGroup config file user group.
	UserGroup string

	// FilePrivilege config file privilege.
	FilePrivilege string

	// FileFormat config file Format.
	FileFormat string

	// FileMode config file mode.
	FileMode int32

	// serial num of release.
	Serialno uint64

	// release id.
	ReleaseID string

	// content id.
	ContentID string

	// content size.
	ContentSize uint64

	// release effect time.
	EffectTime string

	// release name.
	ReleaseName string

	// multi release id.
	MultiReleaseID string

	isRollback bool
}

// Config is struct for one config which marked the current effected release.
type Config struct {
	// config id.
	cfgID string

	// current release metadata.
	current *ReleaseMetadata
}

// NewConfig creates new Config.
func NewConfig(cfgID string) *Config {
	return &Config{cfgID: cfgID}
}

// Effect saves new effected release metadata.
func (cfg *Config) Effect(metadata *ReleaseMetadata) error {
	if metadata == nil {
		return errors.New("invalid metadata: nil")
	}
	cfg.current = metadata
	return nil
}

// Current returns newest release which is effected at the moment.
func (cfg *Config) Current() *ReleaseMetadata {
	return cfg.current
}

const (
	// release details file name.
	releaseDetailsFileName = "release.details"

	// release config lock file name.
	releaseConfigLockFileName = "release.lock"

	// cfgid in details.
	detailsCfgID = "cfgID"

	// cfg name in details.
	detailsCfgName = "cfgName"

	// cfg fpath in details.
	detailsCfgFpath = "cfgPath"

	// serialno in details.
	detailsSerialno = "serialno"

	// releaseid in details.
	detailsReleaseID = "releaseID"

	// content id in details.
	detailsContentID = "contentID"

	// content size in details.
	detailsContentSize = "contentSize"

	// release effected time.
	detailsEffectTime = "effectTime"

	// release name.
	detailsReleaseName = "releaseName"

	// multi releaseid.
	detailsMultiReleaseID = "multiReleaseID"

	// release event type.
	detailsIsRollback = "rollback"
)

func effectCacheConfigPath(effectFileCachePath, cfgID string) string {
	return fmt.Sprintf("%s/%s", effectFileCachePath, cfgID)
}

func effectCacheDetailsFile(effectFileCachePath, cfgID string) string {
	return fmt.Sprintf("%s/%s/%s", effectFileCachePath, cfgID, releaseDetailsFileName)
}

func effectCacheLockFile(effectFileCachePath, cfgID string) string {
	return fmt.Sprintf("%s/.%s.%s", effectFileCachePath, cfgID, releaseConfigLockFileName)
}

// EffectCache is config release effect cache.
type EffectCache struct {
	bizID string
	appID string
	path  string

	// config metadatas, cfgid -> Config.
	configs map[string]*Config

	// mu make ops on configs safely.
	mu sync.RWMutex

	// path of effect file cache.
	effectFileCachePath string
}

// NewEffectCache creates new EffectCache.
func NewEffectCache(bizID, appID, path, effectFileCachePath string) *EffectCache {
	return &EffectCache{
		bizID:               bizID,
		appID:               appID,
		path:                path,
		effectFileCachePath: effectFileCachePath,
		configs:             make(map[string]*Config),
	}
}

// writeDetails writes relese details to file cache.
func (c *EffectCache) writeDetails(metadata *ReleaseMetadata) error {
	if err := os.MkdirAll(effectCacheConfigPath(c.effectFileCachePath, metadata.CfgID), os.ModePerm); err != nil {
		return err
	}

	fl, err := LockFile(effectCacheLockFile(c.effectFileCachePath, metadata.CfgID), true)
	if err != nil {
		return err
	}
	defer UnlockFile(fl)

	detailsFile := effectCacheDetailsFile(c.effectFileCachePath, metadata.CfgID)

	details, err := ini.LooseLoad(detailsFile)
	if err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsCfgID, metadata.CfgID); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsCfgName, metadata.CfgName); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsCfgFpath, metadata.CfgFpath); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsReleaseID, metadata.ReleaseID); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsContentID, metadata.ContentID); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsContentSize, common.ToStr(int(metadata.ContentSize))); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsSerialno, fmt.Sprintf("%d", metadata.Serialno)); err != nil {
		return err
	}
	if _, err := details.Section("").NewKey(detailsEffectTime, metadata.EffectTime); err != nil {
		return err
	}
	details.Section("").NewKey(detailsReleaseName, metadata.ReleaseName)
	details.Section("").NewKey(detailsMultiReleaseID, metadata.MultiReleaseID)
	details.Section("").NewKey(detailsIsRollback, fmt.Sprintf("%v", metadata.isRollback))

	if err := details.SaveTo(detailsFile); err != nil {
		return err
	}
	return nil
}

// readDetails reads relese details from file cache.
func (c *EffectCache) readDetails(cfgID string) (*ReleaseMetadata, error) {
	if err := os.MkdirAll(effectCacheConfigPath(c.effectFileCachePath, cfgID), os.ModePerm); err != nil {
		return nil, err
	}

	fl, err := LockFile(effectCacheLockFile(c.effectFileCachePath, cfgID), true)
	if err != nil {
		return nil, err
	}
	defer UnlockFile(fl)

	details, err := ini.LooseLoad(effectCacheDetailsFile(c.effectFileCachePath, cfgID))
	if err != nil {
		return nil, err
	}

	dCfgID := details.Section("").Key(detailsCfgID).String()
	if dCfgID == "" {
		return nil, errors.New("invalid detail:cfgID")
	}

	cfgName := details.Section("").Key(detailsCfgName).String()
	if cfgName == "" {
		return nil, errors.New("invalid detail:cfgName")
	}
	cfgFpath := details.Section("").Key(detailsCfgFpath).String()

	releaseID := details.Section("").Key(detailsReleaseID).String()
	if releaseID == "" {
		return nil, errors.New("invalid detail:releaseID")
	}

	contentID := details.Section("").Key(detailsContentID).String()
	if contentID == "" {
		return nil, errors.New("invalid detail:contentID")
	}
	contentSize, err := details.Section("").Key(detailsContentSize).Uint64()
	if err != nil {
		return nil, err
	}

	effectTime := details.Section("").Key(detailsEffectTime).String()
	if effectTime == "" {
		return nil, errors.New("invalid detail:effectTime")
	}

	serialno, err := details.Section("").Key(detailsSerialno).Uint64()
	if err != nil {
		return nil, err
	}
	releaseName := details.Section("").Key(detailsReleaseName).String()
	multiReleaseID := details.Section("").Key(detailsMultiReleaseID).String()
	isRollback, _ := details.Section("").Key(detailsIsRollback).Bool()

	metadata := &ReleaseMetadata{
		CfgID:          dCfgID,
		CfgName:        cfgName,
		CfgFpath:       cfgFpath,
		Serialno:       serialno,
		ReleaseID:      releaseID,
		ContentID:      contentID,
		ContentSize:    contentSize,
		EffectTime:     effectTime,
		ReleaseName:    releaseName,
		MultiReleaseID: multiReleaseID,
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

	if _, ok := c.configs[metadata.CfgID]; !ok {
		c.configs[metadata.CfgID] = NewConfig(metadata.CfgID)
	}

	config := c.configs[metadata.CfgID]
	return config.Effect(metadata)
}

// LocalRelease returns local effected release information of target config.
func (c *EffectCache) LocalRelease(cfgID string) (*ReleaseMetadata, error) {
	c.mu.RLock()
	config, ok := c.configs[cfgID]
	if ok && config != nil {
		c.mu.RUnlock()
		return config.Current(), nil
	}
	c.mu.RUnlock()

	md, err := c.readDetails(cfgID)
	if err != nil {
		logger.V(4).Infof("EffectCache[%s %s %s]| suppose no effected release of cfgID[%+v], %+v",
			c.bizID, c.appID, c.path, cfgID, err)
		return nil, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.configs[md.CfgID]; !ok {
		c.configs[md.CfgID] = NewConfig(md.CfgID)
	}

	config = c.configs[md.CfgID]
	config.Effect(md)

	return md, nil
}

// NeedEffected checks whether need to effect the release.
func (c *EffectCache) NeedEffected(cfgID string, serialno uint64) (bool, error) {
	md, err := c.LocalRelease(cfgID)
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
func (c *EffectCache) Debug(cfgID string) string {
	md, err := c.LocalRelease(cfgID)
	if err != nil {
		return fmt.Sprintf("can't get cfgID[%s] debug information, %s", cfgID, err.Error())
	}

	if md == nil {
		return fmt.Sprintf("cfgID[%+v] no effected release", cfgID)
	}

	return fmt.Sprintf("cfgID[%+v] localRelease[%+v] localSerialno[%+v] effecttime[%+v]",
		cfgID, md.ReleaseID, md.Serialno, md.EffectTime)
}
