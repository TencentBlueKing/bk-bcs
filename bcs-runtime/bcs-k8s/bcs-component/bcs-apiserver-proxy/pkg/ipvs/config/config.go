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

// Package config xxx
package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

const (
	// IpvsConfigFileName xxx
	IpvsConfigFileName = "ipvsConfig.yaml"
)

// IpvsConfig xxx
type IpvsConfig struct {
	Scheduler     string   `json:"scheduler"`
	VirtualServer string   `json:"vs"`
	RealServer    []string `json:"rs"`
}

// WriteIpvsConfig xxx
func WriteIpvsConfig(dir string, config IpvsConfig) error {
	_, exist := os.Stat(dir)
	if os.IsNotExist(exist) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Println("create ipvs persist dir failed")
			return err
		}
	}
	viper.SetConfigFile(path.Join(dir, IpvsConfigFileName))
	viper.Set("vs", config.VirtualServer)
	viper.Set("rs", config.RealServer)
	viper.Set("scheduler", config.Scheduler)
	err := viper.WriteConfigAs(path.Join(dir, IpvsConfigFileName))
	if err != nil {
		fmt.Println("persist ipvs config to file failed")
		return err
	}
	return nil
}

// ReadIpvsConfig xxx
func ReadIpvsConfig(dir string) (*IpvsConfig, error) {
	_, exist := os.Stat(dir)
	if os.IsNotExist(exist) {
		err := fmt.Errorf("ipvs config persist dir [%s] not exists", dir)
		return nil, err
	}
	viper.SetConfigFile(path.Join(dir, IpvsConfigFileName))
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("read persist config failed, %v", err)
		return nil, err
	}
	config := &IpvsConfig{
		VirtualServer: viper.GetString("vs"),
		RealServer:    viper.GetStringSlice("rs"),
		Scheduler:     viper.GetString("scheduler"),
	}
	return config, nil
}
