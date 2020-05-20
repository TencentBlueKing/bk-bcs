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

package strategy

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// CreateAction is strategy create action object.
type CreateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateStrategyReq
	resp *pb.CreateStrategyResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateStrategyReq, resp *pb.CreateStrategyResp) *CreateAction {
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

	length = len(act.req.Strategyid)
	if length == 0 {
		return errors.New("invalid params, strategyid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, strategyid too long")
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

	length = len(act.req.Content)
	if length == 0 {
		return errors.New("invalid params, content missing")
	}
	if length > database.BSCPSTRATEGYCONTENTSIZELIMIT {
		return errors.New("invalid params, content too big")
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

func (act *CreateAction) reCreateStrategy() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Strategy{})

	ups := map[string]interface{}{
		"Strategyid":   act.req.Strategyid,
		"State":        int32(pbcommon.StrategyState_SS_CREATED),
		"Content":      act.req.Content,
		"Memo":         act.req.Memo,
		"Creator":      act.req.Creator,
		"LastModifyBy": act.req.Creator,
	}

	exec := act.sd.DB().
		Model(&database.Strategy{}).
		Where(&database.Strategy{Appid: act.req.Appid, Name: act.req.Name}).
		Where("Fstate = ?", pbcommon.StrategyState_SS_DELETED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate the strategy failed, there is no strategy that fit the conditions."
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createStrategy() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Strategy{})

	st := database.Strategy{
		Strategyid:   act.req.Strategyid,
		Name:         act.req.Name,
		Appid:        act.req.Appid,
		Content:      act.req.Content,
		State:        int32(pbcommon.StrategyState_SS_CREATED),
		Creator:      act.req.Creator,
		LastModifyBy: act.req.Creator,
		Memo:         act.req.Memo,
	}

	err := act.sd.DB().
		Where(database.Strategy{Appid: act.req.Appid, Name: act.req.Name}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Strategyid = st.Strategyid

	if st.Strategyid != act.req.Strategyid {
		if st.State == int32(pbcommon.StrategyState_SS_CREATED) {
			return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "the strategy with target name already exist."
		}

		if errCode, errMsg := act.reCreateStrategy(); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
		act.resp.Strategyid = act.req.Strategyid
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

	// create strategy.
	if errCode, errMsg := act.createStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
