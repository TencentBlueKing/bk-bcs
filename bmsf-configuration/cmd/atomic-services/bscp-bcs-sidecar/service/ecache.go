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
	"sync"
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

// EffectCache is config release effect cache.
type EffectCache struct {
	bizID string
	appID string
	path  string

	// config metadatas, cfgid -> Config.
	configs map[string]*Config
	mu      sync.RWMutex
}

// NewEffectCache creates new EffectCache.
func NewEffectCache(bizID, appID, path string) *EffectCache {
	return &EffectCache{
		bizID:   bizID,
		appID:   appID,
		path:    path,
		configs: make(map[string]*Config),
	}
}

// Reset resets the local effect cache.
func (c *EffectCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.configs = make(map[string]*Config)
}

// Effect adds effected release.
func (c *EffectCache) Effect(metadata *ReleaseMetadata) error {
	if metadata == nil {
		return errors.New("invalid metadata: nil")
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
	defer c.mu.RUnlock()

	config, isExist := c.configs[cfgID]
	if isExist && config != nil && config.Current() != nil {
		return config.Current(), nil
	}
	return nil, errors.New("no local effected release")
}

// NeedEffected checks whether need to effect the release.
func (c *EffectCache) NeedEffected(cfgID string, serialno uint64) bool {
	md, _ := c.LocalRelease(cfgID)
	if md == nil {
		return true
	}
	if serialno <= md.Serialno {
		return false
	}
	return true
}

// Debug returns debug information.
func (c *EffectCache) Debug(cfgID string) string {
	md, _ := c.LocalRelease(cfgID)
	if md == nil {
		return fmt.Sprintf("cfgID[%+v] no effected release", cfgID)
	}
	return fmt.Sprintf("cfgID[%+v] localRelease[%+v] localSerialno[%+v] effecttime[%+v]",
		cfgID, md.ReleaseID, md.Serialno, md.EffectTime)
}
