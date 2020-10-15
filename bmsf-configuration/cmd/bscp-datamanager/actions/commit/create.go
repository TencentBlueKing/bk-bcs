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

package commit

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// CreateAction is commit create action object.
type CreateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateCommitReq
	resp *pb.CreateCommitResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateCommitReq, resp *pb.CreateCommitResp) *CreateAction {
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

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
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

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if act.req.Configs == nil {
		act.req.Configs = []byte{}
	}

	if len(act.req.Configs) > database.BSCPCONFIGSSIZELIMIT {
		return errors.New("invalid params, configs content too big")
	}
	if len(act.req.Changes) > database.BSCPCHANGESSIZELIMIT {
		return errors.New("invalid params, configs changes too big")
	}

	if len(act.req.Templateid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, templateid too long")
	}
	if len(act.req.Template) > database.BSCPTPLSIZELIMIT {
		return errors.New("invalid params, template size too big")
	}
	if len(act.req.TemplateRule) > database.BSCPTPLRULESSIZELIMIT {
		return errors.New("invalid params, template rules too long")
	}

	if len(act.req.Configs) != 0 && len(act.req.Template) != 0 {
		return errors.New("invalid params, configs and template concurrence")
	}
	if len(act.req.Configs) != 0 && len(act.req.Templateid) != 0 {
		return errors.New("invalid params, configs and templateid concurrence")
	}
	if len(act.req.Template) != 0 && len(act.req.Templateid) != 0 {
		return errors.New("invalid params, template and templateid concurrence")
	}
	if len(act.req.Template) != 0 && len(act.req.TemplateRule) == 0 {
		return errors.New("invalid params, empty template rules")
	}

	if len(act.req.MultiCommitid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multi commitid too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) createCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	commit := &database.Commit{
		Bid:          act.req.Bid,
		Appid:        act.req.Appid,
		Cfgsetid:     act.req.Cfgsetid,
		Commitid:     act.req.Commitid,
		Op:           act.req.Op,
		Operator:     act.req.Operator,
		Templateid:   act.req.Templateid,
		Template:     act.req.Template,
		TemplateRule: act.req.TemplateRule,
		PrevConfigs:  []byte{},
		Configs:      act.req.Configs,
		Changes:      act.req.Changes,
		Memo:         act.req.Memo,
		State:        int32(pbcommon.CommitState_CS_INIT),
		// normale mode, multi commitid is current commitid(unique index).
		MultiCommitid: act.req.Commitid,
	}

	err := act.sd.DB().
		Create(commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Commitid = act.req.Commitid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createCommitMultiMode() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	commit := database.Commit{
		Bid:           act.req.Bid,
		Appid:         act.req.Appid,
		Cfgsetid:      act.req.Cfgsetid,
		Commitid:      act.req.Commitid,
		Op:            act.req.Op,
		Operator:      act.req.Operator,
		Templateid:    act.req.Templateid,
		Template:      act.req.Template,
		TemplateRule:  act.req.TemplateRule,
		PrevConfigs:   []byte{},
		Configs:       act.req.Configs,
		Changes:       act.req.Changes,
		Memo:          act.req.Memo,
		State:         int32(pbcommon.CommitState_CS_INIT),
		MultiCommitid: act.req.MultiCommitid,
	}

	err := act.sd.DB().
		Where(database.Commit{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Appid: act.req.Appid, MultiCommitid: act.req.MultiCommitid}).
		Assign(commit).
		FirstOrCreate(&commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Commitid = act.req.Commitid

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

	if len(act.req.MultiCommitid) == 0 {
		// create commit.
		if errCode, errMsg := act.createCommit(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		// create commit in multi mode.
		if errCode, errMsg := act.createCommitMultiMode(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}
	return nil
}
