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
	"context"
	"errors"
	"math"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// ListAction action to list config template version.
type ListAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigTemplateVersionListReq
	resp *pb.QueryConfigTemplateVersionListResp

	sd *dbsharding.ShardingDB

	totalCount       int64
	templateVersions []database.ConfigTemplateVersion
}

// NewListAction create new ListAction
func NewListAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigTemplateVersionListReq, resp *pb.QueryConfigTemplateVersionListResp) *ListAction {
	action := &ListAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
	templateVersions := []*pbcommon.ConfigTemplateVersion{}
	for _, version := range act.templateVersions {
		templateVersion := &pbcommon.ConfigTemplateVersion{
			BizId:        version.BizID,
			TemplateId:   version.TemplateID,
			VersionId:    version.VersionID,
			VersionTag:   version.VersionTag,
			ContentId:    version.ContentID,
			ContentSize:  uint32(version.ContentSize),
			Memo:         version.Memo,
			Creator:      version.Creator,
			LastModifyBy: version.LastModifyBy,
			State:        version.State,
			CreatedAt:    version.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    version.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		templateVersions = append(templateVersions, templateVersion)
	}
	act.resp.Data = &pb.QueryConfigTemplateVersionListResp_RespData{
		TotalCount: uint32(act.totalCount),
		Info:       templateVersions,
	}
	return nil
}

func (act *ListAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("template_id", act.req.TemplateId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if act.req.Page == nil {
		return errors.New("invalid input data, page is required")
	}
	if err = common.ValidateInt32("page.start", act.req.Page.Start, 0, math.MaxInt32); err != nil {
		return err
	}
	if err = common.ValidateInt32("page.limit", act.req.Page.Limit, 1, database.BSCPQUERYLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *ListAction) queryTemplateVersionCount() (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, "OK"
	}

	err := act.sd.DB().
		Model(&database.ConfigTemplateVersion{}).
		Where(&database.ConfigTemplateVersion{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Count(&act.totalCount).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *ListAction) queryTemplateVersionList() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.ConfigTemplateVersion{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Find(&act.templateVersions).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *ListAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query config template version total count.
	if errCode, errMsg := act.queryTemplateVersionCount(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query config template version list.
	if errCode, errMsg := act.queryTemplateVersionList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
