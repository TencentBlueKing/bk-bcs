/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package templateversion

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// QueryAction action for query config template version
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager
	sd    *dbsharding.ShardingDB

	req  *pb.QueryTemplateVersionReq
	resp *pb.QueryTemplateVersionResp

	templateVersion database.ConfigTemplateVersion
}

// NewQueryAction create new QueryAction
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryTemplateVersionReq, resp *pb.QueryTemplateVersionResp) *QueryAction {
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
	templateVersion := &pbcommon.TemplateVersion{
		Bid:          act.templateVersion.Bid,
		Templateid:   act.templateVersion.Templateid,
		Versionid:    act.templateVersion.Versionid,
		VersionName:  act.templateVersion.VersionName,
		Content:      act.templateVersion.Content,
		Memo:         act.templateVersion.Memo,
		Creator:      act.templateVersion.Creator,
		LastModifyBy: act.templateVersion.LastModifyBy,
		State:        act.templateVersion.State,
		CreatedAt:    act.templateVersion.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    act.templateVersion.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.TemplateVersion = templateVersion
	return nil
}

func (act *QueryAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Versionid, "versionid"); err != nil {
		return err
	}

	return nil
}

func (act *QueryAction) queryTemplateVersion() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplateVersion{})

	err := act.sd.DB().
		Where(map[string]interface{}{
			"Fbid":       act.req.Bid,
			"Fversionid": act.req.Versionid,
			"Fstate":     int32(pbcommon.ConfigTemplateVersionState_CTVS_CREATED),
		}).
		Last(&act.templateVersion).Error

	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "config template version no found"
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *QueryAction) Do() error {
	// business sharding db
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query config template version.
	if errCode, errMsg := act.queryTemplateVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
