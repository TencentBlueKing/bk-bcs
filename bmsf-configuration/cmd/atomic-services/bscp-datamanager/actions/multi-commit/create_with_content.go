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

package multicommit

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
)

// CreateWithContentAction is multi commit create action object.
type CreateWithContentAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateMultiCommitWithContentReq
	resp *pb.CreateMultiCommitWithContentResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB
}

// NewCreateWithContentAction creates new CreateWithContentAction.
func NewCreateWithContentAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateMultiCommitWithContentReq, resp *pb.CreateMultiCommitWithContentResp) *CreateWithContentAction {
	action := &CreateWithContentAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateWithContentAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateWithContentAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateWithContentAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateWithContentAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_commit_id", act.req.MultiCommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.Metadatas) == 0 {
		return errors.New("invalid input data, empty metadatas")
	}

	for _, metadata := range act.req.Metadatas {
		if err = common.ValidateString("cfg_id", metadata.CfgId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if len(metadata.Contents) == 0 {
			return errors.New("invalid input data, empty contents")
		}

		for _, content := range metadata.Contents {
			content.ContentId = strings.ToUpper(content.ContentId)
			if err = common.ValidateString("content_id", content.ContentId,
				database.BSCPCONTENTIDLENLIMIT, database.BSCPCONTENTIDLENLIMIT); err != nil {
				return err
			}
		}
		act.sortContentsOrder(metadata)
	}

	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

// make the empty or/and labels content to the last, and the instance
// would match self content first and match the empty lables content as default.
func (act *CreateWithContentAction) sortContentsOrder(metadata *pbcommon.CommitMetadataWithContent) {
	finalContents := []*pbcommon.CommitContent{}
	emptyLabelsContents := []*pbcommon.CommitContent{}

	for _, content := range metadata.Contents {
		if len(content.LabelsOr) == 0 && len(content.LabelsAnd) == 0 {
			emptyLabelsContents = append(emptyLabelsContents, content)
			continue
		}

		isLabelsOrEmpty := true
		for _, labels := range content.LabelsOr {
			if len(labels.Labels) != 0 {
				isLabelsOrEmpty = false
				break
			}
		}

		isLabelsAndEmpty := true
		for _, labels := range content.LabelsAnd {
			if len(labels.Labels) != 0 {
				isLabelsAndEmpty = false
				break
			}
		}

		if isLabelsOrEmpty && isLabelsAndEmpty {
			emptyLabelsContents = append(emptyLabelsContents, content)
			continue
		}

		finalContents = append(finalContents, content)
	}

	finalContents = append(finalContents, emptyLabelsContents...)
	metadata.Contents = finalContents
}

func (act *CreateWithContentAction) createCommitMultiMode(cfgID, commitID string,
	contents []*pbcommon.CommitContent) (pbcommon.ErrCode, string) {

	commitMode := pbcommon.CommitMode_CM_CONFIGS
	if len(contents) != 0 {
		commitMode = pbcommon.CommitMode_CM_TEMPLATE
	}

	commit := database.Commit{
		BizID:         act.req.BizId,
		AppID:         act.req.AppId,
		CommitID:      commitID,
		CfgID:         cfgID,
		CommitMode:    int32(commitMode),
		Operator:      act.req.Operator,
		Memo:          act.req.Memo,
		State:         int32(pbcommon.CommitState_CS_INIT),
		MultiCommitID: act.req.MultiCommitId,
	}

	err := act.tx.
		Where(database.Commit{
			BizID:         act.req.BizId,
			AppID:         act.req.AppId,
			CfgID:         cfgID,
			MultiCommitID: act.req.MultiCommitId,
		}).
		Assign(commit).
		FirstOrCreate(&commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateWithContentAction) createCommits() (pbcommon.ErrCode, string) {
	for _, metadata := range act.req.Metadatas {
		// create commit.
		errCode, errMsg := act.createCommitMultiMode(metadata.CfgId, metadata.CommitId, metadata.Contents)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		// create commit contents.
		for _, content := range metadata.Contents {
			errCode, errMsg := act.createCommitContent(metadata.CfgId, metadata.CommitId, content)
			if errCode != pbcommon.ErrCode_E_OK {
				return errCode, errMsg
			}
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateWithContentAction) createCommitContent(cfgID, commitID string,
	content *pbcommon.CommitContent) (pbcommon.ErrCode, string) {

	labelsOr := []map[string]string{}
	for _, labels := range content.LabelsOr {
		if len(labels.Labels) != 0 {
			labelsOr = append(labelsOr, labels.Labels)
		}
	}

	labelsAnd := []map[string]string{}
	for _, labels := range content.LabelsAnd {
		if len(labels.Labels) != 0 {
			labelsAnd = append(labelsAnd, labels.Labels)
		}
	}

	contentIndex := &strategy.ContentIndex{LabelsOr: labelsOr, LabelsAnd: labelsAnd}
	index, err := json.Marshal(contentIndex)
	if err != nil {
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error()
	}

	st := database.Content{
		BizID:        act.req.BizId,
		AppID:        act.req.AppId,
		CfgID:        cfgID,
		CommitID:     commitID,
		ContentID:    content.ContentId,
		ContentSize:  uint64(content.ContentSize),
		Index:        string(index),
		State:        int32(pbcommon.CommonState_CS_VALID),
		Creator:      act.req.Operator,
		Memo:         act.req.Memo,
		LastModifyBy: act.req.Operator,
	}

	err = act.tx.
		Create(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateWithContentAction) createMultiCommit() (pbcommon.ErrCode, string) {
	commit := &database.MultiCommit{
		BizID:         act.req.BizId,
		AppID:         act.req.AppId,
		MultiCommitID: act.req.MultiCommitId,
		Operator:      act.req.Operator,
		Memo:          act.req.Memo,
		State:         int32(pbcommon.CommitState_CS_INIT),
	}

	err := act.tx.
		Create(commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateMultiCommitWithContentResp_RespData{MultiCommitId: act.req.MultiCommitId}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateWithContentAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// create multi commit.
	if errCode, errMsg := act.createMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// create commits.
	if errCode, errMsg := act.createCommits(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// commit tx.
	if err := act.tx.Commit().Error; err != nil {
		act.tx.Rollback()
		return act.Err(pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error())
	}

	return nil
}
