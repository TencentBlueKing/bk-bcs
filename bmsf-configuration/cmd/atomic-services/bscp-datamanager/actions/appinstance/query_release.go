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

package appinstance

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
)

// QueryReleaseAction is appinstance release query action object.
type QueryReleaseAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector *metrics.Collector

	req  *pb.QueryAppInstanceReleaseReq
	resp *pb.QueryAppInstanceReleaseResp

	sd *dbsharding.ShardingDB

	release            database.Release
	appInstance        database.AppInstance
	appInstanceRelease database.AppInstanceRelease

	multiRelease database.MultiRelease

	content *pbcommon.Content
}

// NewQueryReleaseAction creates new QueryReleaseAction.
func NewQueryReleaseAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	collector *metrics.Collector,
	req *pb.QueryAppInstanceReleaseReq, resp *pb.QueryAppInstanceReleaseResp) *QueryReleaseAction {

	action := &QueryReleaseAction{ctx: ctx, viper: viper, smgr: smgr, collector: collector, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryReleaseAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryReleaseAction) Output() error {
	act.resp.Data = &pb.QueryAppInstanceReleaseResp_RespData{
		ReleaseId:      act.release.ReleaseID,
		CommitId:       act.release.CommitID,
		ContentId:      act.content.ContentId,
		ContentSize:    act.content.ContentSize,
		MultiReleaseId: act.release.MultiReleaseID,
		MultiCommitId:  act.multiRelease.MultiCommitID,
		ReleaseName:    act.release.Name,
		Memo:           act.release.Memo,
	}
	return nil
}

func (act *QueryReleaseAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cloud_id", act.req.CloudId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("ip", act.req.Ip,
		database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
		return err
	}
	act.req.Path = filepath.Clean(act.req.Path)
	if err = common.ValidateString("path", act.req.Path,
		database.BSCPNOTEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *QueryReleaseAction) queryAppInstance() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Where(&database.AppInstance{
			BizID:   act.req.BizId,
			AppID:   act.req.AppId,
			CloudID: act.req.CloudId,
			IP:      act.req.Ip,
			Path:    act.req.Path,
		}).
		Last(&act.appInstance).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "appinstance non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryAppInstanceRelease() (pbcommon.ErrCode, string) {
	var sts []database.AppInstanceRelease

	err := act.sd.DB().
		Limit(1).
		Order("Feffect_time DESC, Fid DESC").
		Where(&database.AppInstanceRelease{InstanceID: act.appInstance.ID, CfgID: act.req.CfgId}).
		Find(&sts).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	if len(sts) == 0 {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "app instance release non-exist."
	}
	act.appInstanceRelease = sts[0]

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryRelease() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Where(&database.Release{BizID: act.req.BizId, ReleaseID: act.appInstanceRelease.ReleaseID}).
		Last(&act.release).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Where(&database.MultiRelease{BizID: act.req.BizId, MultiReleaseID: act.release.MultiReleaseID}).
		Last(&act.multiRelease).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "multi release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryConfigContents(start, limit int) ([]database.Content, pbcommon.ErrCode, string) {
	contentList := []database.Content{}

	err := act.sd.DB().
		Offset(start).Limit(limit).
		Order("Fid DESC").
		Where(&database.Content{
			BizID:    act.req.BizId,
			AppID:    act.req.AppId,
			CfgID:    act.req.CfgId,
			CommitID: act.release.CommitID,
		}).
		Find(&contentList).Error

	if err != nil {
		return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	return contentList, pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) matchConfigContent() (pbcommon.ErrCode, string) {
	instance := &pbcommon.AppInstance{
		AppId:   act.req.AppId,
		CloudId: act.req.CloudId,
		Ip:      act.req.Ip,
		Path:    act.req.Path,
		Labels:  act.appInstance.Labels,
	}

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

			if contentIndex.MatchInstance(instance) {
				content := &pbcommon.Content{
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
				act.content = content
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
		return pbcommon.ErrCode_E_DM_RELEASE_CONTENT_NOT_FOUND,
			"can't find any config content on the config for this app instance"
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryReleaseAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query appinstance.
	if errCode, errMsg := act.queryAppInstance(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query appinstance release.
	if errCode, errMsg := act.queryAppInstanceRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query multi release.
	if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// index target content.
	if errCode, errMsg := act.matchConfigContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
