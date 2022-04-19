/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

import (
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NewMockClusterConfig 生成测试用 ClusterConf 对象（默认是本地集群）
func NewMockClusterConfig(clusterID string) *ClusterConf {
	kubeConfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	conf, _ := clientcmd.BuildConfigFromFlags("", kubeConfig)
	return &ClusterConf{conf, clusterID}
}

// ConvertInt2Int64 示例模板中加载出的数据，数字类型为 int，而 dynamicClient 返回结果中为 int64，
// 从而导致模板加载的数据无法被解析。这里使用递归的方式，强制转换，仅用于表单化解析/渲染单元测试！
func ConvertInt2Int64(raw map[string]interface{}) {
	for key, val := range raw {
		switch v := val.(type) {
		case map[string]interface{}:
			ConvertInt2Int64(v)
		case []interface{}:
			newList := []interface{}{}
			for _, item := range v {
				if it, ok := item.(map[string]interface{}); ok {
					ConvertInt2Int64(it)
					newList = append(newList, it)
				} else if it, ok := item.(int); ok {
					newList = append(newList, int64(it))
				} else {
					newList = append(newList, it)
				}
			}
			raw[key] = newList
		case int:
			raw[key] = int64(v)
		default:
			continue
		}
	}
}
