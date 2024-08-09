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

// Package bcs xxx
package bcs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

const (
	// TokenExpired xxx
	TokenExpired = time.Hour * 24
)

// BCSTokenUserType xxx
// bcs-usermamager 用户类型
type BCSTokenUserType int // nolint

const (
	// AdminUser xxx
	AdminUser BCSTokenUserType = 1
	// SaaSUser xxx
	SaaSUser BCSTokenUserType = 2
	// GeneralUser xxx
	GeneralUser BCSTokenUserType = 3
)

// Cluster xxx
type Cluster struct {
	ProjectId   string `json:"projectID"`
	ClusterId   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
}

// String xxx
func (c *Cluster) String() string {
	return fmt.Sprintf("cluster<%s, %s>", c.ClusterName, c.ClusterId)
}

// ListClusters 获取项目集群列表
func ListClusters(ctx context.Context, projectId string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", config.G.BCS.InnerHost)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.G.BCS.Token).
		SetQueryParam("projectID", projectId).
		Get(url)

	if err != nil {
		return nil, err
	}

	var result []*Cluster
	if err := components.UnmarshalBKResult(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCluster 获取单个集群信息
func GetCluster(ctx context.Context, projectId, clusterId string) (*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster/%s", config.G.BCS.InnerHost, clusterId)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.G.BCS.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	var cluster *Cluster
	if err := components.UnmarshalBKResult(resp, &cluster); err != nil {
		return nil, err
	}

	// 共享集群的项目Id和当前项目会不一致
	if !cluster.IsShared && cluster.ProjectId != projectId {
		return nil, errors.New("project or cluster not valid")
	}

	return cluster, nil
}

// Token xxx
type Token struct {
	Token     string `json:"token"`
	ExpiredAt string `json:"expired_at"`
}

// CreateTempToken 创建临时 token
func CreateTempToken(ctx context.Context, username, clusterId string) (*Token, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/usermanager/v1/tokens/temp", config.G.BCS.InnerHost)

	// 管理员账号不做鉴权
	var userType BCSTokenUserType
	if config.G.IsManager(username, clusterId) {
		userType = AdminUser
	} else {
		userType = GeneralUser
	}

	data := map[string]interface{}{
		"usertype":   userType,
		"username":   username,
		"expiration": TokenExpired.Seconds(),
	}
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.G.BCS.Token).
		SetBody(data).
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
