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

package multicfg

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// MultiCredConf 凭证
type MultiCredConf struct {
	confMap map[string]*viper.Viper
}

// makeMicroConf 配置文件
func makeMicroConf(filePath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// 自动 watch 配置
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := config.G.ReadCredViper(filePath, v); err != nil {
			logger.Errorf("reload credential error, %s", err)
		}
		logger.Infof("reload credential conf from %s, len=%d", filePath, len(config.G.Credentials[filePath]))
	})

	return v, nil
}

// MultiCredConf 新增
func NewMultiCredConf(filePaths []string) (*MultiCredConf, error) {
	multiCredConf := &MultiCredConf{
		confMap: make(map[string]*viper.Viper),
	}

	for _, filePath := range filePaths {
		_, ok := multiCredConf.confMap[filePath]
		if ok {
			return nil, errors.New("credential config is duplicated")
		}

		conf, err := makeMicroConf(filePath)
		if err != nil {
			return nil, err
		}

		if err := config.G.ReadCredViper(filePath, conf); err != nil {
			return nil, err
		}
		logger.Infof("load credential conf from %s, len=%d", filePath, len(config.G.Credentials[filePath]))

		multiCredConf.confMap[filePath] = conf

	}

	return multiCredConf, nil
}
