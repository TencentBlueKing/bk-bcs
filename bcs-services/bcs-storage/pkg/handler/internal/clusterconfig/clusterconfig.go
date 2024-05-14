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

// Package clusterconfig xxx
package clusterconfig

import (
	"context"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/clusterconfig"
)

type general struct {
	clusterID         string
	clusterIDNot      string
	service           string
	version           string
	ctx               context.Context
	bcsStorageCluster *storage.BcsStorageCluster
	clsFunc           func(g *general) ([]operator.M, error)
}

func (g *general) preService() {
	if g.service == "" {
		g.service = "test"
	}
}

func (g *general) getSvcCondition() *operator.Condition {
	return operator.NewLeafCondition(
		operator.Eq,
		operator.M{
			constants.ServiceTag: g.service,
		},
	)
}

func (g *general) getTemplateCondition() *operator.Condition {
	return operator.NewLeafCondition(
		operator.Eq,
		operator.M{
			constants.ServiceTag: g.service,
		},
	)
}

func (g *general) handler() (config *types.DeployConfig, err error) {
	var clsConfig []operator.M
	g.preService()
	service := g.service
	opt := &lib.StoreGetOption{
		Cond: g.getSvcCondition(),
	}
	if clsConfig, err = g.clsFunc(g); err != nil {
		return &types.DeployConfig{}, err
	}
	return clusterconfig.GenerateData(g.ctx, opt, clsConfig, service)
}

// getClusterIDCondition getClsCondition
func (g *general) getClusterIDCondition() *operator.Condition {
	features := operator.M{
		constants.ClusterIDTag: g.clusterID,
	}
	return operator.NewLeafCondition(operator.Eq, features)
}

func (g *general) getMultiClsCondition() *operator.Condition {
	var condList []*operator.Condition
	if g.clusterIDNot != "" {
		condList = append(condList, operator.NewLeafCondition(
			operator.Nin,
			operator.M{
				constants.ClusterIDTag: strings.Split(g.clusterIDNot, ","),
			},
		))
	} else if g.clusterID != "" {
		condList = append(condList, operator.NewLeafCondition(
			operator.In,
			operator.M{
				constants.ClusterIDTag: strings.Split(g.clusterID, ","),
			},
		))
	}
	return operator.NewBranchCondition(operator.And, condList...)
}

func (g *general) putStableVersion() (err error) {
	g.preService()
	if err = clusterconfig.SaveStableVersion(g.ctx, g.service, g.version); err != nil {
		return errors.Wrapf(err, "Failed to set stable version of %s", g.service)
	}
	return nil
}

func (g *general) getReqData() (operator.M, error) {
	g.bcsStorageCluster.NeedNat = true

	var renderConfig types.RenderConfig
	zk := wrapIP(g.bcsStorageCluster.ZkIP, "2181")
	dns := wrapIP(g.bcsStorageCluster.DnsIP, "53")
	clusterID := g.clusterID
	lin := strings.Split(clusterID, "-")

	renderConfig.MesosZk = strings.Join(zk, ",")
	renderConfig.MesosZkSpace = strings.Join(zk, " ")
	renderConfig.MesosZkSemicolon = strings.Join(zk, ";")
	renderConfig.MesosZkRaw = strings.Join(unwrapIP(zk), ",")
	renderConfig.MesosMaster = strings.Join(g.bcsStorageCluster.MasterIP, ",")
	renderConfig.MesosQuorum = strconv.Itoa((len(g.bcsStorageCluster.MasterIP) + 1) / 2)
	renderConfig.Dns = strings.Join(dns, " ")
	renderConfig.ClusterId = clusterID
	renderConfig.ClusterIdNum = lin[len(lin)-1]
	renderConfig.City = g.bcsStorageCluster.City
	renderConfig.JfrogUrl = g.bcsStorageCluster.JfrogUrl
	renderConfig.NeedNat = func() string {
		if g.bcsStorageCluster.NeedNat {
			return "true"
		}
		return "false"
	}()

	template, err := clusterconfig.GetTemplate(g.ctx, g.getTemplateCondition())
	if err != nil {
		return nil, err
	}

	str := renderConfig.Render(template)
	r := lib.CopyMap(operator.M{constants.ClusterIDTag: clusterID})
	blog.Infof("renderConfig data: %s", str)

	var data map[string]interface{}
	err = codec.DecJson([]byte(str), &data)
	if err != nil {
		return nil, err
	}

	r[constants.DataTag] = data
	return r, nil
}

func (g *general) putClsConfig() error {
	data, err := g.getReqData()
	if err != nil {
		return err
	}
	opt := &lib.StorePutOption{
		Cond:          g.getClusterIDCondition(),
		CreateTimeKey: constants.CreateTimeTag,
		UpdateTimeKey: constants.UpdateTimeTag,
	}
	return clusterconfig.SaveClusterInfoConfig(g.ctx, data, opt)
}

func getMultiCls(g *general) ([]operator.M, error) {
	mList, err := clusterconfig.GetClusterInfo(g.ctx, g.getMultiClsCondition())
	if err != nil {
		return nil, err
	}
	return mList, nil
}

func getClusterInfo(g *general) ([]operator.M, error) {
	mList, err := clusterconfig.GetClusterInfo(g.ctx, g.getClusterIDCondition())
	if err != nil {
		return nil, err
	}
	return mList, nil
}

// HandlerGetClusterConfig GetClusterConfig业务方法
func HandlerGetClusterConfig(ctx context.Context, req *storage.GetClusterConfigRequest) (config *types.DeployConfig,
	err error) {
	g := &general{
		ctx:       ctx,
		clsFunc:   getClusterInfo,
		service:   req.Service,
		clusterID: req.ClusterId,
	}

	return g.handler()
}

// HandlerPutClusterConfig PutClusterConfig业务方法
func HandlerPutClusterConfig(ctx context.Context, req *storage.PutClusterConfigRequest,
	rsp *storage.PutClusterConfigResponse) {
	g := &general{
		ctx:       ctx,
		clsFunc:   getClusterInfo,
		service:   req.Service,
		clusterID: req.ClusterId,
		bcsStorageCluster: &storage.BcsStorageCluster{
			Service:  req.Service,
			ZkIP:     req.ZkIP,
			MasterIP: req.MasterIP,
			DnsIP:    req.DnsIP,
			JfrogUrl: req.JfrogUrl,
			NeedNat:  req.NeedNat,
		},
	}

	if err := g.putClsConfig(); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("HandlerPutClusterConfig %s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return
	}

	data, err := g.handler()
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("HandlerPutClusterConfig %s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return
	}

	if err = util.MapToStruct(util.StructToMap(data), rsp.Data, "HandlerPutClusterConfig"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("HandlerPutClusterConfig %s | err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return
	}

	// 查询成功，并返回数据
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr
}

// HandlerGetServiceConfig GetServiceConfig业务方法
func HandlerGetServiceConfig(ctx context.Context, req *storage.GetServiceConfigRequest) (config *types.DeployConfig,
	err error) {
	g := &general{
		ctx:          ctx,
		service:      req.Service,
		clsFunc:      getMultiCls,
		clusterID:    req.ClusterId,
		clusterIDNot: req.ClusterIdNot,
	}

	return g.handler()
}

// HandlerGetStableVersion GetStableVersion业务方法
func HandlerGetStableVersion(ctx context.Context, req *storage.GetStableVersionRequest) (string, error) {
	g := &general{
		service: req.Service,
	}
	g.preService()
	opt := &lib.StoreGetOption{
		Cond: g.getSvcCondition(),
	}

	return clusterconfig.GetStableSvcVersion(ctx, opt)
}

// HandlerPutStableVersion PutStableVersion业务方法
func HandlerPutStableVersion(ctx context.Context, req *storage.PutStableVersionRequest) error {
	g := &general{
		ctx:     ctx,
		service: req.Service,
		version: req.Version,
	}

	return g.putStableVersion()
}

func wrapIP(s []string, df string) (r []string) {
	r = make([]string, 0, len(s))
	for _, v := range s {
		if !strings.Contains(v, ":") {
			v += ":" + df
		}
		r = append(r, v)
	}
	return r
}

func unwrapIP(s []string) (r []string) {
	r = make([]string, 0, len(s))
	for _, v := range s {
		if strings.Contains(v, ":") {
			v = strings.Split(v, ":")[0]
		}
		r = append(r, v)
	}
	return r
}
