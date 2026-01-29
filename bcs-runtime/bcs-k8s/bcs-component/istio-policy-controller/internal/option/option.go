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
// package xxx
package option

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"istio.io/api/networking/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

// ConnectionPoolSettings_HTTPSettings_H2UpgradePolicy Policy for upgrading http1.1 connections to http2.
type ConnectionPoolSettings_HTTPSettings_H2UpgradePolicy int32

// Configuration 是整个 YAML 配置的根结构
type Configuration struct {
	Global   Global
	Services []Service
}

// Global 对应 global 字段
type Global struct {
	TrafficPolicy *v1alpha3.TrafficPolicy
	Setting       Setting
}

// Service 单个服务的配置
type Service struct {
	Name          string
	Namespace     string
	TrafficPolicy *v1alpha3.TrafficPolicy
	Setting       Setting
}

// Setting 全局设置
type Setting struct {
	MergeMode                   string // e.g., "merge"
	DeletePolicyOnServiceDelete bool
	AutoGenerateVS              bool
	UpdateUnmanagedResources    bool
}

// ControllerOption controller option
type ControllerOption struct {
	// Address address for server
	Address string

	// MetricPort port for metric server
	MetricPort int

	conf.LogConfig

	// ConfigPath config file path
	ConfigPath string
	// Cfg config object
	Cfg *Configuration
}

// InitCfg init config
func (o *ControllerOption) InitCfg() error {
	if o.ConfigPath == "" {
		return fmt.Errorf("config file name is empty")
	}

	content, err := os.ReadFile(o.ConfigPath)
	if err != nil {
		return err
	}

	ctrl.Log.WithName("config").Info(fmt.Sprintf("config content: %s", string(content)))

	o.Cfg = &Configuration{}
	if err := yaml.Unmarshal(content, o.Cfg); err != nil {
		return err
	}

	return nil
}
