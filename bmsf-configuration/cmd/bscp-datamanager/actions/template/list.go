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

package template

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// ListAction action to list config template
type ListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigTemplateListReq
	resp *pb.QueryConfigTemplateListResp

	sd *dbsharding.ShardingDB

	templates []database.ConfigTemplate
}

// NewListAction create new ListAction
func NewListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigTemplateListReq, resp *pb.QueryConfigTemplateListResp) *ListAction {
	action := &ListAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ListAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ListAction) Output() error {
	configTemplates := []*pbcommon.ConfigTemplate{}
	for _, tpl := range act.templates {
		configTemplate := &pbcommon.ConfigTemplate{
			Bid:          tpl.Bid,
			Setid:        tpl.Setid,
			Templateid:   tpl.Templateid,
			Name:         tpl.Name,
			Fpath:        tpl.Fpath,
			User:         tpl.User,
			Group:        tpl.Group,
			Permission:   tpl.Permission,
			FileEncoding: tpl.FileEncoding,
			EngineType:   tpl.EngineType,
			Creator:      tpl.Creator,
			LastModifyBy: tpl.LastModifyBy,
			Memo:         tpl.Memo,
			State:        tpl.State,
			CreatedAt:    tpl.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    tpl.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		configTemplates = append(configTemplates, configTemplate)
	}
	act.resp.ConfigTemplates = configTemplates
	return nil
}

func (act *ListAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return nil
	}

	if err := common.VerifyID(act.req.Setid, "setid"); err != nil {
		return nil
	}

	if err := common.VerifyQueryLimit(act.req.Limit); err != nil {
		return nil
	}

	return nil
}

func (act *ListAction) queryConfigTemplateList() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplate{})

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(map[string]interface{}{
			"Fbid":   act.req.Bid,
			"Fsetid": act.req.Setid,
			"Fstate": int32(pbcommon.ConfigTemplateState_CTS_CREATED),
		}).
		Find(&act.templates).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *ListAction) Do() error {
	// business sharding db
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query config template list
	if errCode, errMsg := act.queryConfigTemplateList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
