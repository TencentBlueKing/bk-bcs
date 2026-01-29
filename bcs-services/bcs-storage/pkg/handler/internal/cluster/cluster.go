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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/cluster"
)

var (
	clusterFeatTags = []string{constants.ClusterIDUpperTag}
)

type general struct {
	ctx              context.Context
	clusterID        string
	resourceFeatList []string
	data             map[string]interface{}
}

func (g *general) getCluCondition() *operator.Condition {
	features := operator.M{constants.ClusterIDUpperTag: g.clusterID}
	return operator.NewLeafCondition(operator.Eq, features)
}

func (g *general) getFeatures() operator.M {
	features := make(operator.M)
	features[constants.ClusterIDUpperTag] = g.clusterID
	return features
}

func (g *general) putResources() error {
	return cluster.PutData(g.ctx, g.data, g.getFeatures(), g.resourceFeatList)
}

func (g *general) deleteResources() ([]operator.M, error) {
	condition := g.getCluCondition()
	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}
	rmOption := &lib.StoreRemoveOption{
		Cond: condition,
		// when resource to be deleted not found, do not return error
		IgnoreNotFound: true,
	}
	return cluster.DeleteBatchData(g.ctx, getOption, rmOption)
}

// HandlerCreateClusterInfoReq  CreateClusterInfoReq业务方法
func HandlerCreateClusterInfoReq(ctx context.Context, req *storage.PutClusterInfoRequest) (operator.M, error) {

	data := map[string]interface{}{
		"clusterID": req.ClusterID,
		"data":      req.Data.AsMap(),
	}
	r := &general{
		ctx:              ctx,
		clusterID:        req.ClusterID,
		resourceFeatList: clusterFeatTags,
		data:             data,
	}

	return data, r.putResources()
}

// HandlerDeleteClusterInfoReq  DeleteClusterInfoReq业务方法
func HandlerDeleteClusterInfoReq(ctx context.Context, req *storage.DeleteClusterInfoRequest) ([]operator.M, error) {

	r := &general{
		ctx:              ctx,
		clusterID:        req.ClusterID,
		resourceFeatList: clusterFeatTags,
	}

	return r.deleteResources()
}
