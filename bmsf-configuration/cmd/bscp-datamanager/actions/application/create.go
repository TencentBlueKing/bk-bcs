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

package application

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// CreateAction is app create action object.
type CreateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateAppReq
	resp *pb.CreateAppResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateAppReq, resp *pb.CreateAppResp) *CreateAction {
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

	if act.req.DeployType != int32(pbcommon.DeployType_DT_BCS) &&
		act.req.DeployType != int32(pbcommon.DeployType_DT_GSE) &&
		act.req.DeployType != int32(pbcommon.DeployType_DT_GSE_PLUGIN) {
		return errors.New("invalid params, unknow deployType, 0:bcs  1:gse 2:gse plugin")
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

func (act *CreateAction) reCreateApp() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.App{})

	ups := map[string]interface{}{
		"Appid":        act.req.Appid,
		"State":        int32(pbcommon.AppState_AS_CREATED),
		"DeployType":   act.req.DeployType,
		"Memo":         act.req.Memo,
		"Creator":      act.req.Creator,
		"LastModifyBy": act.req.Creator,
	}

	exec := act.sd.DB().
		Model(&database.App{}).
		Where(&database.App{Bid: act.req.Bid, Name: act.req.Name}).
		Where("Fstate = ?", pbcommon.AppState_AS_DELETED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate the app failed, there is no app that fit the conditions."
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createApp() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.App{})

	st := database.App{
		Appid:        act.req.Appid,
		Name:         act.req.Name,
		Bid:          act.req.Bid,
		DeployType:   act.req.DeployType,
		State:        int32(pbcommon.AppState_AS_CREATED),
		Creator:      act.req.Creator,
		LastModifyBy: act.req.Creator,
		Memo:         act.req.Memo,
	}

	err := act.sd.DB().
		Where(database.App{Bid: act.req.Bid, Name: act.req.Name}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Appid = st.Appid

	if st.Appid != act.req.Appid {
		if st.State == int32(pbcommon.AppState_AS_CREATED) {
			return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "the app with target name already exist."
		}

		if errCode, errMsg := act.reCreateApp(); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
		act.resp.Appid = act.req.Appid
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

	// create app.
	if errCode, errMsg := act.createApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
