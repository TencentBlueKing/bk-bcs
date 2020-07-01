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
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// QueryAction app metadata query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryAppMetadataReq
	resp *pb.QueryAppMetadataResp

	sd *dbsharding.ShardingDB

	business database.Business
	app      database.App
	cluster  database.Cluster
	zone     database.Zone
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryAppMetadataReq, resp *pb.QueryAppMetadataResp) *QueryAction {
	action := &QueryAction{viper: viper, smgr: smgr, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	act.resp.Bid = act.business.Bid
	act.resp.Appid = act.app.Appid
	act.resp.Clusterid = act.cluster.Clusterid
	act.resp.Zoneid = act.zone.Zoneid
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

	if len(act.req.ClusterLabels) > database.BSCPCLUSTERLABELSLENLIMIT {
		return errors.New("invalid params, clusterLabels too long")
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

func (act *QueryAction) queryBusiness() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Business{})

	err := act.sd.DB().
		Where(&database.Business{Name: act.req.BusinessName}).
		Last(&act.business).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "business non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryAction) queryApp() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.App{})

	err := act.sd.DB().
		Where(&database.App{Bid: act.business.Bid, Name: act.req.AppName}).
		Where("Fstate = ?", pbcommon.AppState_AS_CREATED).
		Last(&act.app).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "app non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryAction) queryCluster() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Cluster{})

	err := act.sd.DB().
		Where(&database.Cluster{Bid: act.business.Bid, Appid: act.app.Appid, Name: act.req.ClusterName}).
		Where("Flabels = ?", act.req.ClusterLabels).
		Where("Fstate = ?", pbcommon.ClusterState_CS_CREATED).
		Last(&act.cluster).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "cluster non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryAction) queryZone() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Zone{})

	err := act.sd.DB().
		Where(&database.Zone{Bid: act.business.Bid, Appid: act.app.Appid,
			Clusterid: act.cluster.Clusterid, Name: act.req.ZoneName}).
		Where("Fstate = ?", pbcommon.ZoneState_ZS_CREATED).
		Last(&act.zone).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "zone non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// BSCP sharding db.
	sd, err := act.smgr.ShardingDB(dbsharding.BSCPDBKEY)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query business information.
	if errCode, errMsg := act.queryBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// business sharding db.
	sd, err = act.smgr.ShardingDB(act.business.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query app information.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query cluster information.
	if errCode, errMsg := act.queryCluster(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query zone information.
	if errCode, errMsg := act.queryZone(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
