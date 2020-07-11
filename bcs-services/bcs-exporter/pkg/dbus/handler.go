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

package dbus

import (
	"plugin"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg/output"
)

const (
	InitFuncName = "Init"
)

func RegisterOutputPlugins(c *config.Config) error {
	for _, pluginPath := range c.OutputPlugins {
		plugRaw, err := plugin.Open(pluginPath)
		if err != nil {
			blog.Errorf("try to open plugin %s failed: %v", pluginPath, err)
			continue
		}

		initFuncRaw, err := plugRaw.Lookup(InitFuncName)
		if err != nil {
			blog.Errorf("try to lookup init function in plugin %s failed: %v", pluginPath, err)
			continue
		}

		initFunc, ok := initFuncRaw.(func(*config.Config) (output.PluginIf, error))
		if !ok {
			blog.Errorf("try to assert the init function in plugin %s failed: %v", pluginPath, err)
			continue
		}

		thePlugin, err := initFunc(c)
		if err != nil {
			blog.Errorf("try to call the init function in plugin %s failed: %v", pluginPath, err)
			continue
		}

		if _, ok := msgBus.factories[thePlugin.Key()]; ok {
			blog.Errorf("duplicated plugin %s", pluginPath)
			continue
		}

		blog.Info("success to init plugin: %s", thePlugin.Name())
		msgBus.factories[thePlugin.Key()] = thePlugin
	}
	return nil
}
