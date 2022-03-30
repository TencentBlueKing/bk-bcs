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
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// CheckNSAccess 检查该 API 能否访问指定命名空间
func CheckNSAccess(ctx context.Context, projectID, clusterID, namespace string) error {
	clusterInfo, err := cluster.GetClusterInfo(clusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}

	if !cli.IsProjNSinSharedCluster(ctx, projectID, clusterID, namespace) {
		return errorx.New(errcode.NoPerm, "在该共享集群中，该命名空间不属于指定项目")
	}
	return nil
}

// CheckSubscribable 检查指定参数能否进行订阅
func CheckSubscribable(ctx context.Context, req *clusterRes.SubscribeReq) error {
	clusterInfo, err := cluster.GetClusterInfo(req.ClusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}

	if !slice.StringInSlice(req.Kind, cluster.SharedClusterAccessibleResKinds) {
		return errorx.New(errcode.NoPerm, "在共享集群中，只有指定的数类资源可以执行订阅功能")
	}

	// 命名空间可以直接查询，但是不属于项目的需要被过滤掉
	if req.Kind == res.NS {
		return nil
	}

	if !cli.IsProjNSinSharedCluster(ctx, req.ProjectID, req.ClusterID, req.Namespace) {
		return errorx.New(errcode.NoPerm, "在该共享集群中，该命名空间不属于指定项目")
	}
	return nil
}

// CheckCObjAccess 检查指定 CObj 是否可查看/操作
func CheckCObjAccess(ctx context.Context, projectID, clusterID, crdName, namespace string) error {
	clusterInfo, err := cluster.GetClusterInfo(clusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}

	if !slice.StringInSlice(crdName, conf.G.SharedCluster.EnabledCRDs) {
		return errorx.New(errcode.NoPerm, "共享集群暂时只支持查询部分自定义资源")
	}

	if !cli.IsProjNSinSharedCluster(ctx, projectID, clusterID, namespace) {
		return errorx.New(errcode.NoPerm, "在该共享集群中，该命名空间不属于指定项目")
	}
	return nil
}
