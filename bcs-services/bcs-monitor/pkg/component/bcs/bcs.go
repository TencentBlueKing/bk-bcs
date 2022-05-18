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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

const (
	TokenExpired = time.Hour * 24
)

// bcs-usermamager 用户类型
type BCSTokenUserType int

const (
	AdminUser   BCSTokenUserType = 1
	SaaSUser    BCSTokenUserType = 2
	GeneralUser BCSTokenUserType = 3
)

// Cluster 集群信息
type Cluster struct {
	ProjectId   string `json:"projectID"`
	ClusterId   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
}

// String
func (c *Cluster) String() string {
	return fmt.Sprintf("cluster<%s, %s>", c.ClusterName, c.ClusterId)
}

// ListClusters 获取项目集群列表
func ListClusters(ctx context.Context, bcsConf *config.BCSConf, projectId string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", bcsConf.Host)

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(bcsConf.Token).
		SetQueryParam("projectID", projectId).
		Get(url)

	if err != nil {
		return nil, err
	}

	var result []*Cluster
	if err := component.UnmarshalBKResult(resp, &result); err != nil {
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

// GetCluster 获取集群详情
func GetCluster(ctx context.Context, bcsConf *config.BCSConf, projectId, clusterId string) (*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster/%s", bcsConf.Host, clusterId)

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(bcsConf.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	var cluster *Cluster
	if err := component.UnmarshalBKResult(resp, &cluster); err != nil {
		return nil, err
	}

	// 共享集群的项目Id和当前项目会不一致
	if !cluster.IsShared && cluster.ProjectId != projectId {
		return nil, errors.New("project or cluster not valid")
	}

	return cluster, nil
}
