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

package construction

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/integrator"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
)

// ConstructAction constructs application with cluster/zone.
type ConstructAction struct {
	viper          *viper.Viper
	businessSvrCli pbbusinessserver.BusinessClient
	md             *structs.IntegrationMetadata

	req  *pb.IntegrateReq
	resp *pb.IntegrateResp

	business *pbcommon.Business
	appid    string
}

// NewConstructAction creates new ConstructAction.
func NewConstructAction(viper *viper.Viper, businessSvrCli pbbusinessserver.BusinessClient, md *structs.IntegrationMetadata,
	req *pb.IntegrateReq, resp *pb.IntegrateResp) *ConstructAction {
	action := &ConstructAction{viper: viper, businessSvrCli: businessSvrCli, md: md, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ConstructAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ConstructAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_ITG_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ConstructAction) Output() error {
	// do nothing.
	return nil
}

func (act *ConstructAction) verify() error {
	length := len(act.md.Spec.BusinessName)
	if length == 0 {
		return errors.New("invalid params, businessName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, businessName too long")
	}

	length = len(act.md.Spec.AppName)
	if length == 0 {
		return errors.New("invalid params, appName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, appName too long")
	}

	if act.md.Spec.DeployType != int32(pbcommon.DeployType_DT_BCS) &&
		act.md.Spec.DeployType != int32(pbcommon.DeployType_DT_GSE) {
		return errors.New("invalid params, unknow deployType, 0:bcs  1:gse")
	}

	if len(act.md.Spec.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}

	for clusterItemid, cluster := range act.md.Construction.Clusters {
		length := len(cluster.Name)
		if length == 0 {
			return fmt.Errorf("invalid params, construction cluster item[%d], name missing", clusterItemid)
		}
		if length > database.BSCPNAMELENLIMIT {
			return fmt.Errorf("invalid params, construction cluster item[%d], name too long", clusterItemid)
		}

		if len(cluster.RClusterid) > database.BSCPLONGSTRLENLIMIT {
			return fmt.Errorf("invalid params, construction cluster item[%d], rclusterid too long", clusterItemid)
		}

		if len(cluster.Memo) > database.BSCPLONGSTRLENLIMIT {
			return fmt.Errorf("invalid params, construction cluster item[%d], memo too long", clusterItemid)
		}

		for zoneItemid, zone := range cluster.Zones {
			length := len(zone.Name)
			if length == 0 {
				return fmt.Errorf("invalid params, construction zone item[%d-%d], name missing", clusterItemid, zoneItemid)
			}
			if length > database.BSCPNAMELENLIMIT {
				return fmt.Errorf("invalid params, construction zone item[%d-%d], name too long", clusterItemid, zoneItemid)
			}

			if len(cluster.Memo) > database.BSCPLONGSTRLENLIMIT {
				return fmt.Errorf("invalid params, construction cluster item[%d-%d], memo too long", clusterItemid, zoneItemid)
			}
		}
	}
	return nil
}

func (act *ConstructAction) queryBusiness() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryBusinessReq{
		Seq:  act.req.Seq,
		Name: act.md.Spec.BusinessName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Construct[%d]| request to businessserver QueryBusiness, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryBusiness, %+v", err)
	}
	act.business = resp.Business

	return resp.ErrCode, resp.ErrMsg
}

func (act *ConstructAction) createApp() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateAppReq{Seq: act.req.Seq,
		Bid:        act.business.Bid,
		Name:       act.md.Spec.AppName,
		DeployType: act.md.Spec.DeployType,
		Memo:       act.md.Spec.Memo,
		Creator:    act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Construct[%d]| request to businessserver CreateApp, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateApp, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK && resp.ErrCode != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return resp.ErrCode, resp.ErrMsg
	}
	act.appid = resp.Appid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ConstructAction) createCluster(name, rClusterid, memo string) (string, error) {
	r := &pbbusinessserver.CreateClusterReq{
		Seq:        act.req.Seq,
		Bid:        act.business.Bid,
		Name:       name,
		Appid:      act.appid,
		RClusterid: rClusterid,
		Memo:       memo,
		Creator:    act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Construct[%d]| request to businessserver CreateCluster, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateCluster(ctx, r)
	if err != nil {
		return "", err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK && resp.ErrCode != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Clusterid, nil
}

func (act *ConstructAction) createZone(name, clusterid, memo string) (string, error) {
	r := &pbbusinessserver.CreateZoneReq{
		Seq:       act.req.Seq,
		Bid:       act.business.Bid,
		Appid:     act.appid,
		Clusterid: clusterid,
		Name:      name,
		Memo:      memo,
		Creator:   act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Construct[%d]| request to businessserver CreateZone, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateZone(ctx, r)
	if err != nil {
		return "", err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK && resp.ErrCode != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Zoneid, nil
}

func (act *ConstructAction) construct() (pbcommon.ErrCode, string) {
	// constructs cluster level.
	for _, cluster := range act.md.Construction.Clusters {
		clusterid, err := act.createCluster(cluster.Name, cluster.RClusterid, cluster.Memo)
		if err != nil {
			return pbcommon.ErrCode_E_ITG_CONSTRUCT_CLUSTER_FAILED, fmt.Sprintf("construct cluster[%+v], %+v", cluster.Name, err)
		}

		// construct zone level under target cluster.
		for _, zone := range cluster.Zones {
			if _, err := act.createZone(zone.Name, clusterid, zone.Memo); err != nil {
				return pbcommon.ErrCode_E_ITG_CONSTRUCT_ZONE_FAILED, fmt.Sprintf("construct zone[%+v] under cluster[%+v], %+v",
					zone.Name, cluster.Name, err)
			}
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ConstructAction) Do() error {
	// query business information used for application construction.
	if errCode, errMsg := act.queryBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create application if no-exist.
	if errCode, errMsg := act.createApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// construct target application base on metadata desc.
	if errCode, errMsg := act.construct(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
