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

package configset

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// QueryAction is configset query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigSetReq
	resp *pb.QueryConfigSetResp

	sd *dbsharding.ShardingDB

	configSet database.ConfigSet
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigSetReq, resp *pb.QueryConfigSetResp) *QueryAction {
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
	configset := &pbcommon.ConfigSet{
		Bid:          act.configSet.Bid,
		Cfgsetid:     act.configSet.Cfgsetid,
		Appid:        act.configSet.Appid,
		Name:         act.configSet.Name,
		Fpath:        act.configSet.Fpath,
		Creator:      act.configSet.Creator,
		LastModifyBy: act.configSet.LastModifyBy,
		Memo:         act.configSet.Memo,
		State:        act.configSet.State,
		CreatedAt:    act.configSet.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    act.configSet.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.ConfigSet = configset
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

	if len(act.req.Cfgsetid) == 0 && (len(act.req.Appid) == 0 || len(act.req.Name) == 0) {
		return errors.New("invalid params, cfgsetid or appid-name-fpath missing")
	}

	if len(act.req.Cfgsetid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.Name) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	if len(act.req.Cfgsetid) != 0 {
		// maybe empty fpath param, do not parse fpath to /
		act.req.Fpath = ""
	} else {
		act.req.Fpath = common.ParseFpath(act.req.Fpath)
		if len(act.req.Fpath) > database.BSCPCFGSETFPATHLENLIMIT {
			return errors.New("invalid params, fpath too long")
		}
	}
	return nil
}

func (act *QueryAction) queryConfigSet() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigSet{})

	err := act.sd.DB().
		Where(&database.ConfigSet{Appid: act.req.Appid, Cfgsetid: act.req.Cfgsetid, Name: act.req.Name, Fpath: act.req.Fpath}).
		Where("Fstate = ?", pbcommon.ConfigSetState_CSS_CREATED).
		Last(&act.configSet).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "configset non-exist."
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

	// query configset.
	if errCode, errMsg := act.queryConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
