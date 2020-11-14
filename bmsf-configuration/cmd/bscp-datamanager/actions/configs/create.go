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
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// CreateAction is configs create action object.
type CreateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateConfigsReq
	resp *pb.CreateConfigsResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateConfigsReq, resp *pb.CreateConfigsResp) *CreateAction {
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

	length = len(act.req.Cid)
	if length == 0 {
		return errors.New("invalid params, cid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cid too long")
	}

	if act.req.Content == nil {
		act.req.Content = []byte{}
	}

	if len(act.req.Content) > database.BSCPCFGCONTENTSSIZELIMIT {
		return fmt.Errorf("invalid params, content is too large[%d]", len(act.req.Content))
	}
	if len(act.req.CfgLink) > database.BSCPCFGLINKLENLIMIT {
		return fmt.Errorf("invalid params, configs link is too long[%d]", len(act.req.CfgLink))
	}

	length = len(act.req.Creator)
	if length == 0 {
		return errors.New("invalid params, creator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, creator missing")
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

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}

	return nil
}

func (act *CreateAction) createConfigs() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Configs{})

	st := database.Configs{
		Bid:          act.req.Bid,
		Appid:        act.req.Appid,
		Clusterid:    act.req.Clusterid,
		Zoneid:       act.req.Zoneid,
		Index:        act.req.Index,
		Cfgsetid:     act.req.Cfgsetid,
		Commitid:     act.req.Commitid,
		Cid:          act.req.Cid,
		CfgLink:      act.req.CfgLink,
		Content:      act.req.Content,
		State:        act.req.State,
		Creator:      act.req.Creator,
		Memo:         act.req.Memo,
		LastModifyBy: act.req.Creator,
	}

	err := act.sd.DB().
		Where(database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.req.Commitid, Appid: act.req.Appid}).
		Where("Fclusterid = ?", act.req.Clusterid).
		Where("Fzoneid = ?", act.req.Zoneid).
		Where("Findex = ?", act.req.Index).
		Assign(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
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

	// create configs.
	if errCode, errMsg := act.createConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
