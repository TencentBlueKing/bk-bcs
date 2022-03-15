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

package bcs

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

const (
	TokenExpired = time.Hour * 24
)

type Cluster struct {
	ClusterId   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
}

// ListClusters 获取项目集群列表
func ListClusters(ctx context.Context, bcsConf *config.BCSConf, projectId string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", bcsConf.Host)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetBearerAuthToken(bcsConf.Token).
		SetQueryParam("projectID", projectId).
		Get(url)

	if err != nil {
		return nil, err
	}

	var result []*Cluster
	if err := components.UnmarshalBKResult(resp, result); err != nil {
		return nil, err
	}

	clusters := make([]*Cluster, 0, len(result))
	for _, cluster := range result {
		// 过滤掉共享集群
		if cluster.IsShared {
			continue
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

type Token struct {
	Token     string `json:"token"`
	ExpiredAt string `json:"expired_at"`
}

// CreateTempToken 创建临时 token
func CreateTempToken(ctx context.Context, bcsConf *config.BCSConf, username string) (*Token, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/usermanager/v1/tokens/temp", bcsConf.Host)

	data := map[string]interface{}{
		"username":   username,
		"expiration": TokenExpired.Seconds(),
	}
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetBearerAuthToken(bcsConf.Token).
		SetBodyJsonMarshal(data).
		Post(url)

	if err != nil {
		return nil, err
	}

	token := &Token{}
	if err := components.UnmarshalBKResult(resp, token); err != nil {
		return nil, err
	}

	return token, nil
}
