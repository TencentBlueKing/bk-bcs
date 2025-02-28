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

// Package bcsstorage xxx
package bcsstorage

import (
	"fmt"
	"time"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
)

var (
	// defaultTimeOut 默认请求超时时间为60秒
	defaultTimeOut = time.Second * 60
)

// GetMultiClusterResourceQuotaResponse 获取多集群资源配额的响应结构
type GetMultiClusterResourceQuotaResponse struct {
	Code    int         `json:"code"`    // 响应状态码
	Result  bool        `json:"result"`  // 请求结果标识
	Message string      `json:"message"` // 响应消息
	Data    []QuotaInfo `json:"data"`    // 资源配额信息列表
}

// QuotaInfo 定义了资源配额信息的结构体
type QuotaInfo struct {
	Data MultiClusterResourceQuota `json:"data"` // 多集群资源配额数据
}

// GetMultiClusterResourceQuota 根据集群ID和配额名称获取多集群资源配额信息
// 参数:
//   - clusterID: 集群ID
//   - name: 资源配额名称
//
// 返回:
//   - *MultiClusterResourceQuota: 多集群资源配额信息
//   - error: 错误信息
func GetMultiClusterResourceQuota(clusterID, name string) (*MultiClusterResourceQuota, error) {
	var (
		// 构造请求URL
		reqURL = fmt.Sprintf(
			"%s/bcsapi/v4/storage/k8s/dynamic/cluster_resources/clusters/%s/MultiClusterResourceQuota",
			config.GlobalConf.BcsGateway.Host, clusterID)

		// 响应数据结构
		respData = &GetMultiClusterResourceQuotaResponse{}

		// 查询参数
		params = map[string]string{
			"data.metadata.name": name,
		}
	)

	// 发起HTTP GET请求获取资源配额信息
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(reqURL).
		Query(params).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("Authorization", fmt.Sprintf(`Bearer %s`, config.GlobalConf.BcsGateway.Token)).
		EndStruct(&respData)
	if len(errs) > 0 {
		logging.Error("GetMultiClusterResourceQuota err: %v", errs[0])
		return nil, errs[0]
	}

	// 检查响应结果
	if !respData.Result {
		logging.Error("GetMultiClusterResourceQuota failed: %s", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}
	logging.Info("GetMultiClusterResourceQuota %s successfully", reqURL)

	// 返回查询到的资源配额信息
	if len(respData.Data) > 0 {
		return &respData.Data[0].Data, nil
	}

	return nil, fmt.Errorf("GetMultiClusterResourceQuota failed")
}
