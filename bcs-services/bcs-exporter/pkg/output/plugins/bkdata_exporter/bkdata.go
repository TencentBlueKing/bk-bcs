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

package main

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/protocol"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg/output"
)

// bkdataOutput 对接蓝鲸数据平台
type bkdataOutput struct {
	client AsyncProducer
	cfg    *config.Config
}

//Init plugin entrance
func Init(_ *config.Config) (output.PluginIf, error) { //nolint
	return &bkdataOutput{}, nil
}

// Key 返回唯一编码，决定消息的路由，全局唯一
func (cli *bkdataOutput) Key() output.PluginKey {
	return output.PluginKey(protocol.BKDataPlugin)
}

// SetCfg  设置配置
func (cli *bkdataOutput) SetCfg(cfg *config.Config) error {
	cli.cfg = cfg
	client, err := createClient(cfg)
	cli.client = client
	return err
}

// Name 返回Output 插件的名，否则会影响插件的注册
func (cli *bkdataOutput) Name() output.PluginName {
	return output.PluginName("bkdata_exporter")
}

// AddData 接收外部发送过来的数据
func (cli *bkdataOutput) AddData(mapStr pkg.MapStr) error {

	if dataID, ok := mapStr["extID"]; ok {
		if data, dataOk := mapStr["data"]; dataOk {
			var tmpid uint32
			switch dataID.(type) {
			case int:
				tmpid = uint32(dataID.(int))
			case int64:
				tmpid = uint32(dataID.(int64))
			case uint64:
				tmpid = uint32(dataID.(uint64))
			case float32:
				tmpid = uint32(dataID.(float32))
			case float64:
				tmpid = uint32(dataID.(float64))
			case uint32:
				tmpid = dataID.(uint32)
			}
			blog.V(3).Infof("data:%s", string(data.([]byte)))
			cli.client.Input(&ProducerMessage{DataID: tmpid, Value: data.([]byte)})
		}
	}
	return nil
}
