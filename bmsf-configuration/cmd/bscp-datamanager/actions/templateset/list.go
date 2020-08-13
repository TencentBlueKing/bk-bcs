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

package templateset

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// ListAction action to list config template set
type ListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigTemplateSetListReq
	resp *pb.QueryConfigTemplateSetListResp

	sd *dbsharding.ShardingDB

	templateSets []database.ConfigTemplateSet
}

// NewListAction create new ListAction
func NewListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigTemplateSetListReq, resp *pb.QueryConfigTemplateSetListResp) *ListAction {
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
	configTemplateSets := []*pbcommon.ConfigTemplateSet{}
	for _, tplSet := range act.templateSets {
		configTemplateSet := &pbcommon.ConfigTemplateSet{
			Bid:          tplSet.Bid,
			Setid:        tplSet.Setid,
			Name:         tplSet.Name,
			Fpath:        tplSet.Fpath,
			Creator:      tplSet.Creator,
			LastModifyBy: tplSet.LastModifyBy,
			Memo:         tplSet.Memo,
			State:        tplSet.State,
			CreatedAt:    tplSet.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    tplSet.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		configTemplateSets = append(configTemplateSets, configTemplateSet)
	}
	act.resp.TemplateSets = configTemplateSets
	return nil
}

func (act *ListAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyQueryLimit(act.req.Limit); err != nil {
		return err
	}

	return nil
}

func (act *ListAction) queryConfigTemplateSetList() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplateSet{})

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(map[string]interface{}{
			"Fbid":   act.req.Bid,
			"Fstate": int32(pbcommon.ConfigTemplateSetState_CTSS_CREATED),
		}).
		Find(&act.templateSets).Error

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

	// query config template set list
	if errCode, errMsg := act.queryConfigTemplateSetList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
