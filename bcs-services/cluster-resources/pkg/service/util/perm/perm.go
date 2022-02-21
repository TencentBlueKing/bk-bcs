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

// Package perm handler 权限检查
package perm

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// AccessClusterPermCheck 检查该 API 能否访问指定集群
func AccessClusterPermCheck(projectID, clusterID string) error {
	clusterInfo, err := cluster.GetClusterInfo(clusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}
	return fmt.Errorf("不支持使用该 API 访问共享集群资源")
}

// AccessNSPermCheck 检查该 API 能够访问指定命名空间
func AccessNSPermCheck(projectID, clusterID, namespace string) error {
	clusterInfo, err := cluster.GetClusterInfo(clusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}

	if !cli.IsProjNSinSharedCluster(projectID, clusterID, namespace) {
		return fmt.Errorf("在该共享集群中，该命名空间不属于指定项目")
	}
	return nil
}

// SubscribableCheck 检查指定参数能否进行订阅
func SubscribableCheck(req *clusterRes.SubscribeReq) error {
	clusterInfo, err := cluster.GetClusterInfo(req.ClusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}

	if !util.StringInSlice(req.Kind, cluster.SharedClusterAccessibleResKinds) {
		return fmt.Errorf("在共享集群中，只有指定的数类资源可以执行订阅功能")
	}

	if req.Kind == res.NS {
		return nil
	}

	if !cli.IsProjNSinSharedCluster(req.ProjectID, req.ClusterID, req.Namespace) {
		return fmt.Errorf("在该共享集群中，该命名空间不属于指定项目")
	}
	return nil
}
