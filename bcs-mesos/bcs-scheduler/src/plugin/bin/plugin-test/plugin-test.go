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

package plugin_test

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/plugin"
)

// plugin must implement
// func GetHostAttributes([]string) (map[string]*types.HostAttributes,error)
// func input: ip list, example: []string{"127.0.0.10","127.0.0.11","127.0.0.12"}
// func ouput: map key = ip, example: map["127.0.0.10"] = &types.HostAttributes{}
// implement func Init(para *types.InitPluginParameter) error
// func input: *types.InitPluginParameter
// func output: error

//for example

var initPara *plugin.InitPluginParameter

func Init(para *plugin.InitPluginParameter) error {
	initPara = para
	return nil
}

func Uninit() {
	//TODO
}

func GetHostAttributes(para *plugin.HostPluginParameter) (map[string]*plugin.HostAttributes, error) {
	atrrs := make(map[string]*plugin.HostAttributes)

	for _, ip := range para.Ips {
		hostAttr := &plugin.HostAttributes{
			Ip:         ip,
			Attributes: make([]*plugin.Attribute, 0),
		}

		atrri := &plugin.Attribute{
			Name:   "ip-resources",
			Type:   plugin.ValueScalar,
			Scalar: plugin.Value_Scalar{Value: 10},
		}
		hostAttr.Attributes = append(hostAttr.Attributes, atrri)

		atrrs[ip] = hostAttr
	}

	return atrrs, nil
}
