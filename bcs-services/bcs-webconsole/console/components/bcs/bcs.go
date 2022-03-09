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
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"

	"github.com/pkg/errors"
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

type Result struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []*Cluster `json:"data"`
}

func ListClusters(ctx context.Context, projectId string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", config.G.BCS.Host)
	resp, err := components.GetClient().R().
		SetBearerAuthToken(config.G.BCS.Token).
		SetQueryParam("projectID", projectId).
		Get(url)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode)
	}

	var result Result
	if err := resp.Unmarshal(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(fmt.Sprintf("query clustermanager error, %s", result.Message))
	}

	var clusters []*Cluster
	for _, cluster := range result.Data {
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
func CreateTempToken(ctx context.Context, username string) (*Token, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/usermanager/v1/tokens/temp", config.G.BCS.Host)
	data := map[string]interface{}{
		"username":   username,
		"expiration": TokenExpired.Seconds(),
	}
	resp, err := components.GetClient().R().SetBearerAuthToken(config.G.BCS.Token).SetBodyJsonMarshal(data).Post(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode)
	}

	type TokenResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *Token `json:"data"`
	}

	var result TokenResult
	if err := resp.Unmarshal(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(fmt.Sprintf("create token error, %s", result.Message))
	}
	return result.Data, nil
}
