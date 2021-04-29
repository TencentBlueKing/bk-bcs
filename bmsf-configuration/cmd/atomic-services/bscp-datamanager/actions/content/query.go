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
	"encoding/json"
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
)

// QueryAction is content query action object.
type QueryAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryConfigContentReq
	resp *pb.QueryConfigContentResp

	sd *dbsharding.ShardingDB

	content     *pbcommon.Content
	contentList []database.Content
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryConfigContentReq, resp *pb.QueryConfigContentResp) *QueryAction {
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
	act.resp.Data = act.content
	return nil
}

func (act *QueryAction) verify() error {
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
	if act.req.Labels == nil {
		act.req.Labels = make(map[string]string)
	}
	return nil
}

func (act *QueryAction) queryConfigContents(start, limit int) ([]database.Content,
	pbcommon.ErrCode, string) {

	contentList := []database.Content{}

	err := act.sd.DB().
		Offset(start).Limit(limit).
		Order("Fid DESC").
		Where(&database.Content{
			BizID:    act.req.BizId,
			AppID:    act.req.AppId,
			CfgID:    act.req.CfgId,
			CommitID: act.req.CommitId,
		}).
		Find(&contentList).Error

	if err != nil {
		return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	return contentList, pbcommon.ErrCode_E_OK, ""
}

func (act *QueryAction) matchConfigContent() (pbcommon.ErrCode, string) {
	index := 0
	limit := database.BSCPQUERYLIMITMB

	for {
		contents, errCode, errMsg := act.queryConfigContents(index, limit)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		for _, content := range contents {
			contentIndex := strategy.ContentIndex{}
			if err := json.Unmarshal([]byte(content.Index), &contentIndex); err != nil {
				return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error()
			}

			if contentIndex.MatchLabels(act.req.Labels) {
				target := &pbcommon.Content{
					BizId:        content.BizID,
					AppId:        content.AppID,
					CfgId:        content.CfgID,
					CommitId:     content.CommitID,
					ContentId:    content.ContentID,
					ContentSize:  uint32(content.ContentSize),
					Index:        content.Index,
					Creator:      content.Creator,
					LastModifyBy: content.LastModifyBy,
					Memo:         content.Memo,
					State:        content.State,
					CreatedAt:    content.CreatedAt.Format("2006-01-02 15:04:05"),
					UpdatedAt:    content.UpdatedAt.Format("2006-01-02 15:04:05"),
				}
				act.content = target
				return pbcommon.ErrCode_E_OK, ""
			}
		}

		// no more contents to match.
		if len(contents) < limit {
			break
		}

		// no enough contents.
		index += len(contents)
	}

	if act.content == nil {
		return pbcommon.ErrCode_E_DM_RELEASE_CONTENT_NOT_FOUND, "content not found"
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

	// index target content.
	if errCode, errMsg := act.matchConfigContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
