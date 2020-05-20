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

// CreateAction is configset create action object.
type CreateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateConfigSetReq
	resp *pb.CreateConfigSetResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateConfigSetReq, resp *pb.CreateConfigSetResp) *CreateAction {
	action := &CreateAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	length = len(act.req.Name)
	if length == 0 {
		return errors.New("invalid params, name missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	act.req.Fpath = common.ParseFpath(act.req.Fpath)
	if len(act.req.Fpath) > database.BSCPCFGSETFPATHLENLIMIT {
		return errors.New("invalid params, fpath too long")
	}

	length = len(act.req.Creator)
	if length == 0 {
		return errors.New("invalid params, creator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, creator too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) reCreateConfigSet() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigSet{})

	ups := map[string]interface{}{
		"Bid":          act.req.Bid,
		"Cfgsetid":     act.req.Cfgsetid,
		"State":        int32(pbcommon.ConfigSetState_CSS_CREATED),
		"Memo":         act.req.Memo,
		"Creator":      act.req.Creator,
		"LastModifyBy": act.req.Creator,
	}

	exec := act.sd.DB().
		Model(&database.ConfigSet{}).
		Where(&database.ConfigSet{Appid: act.req.Appid, Name: act.req.Name, Fpath: act.req.Fpath}).
		Where("Fstate = ?", pbcommon.ConfigSetState_CSS_DELETED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate the configset failed, there is no configset that fit the conditions."
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createConfigSet() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigSet{})

	st := database.ConfigSet{
		Bid:          act.req.Bid,
		Appid:        act.req.Appid,
		Cfgsetid:     act.req.Cfgsetid,
		Name:         act.req.Name,
		Fpath:        act.req.Fpath,
		State:        int32(pbcommon.ConfigSetState_CSS_CREATED),
		Creator:      act.req.Creator,
		Memo:         act.req.Memo,
		LastModifyBy: act.req.Creator,
	}

	err := act.sd.DB().
		Where(database.ConfigSet{Appid: act.req.Appid, Name: act.req.Name, Fpath: act.req.Fpath}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Cfgsetid = st.Cfgsetid

	if st.Cfgsetid != act.req.Cfgsetid {
		if st.State == int32(pbcommon.ConfigSetState_CSS_CREATED) {
			return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "the configset with target fpath+name already exist."
		}

		// other states: already exist but deleted.
		if errCode, errMsg := act.reCreateConfigSet(); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
		act.resp.Cfgsetid = act.req.Cfgsetid
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// create configset.
	if errCode, errMsg := act.createConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
