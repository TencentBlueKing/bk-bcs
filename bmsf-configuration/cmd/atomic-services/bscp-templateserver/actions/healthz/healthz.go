/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package healthz

import (
	"context"
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/healthz"
	"bk-bscp/internal/healthz/bkrepo"
	"bk-bscp/internal/healthz/etcd"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/version"
)

// Action healthz check action object.
type Action struct {
	ctx   context.Context
	viper *viper.Viper

	req  *pb.HealthzReq
	resp *pb.HealthzResp

	componentHealthzInfos []*pbcommon.ComponentHealthzInfo
}

// NewAction creates new Action.
func NewAction(ctx context.Context, viper *viper.Viper,
	req *pb.HealthzReq, resp *pb.HealthzResp) *Action {
	action := &Action{ctx: ctx, viper: viper, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *Action) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *Action) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *Action) Output() error {
	// do nothing.
	return nil
}

func (act *Action) verify() error {
	// do nothing.
	return nil
}

func (act *Action) healthzBKRepo() {
	info := &pbcommon.ComponentHealthzInfo{
		Component: healthz.ComponentNameBKRepo,
		IsHealthy: true,
		Message:   healthz.HealthStateMessage,
	}

	isHealthy, err := bkrepoz.Healthz(act.viper.GetString("bkrepo.host"))
	if err != nil {
		info.IsHealthy = false
		info.Message = err.Error()
		return
	}

	if !isHealthy {
		info.IsHealthy = false
		info.Message = "bkrepo is not working"
		return
	}

	act.componentHealthzInfos = append(act.componentHealthzInfos, info)
}

func (act *Action) healthzEtcd() {
	info := &pbcommon.ComponentHealthzInfo{
		Component: healthz.ComponentNameEtcd,
		IsHealthy: true,
		Message:   healthz.HealthStateMessage,
	}

	isHealthy, err := etcdz.Healthz(act.viper.GetStringSlice("etcdCluster.endpoints")[0],
		act.viper.GetString("etcdCluster.tls.caFile"),
		act.viper.GetString("etcdCluster.tls.certFile"),
		act.viper.GetString("etcdCluster.tls.keyFile"),
		act.viper.GetString("etcdCluster.tls.certPassword"))
	if err != nil {
		info.IsHealthy = false
		info.Message = err.Error()
		return
	}

	if !isHealthy {
		info.IsHealthy = false
		info.Message = "etcd cluster is not working"
		return
	}

	act.componentHealthzInfos = append(act.componentHealthzInfos, info)
}

// Do makes the workflows of this action base on input messages.
func (act *Action) Do() error {
	// bkrepo health state.
	act.healthzBKRepo()

	// etcd health state.
	act.healthzEtcd()

	info := &pbcommon.ModuleHealthzInfo{
		Module:    "bk-bscp-templateserver",
		Version:   version.VERSION,
		BuildTime: version.BUILDTIME,
		GitHash:   version.GITHASH,
		IsHealthy: true,
		Message:   healthz.HealthStateMessage,
	}

	for _, component := range act.componentHealthzInfos {
		if !component.IsHealthy {
			info.IsHealthy = false
			info.Message = component.Message
		}
	}
	info.Components = act.componentHealthzInfos
	act.resp.Data = info

	return nil
}
