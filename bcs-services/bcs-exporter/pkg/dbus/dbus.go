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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg/output"
)

var msgBus = DataBus{factories: make(map[output.PluginKey]output.PluginIf)}

// DataBus 内部数据流转的接口
type DataBus struct {
	factories map[output.PluginKey]output.PluginIf
}

// New 创建DataBus实例的方法
func New(cfg *config.Config) (MsgBusIf, error) {

	for _, item := range msgBus.factories {
		if err := item.SetCfg(cfg); nil != err {
			blog.Errorf("failed to set config, error info is %s ", err.Error())
			return nil, err
		}
	}

	return &msgBus, nil
}

// Write MsgBusIf 接口实现
func (cli *DataBus) Write(typeid, dataID int, data []byte) (int, error) {

	target, ok := cli.factories[output.PluginKey(typeid)]
	if !ok {
		return len(data), fmt.Errorf("not fond the plugin-type[%d]", typeid)
	}

	err := target.AddData(pkg.MapStr{
		"extID": dataID,
		"data":  data,
	})

	return len(data), err
}
