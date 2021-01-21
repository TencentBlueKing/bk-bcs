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

package content

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// CreateAction is cont ent create action object.
type CreateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateConfigContentReq
	resp *pb.CreateConfigContentResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateConfigContentReq, resp *pb.CreateConfigContentResp) *CreateAction {
	action := &CreateAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("commit_id", act.req.CommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	act.req.ContentId = strings.ToUpper(act.req.ContentId)
	if err = common.ValidateString("content_id", act.req.ContentId,
		database.BSCPCONTENTIDLENLIMIT, database.BSCPCONTENTIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("creator", act.req.Creator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("index", act.req.Index,
		database.BSCPNOTEMPTY, database.BSCPSTRATEGYCONTENTSIZELIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) createConfigContent() (pbcommon.ErrCode, string) {
	st := database.Content{
		BizID:        act.req.BizId,
		AppID:        act.req.AppId,
		CfgID:        act.req.CfgId,
		CommitID:     act.req.CommitId,
		ContentID:    act.req.ContentId,
		ContentSize:  uint64(act.req.ContentSize),
		Index:        act.req.Index,
		State:        act.req.State,
		Creator:      act.req.Creator,
		Memo:         act.req.Memo,
		LastModifyBy: act.req.Creator,
	}

	err := act.sd.DB().
		Create(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// create configs.
	if errCode, errMsg := act.createConfigContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
