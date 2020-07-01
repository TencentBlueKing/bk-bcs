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

package metadata

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// QueryAction query metadata for bcs sidecar instance.
type QueryAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryAppMetadataReq
	resp *pb.QueryAppMetadataResp
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryAppMetadataReq, resp *pb.QueryAppMetadataResp) *QueryAction {
	action := &QueryAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CONNS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	// do nothing.
	return nil
}

func (act *QueryAction) verify() error {
	length := len(act.req.BusinessName)
	if length == 0 {
		return errors.New("invalid params, businessName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, businessName too long")
	}

	length = len(act.req.AppName)
	if length == 0 {
		return errors.New("invalid params, appName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, appName too long")
	}

	length = len(act.req.ClusterName)
	if length == 0 {
		return errors.New("invalid params, clusterName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, clusterName too long")
	}

	length = len(act.req.ZoneName)
	if length == 0 {
		return errors.New("invalid params, zoneName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, zoneName too long")
	}
	return nil
}

func (act *QueryAction) query() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppMetadataReq{
		Seq:           act.req.Seq,
		BusinessName:  act.req.BusinessName,
		AppName:       act.req.AppName,
		ClusterName:   act.req.ClusterName,
		ZoneName:      act.req.ZoneName,
		ClusterLabels: act.req.ClusterLabels,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryAppMetadata[%d]| request to datamanager QueryAppMetadata, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryAppMetadata(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryAppMetadata, %+v", err)
	}

	act.resp.Bid = resp.Bid
	act.resp.Appid = resp.Appid
	act.resp.Clusterid = resp.Clusterid
	act.resp.Zoneid = resp.Zoneid

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// query app metadata.
	if errCode, errMsg := act.query(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
