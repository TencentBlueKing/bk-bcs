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

package cc

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/config"
)

// LoadSettings load service's configuration
func LoadSettings(sys *SysOption) error {
	if len(sys.ConfigFiles) == 0 {
		return errors.New("service's configuration file path is not configured")
	}

	if err := initGlobalConf(sys.ConfigFiles); err != nil {
		return err
	}

	conf, err := mergeConfigFile(sys.ConfigFiles)
	if err != nil {
		return err
	}
	// configure file is configured, then load configuration from file.
	s, err := loadFromFile(conf)
	if err != nil {
		return err
	}

	if err = s.trySetFlagBindIP(sys.BindIP); err != nil {
		return err
	}

	if err = s.trySetFlagPort(sys.Port, sys.GRPCPort); err != nil {
		return err
	}

	// set the default value if user not configured.
	s.trySetDefault()

	if err := s.Validate(); err != nil {
		return err
	}

	initRuntime(s)

	return nil
}

// mergeConfigFile 合并多个配置文件
func mergeConfigFile(filenames []string) ([]byte, error) {
	masterConf := map[string]interface{}{}

	for _, filename := range filenames {
		file, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("load Setting from file: %s failed, err: %v", filename, err)
		}

		var tmpConf map[string]interface{}
		if err := yaml.Unmarshal(file, &tmpConf); err != nil {
			return nil, fmt.Errorf("unmarshal Setting yaml from file: %s failed, err: %v", filename, err)
		}

		// 后面的配置文件按Key直接覆盖前面的配置
		for k, v := range tmpConf {
			masterConf[k] = v
		}
	}

	// 合并后的配置
	return yaml.Marshal(masterConf)
}

// loadFromFile load service's configuration from local config file.
func loadFromFile(conf []byte) (Setting, error) {
	if len(conf) == 0 {
		return nil, errors.New("file name is not set")
	}

	var s Setting
	switch ServiceName() {
	case APIServerName:
		s = new(ApiServerSetting)
	case AuthServerName:
		s = new(AuthServerSetting)
	case CacheServiceName:
		s = new(CacheServiceSetting)
	case ConfigServerName:
		s = new(ConfigServerSetting)
	case DataServiceName:
		s = new(DataServiceSetting)
	case FeedServerName:
		s = new(FeedServerSetting)
	case VaultServerName:
		s = new(VaultServerSetting)
	default:
		return nil, fmt.Errorf("unknown %s service name", ServiceName())
	}

	if err := yaml.Unmarshal(conf, s); err != nil {
		return nil, fmt.Errorf("unmarshal Setting yaml from conf:\n%s failed, err: %v", conf, err)
	}

	return s, nil
}

// 初始化全局配置
func initGlobalConf(filenames []string) error {
	v := viper.New()
	for _, f := range filenames {
		fmt.Println(f)
		v.SetConfigType("yaml")
		v.SetConfigFile(f)
		if err := v.MergeInConfig(); err != nil {
			return err
		}
	}

	return config.G.ReadFromViper(v)
}
