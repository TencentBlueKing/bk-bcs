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

// Package cluster xxx
package cluster

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// Cluster BCS 集群信息
type Cluster struct {
	ID       string `json:"clusterID"`
	Name     string `json:"clusterName"`
	ProjID   string `json:"projectID"`
	Status   string `json:"status"`
	IsShared bool   `json:"is_shared"`
	Type     string `json:"-"`
}

// String :
func (c *Cluster) String() string {
	return fmt.Sprintf("cluster<%s|%s, %s>", c.Name, c.Type, c.ID)
}

// GetClusterInfo xxx
func GetClusterInfo(ctx context.Context, clusterID string) (*Cluster, error) {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return fetchMockClusterInfo(clusterID)
	}
	return clusterMgrCli.fetchClusterInfoWithCache(ctx, clusterID)
}

// FromContext 通过 Context 获取集群信息
func FromContext(ctx context.Context) (*Cluster, error) {
	c := ctx.Value(ctxkey.ClusterKey)
	if c == nil {
		return nil, errorx.New(errcode.General, "cluster info not exists in context")
	}
	return c.(*Cluster), nil
}
