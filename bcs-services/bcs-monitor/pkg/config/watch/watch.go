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
 *
 */

package watch

import (
	"errors"
	"sync"
	"time"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// multiConf 凭证
type multiCredConf struct {
	mtx           sync.RWMutex
	confMap       map[string]*viper.Viper
	lastWatchTime map[string]time.Time
}

// LoadWithLimit fsnotify 有多个event， 这里限制1秒内置 reload 一次
func (m *multiCredConf) LoadWithLimit(name string, v *viper.Viper) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if time.Since(m.lastWatchTime[name]) < time.Second {
		return nil
	}

	if err := config.G.ReadCredViper(name, v); err != nil {
		logger.Errorf("reload credential error, %s", err)
		return err
	}

	logger.Infof("reload credential conf from %s, len=%d", name, len(config.G.Credentials[name]))
	m.lastWatchTime[name] = time.Now()
	return nil
}

// makeConf 配置文件
func makeConf(filePath string, loadFunc func(name string, v *viper.Viper) error) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// 自动 watch 配置
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		loadFunc(filePath, v)
	})

	return v, nil
}

// MultiCredWatch 新增
func MultiCredWatch(filePaths []string) error {
	confs := &multiCredConf{
		confMap:       make(map[string]*viper.Viper),
		lastWatchTime: make(map[string]time.Time),
	}

	for _, filePath := range filePaths {
		_, ok := confs.confMap[filePath]
		if ok {
			return errors.New("credential config is duplicated")
		}

		conf, err := makeConf(filePath, confs.LoadWithLimit)
		if err != nil {
			return err
		}

		if err := config.G.ReadCredViper(filePath, conf); err != nil {
			return err
		}
		logger.Infof("load credential conf from %s, len=%d", filePath, len(config.G.Credentials[filePath]))

		confs.confMap[filePath] = conf
	}

	return nil
}
