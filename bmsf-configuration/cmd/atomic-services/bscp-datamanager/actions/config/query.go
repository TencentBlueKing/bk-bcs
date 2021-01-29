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

package config

import (
	"context"
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// QueryAction is config query action object.
type QueryAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigReq
	resp *pb.QueryConfigResp

	sd *dbsharding.ShardingDB

	config database.Config
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigReq, resp *pb.QueryConfigResp) *QueryAction {
	action := &QueryAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
	config := &pbcommon.Config{
		BizId:         act.config.BizID,
		CfgId:         act.config.CfgID,
		AppId:         act.config.AppID,
		Name:          act.config.Name,
		Fpath:         act.config.Fpath,
		User:          act.config.User,
		UserGroup:     act.config.UserGroup,
		FilePrivilege: act.config.FilePrivilege,
		FileFormat:    act.config.FileFormat,
		FileMode:      act.config.FileMode,
		Creator:       act.config.Creator,
		LastModifyBy:  act.config.LastModifyBy,
		Memo:          act.config.Memo,
		State:         act.config.State,
		CreatedAt:     act.config.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     act.config.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.Data = config
	return nil
}

func (act *QueryAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.CfgId) == 0 && (len(act.req.AppId) == 0 || len(act.req.Name) == 0) {
		return errors.New("invalid input data, cfg_id or app_id/name/fpath is required")
	}

	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("name", act.req.Name,
		database.BSCPEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	if len(act.req.CfgId) != 0 {
		// maybe empty fpath param, do not parse fpath to /
		act.req.Fpath = ""
	} else {
		act.req.Fpath = common.ParseFpath(act.req.Fpath)
		if err = common.ValidateString("fpath", act.req.Fpath,
			database.BSCPEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
			return err
		}
	}
	return nil
}

func (act *QueryAction) queryConfig() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Where(&database.Config{
			BizID: act.req.BizId,
			AppID: act.req.AppId,
			CfgID: act.req.CfgId,
			Name:  act.req.Name,
			Fpath: act.req.Fpath,
		}).
		Last(&act.config).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "config non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query config.
	if errCode, errMsg := act.queryConfig(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
