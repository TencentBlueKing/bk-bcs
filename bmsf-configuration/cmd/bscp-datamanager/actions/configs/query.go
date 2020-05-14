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

package configs

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// QueryAction is configs query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigsReq
	resp *pb.QueryConfigsResp

	sd *dbsharding.ShardingDB

	configs database.Configs
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigsReq, resp *pb.QueryConfigsResp) *QueryAction {
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
	configs := &pbcommon.Configs{
		Bid:          act.configs.Bid,
		Cfgsetid:     act.configs.Cfgsetid,
		Appid:        act.configs.Appid,
		Clusterid:    act.configs.Clusterid,
		Zoneid:       act.configs.Zoneid,
		Commitid:     act.configs.Commitid,
		Cid:          act.configs.Cid,
		CfgLink:      act.configs.CfgLink,
		Content:      act.configs.Content,
		Creator:      act.configs.Creator,
		LastModifyBy: act.configs.LastModifyBy,
		Memo:         act.configs.Memo,
		State:        act.configs.State,
		CreatedAt:    act.configs.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    act.configs.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.Configs = configs
	return nil
}

func (act *QueryAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}

	if len(act.req.Clusterid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, clusterid too long")
	}

	if len(act.req.Zoneid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, zoneid too long")
	}

	if len(act.req.Index) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, index too long")
	}

	return nil
}

func (act *QueryAction) queryConfigs() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Configs{})

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.req.Commitid}).
		Where("Fappid = ?", act.req.Appid).
		Where("Fclusterid = ?", act.req.Clusterid).
		Where("Fzoneid = ?", act.req.Zoneid).
		Where("Findex = ?", act.req.Index).
		Last(&act.configs).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "configs non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query configs.
	if errCode, errMsg := act.queryConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
