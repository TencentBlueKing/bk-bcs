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

package hostconfig

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/hostconfig"
)

var (
	hostFeatTags      = []string{constants.IpTag}
	hostQueryFeatTags = []string{constants.ClusterIDTag}
)

type general struct {
	ips      []string
	featList []string
	raw      operator.M
	ctx      context.Context
	host     *storage.BcsStorageHost
	isPut    bool
}

func (g *general) preRaw() {
	g.raw = util.StructToMap(g.host)
}

func (g *general) getFeat() *operator.Condition {
	features := make(operator.M, len(g.featList))

	for _, key := range g.featList {
		features[key] = g.raw[key]
	}

	return operator.NewLeafCondition(operator.Eq, features)
}

func (g *general) getFeatM() operator.M {
	features := make(operator.M, len(g.featList))

	for _, key := range g.featList {
		features[key] = g.raw[key]
	}

	return features
}

func (g *general) getHost() ([]operator.M, error) {
	g.preRaw()
	return hostconfig.QueryHost(g.ctx, g.getFeat())
}

func (g *general) getReqData(features operator.M) operator.M {
	tmp := util.StructToMap(g.host.Data)

	data := lib.CopyMap(features)
	data[constants.ClusterIDTag] = g.host.ClusterId
	data[constants.DataTag] = tmp
	return data
}

func (g *general) putHost() error {
	g.preRaw()

	features := g.getFeatM()
	condition := operator.NewLeafCondition(operator.Eq, features)
	data := g.getReqData(features)

	return hostconfig.PutHostToDB(g.ctx, data, condition)
}

func (g *general) removeHost() error {
	g.preRaw()
	return hostconfig.RemoveHost(g.ctx, g.getFeat())
}

func (g *general) listHost() ([]operator.M, error) {
	g.preRaw()
	return hostconfig.QueryHost(g.ctx, g.getFeat())
}

func (g *general) getRelationFeat() *types.BcsStorageClusterRelationIf {
	return &types.BcsStorageClusterRelationIf{
		Ips: g.ips,
	}
}

func (g *general) getRelationData() operator.M {
	return operator.M{
		constants.ClusterIDTag: g.host.ClusterId,
	}
}

func (g *general) doRelation() error {
	g.preRaw()
	relation := g.getRelationFeat()
	data := g.getRelationData()
	condition := operator.NewLeafCondition(operator.In, operator.M{constants.IpTag: relation.Ips})
	opt := &lib.StoreGetOption{
		Cond:   condition,
		Fields: []string{constants.IpTag},
	}
	return hostconfig.DoRelation(g.ctx, opt, data, g.isPut, relation)
}

// HandlerGetHost GetHost 业务方法
func HandlerGetHost(ctx context.Context, req *storage.GetHostRequest) ([]operator.M, error) {
	g := &general{
		ctx:      ctx,
		featList: hostFeatTags,
		host: &storage.BcsStorageHost{
			Ip: req.Ip,
		},
	}

	return g.getHost()
}

// HandlerPutHost PutHost 业务方法
func HandlerPutHost(ctx context.Context, req *storage.PutHostRequest) error {
	host := &storage.BcsStorageHost{
		Ip:        req.Ip,
		ClusterId: req.ClusterId,
	}
	g := &general{
		ctx:      ctx,
		featList: hostFeatTags,
		host:     host,
	}

	return g.putHost()
}

// HandlerDeleteHost DeleteHost 业务方法
func HandlerDeleteHost(ctx context.Context, req *storage.DeleteHostRequest) error {
	g := &general{
		ctx:      ctx,
		featList: hostFeatTags,
		host: &storage.BcsStorageHost{
			Ip: req.Ip,
		},
	}

	return g.putHost()
}

// HandlerListHost ListHost 业务方法
func HandlerListHost(ctx context.Context, req *storage.ListHostRequest) ([]operator.M, error) {
	g := &general{
		ctx:      ctx,
		featList: hostQueryFeatTags,
		host: &storage.BcsStorageHost{
			ClusterId: req.ClusterId,
		},
	}

	return g.listHost()
}

// HandlerPutClusterRelation PutClusterRelation 业务方法
func HandlerPutClusterRelation(ctx context.Context, req *storage.PutClusterRelationRequest) error {
	g := &general{
		ctx:      ctx,
		featList: hostFeatTags,
		host: &storage.BcsStorageHost{
			ClusterId: req.ClusterId,
		},
		ips:   req.Ips,
		isPut: true,
	}

	return g.doRelation()
}

// HandlerPostClusterRelation PostClusterRelation 业务方法
func HandlerPostClusterRelation(ctx context.Context, req *storage.PostClusterRelationRequest) error {
	g := &general{
		ctx:      ctx,
		featList: hostFeatTags,
		host: &storage.BcsStorageHost{
			ClusterId: req.ClusterId,
		},
		ips:   req.Ips,
		isPut: false,
	}

	return g.doRelation()
}
