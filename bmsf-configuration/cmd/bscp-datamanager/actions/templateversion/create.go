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

// CreateAction action to create config template
type CreateAction struct {
	viper *viper.Viper

	smgr *dbsharding.ShardingManager
	sd   *dbsharding.ShardingDB

	req  *pb.CreateTemplateVersionReq
	resp *pb.CreateTemplateVersionResp
}

// NewCreateAction create new CreateAction
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateTemplateVersionReq, resp *pb.CreateTemplateVersionResp) *CreateAction {
	action := &CreateAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return error
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handle input message
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handle output message
func (act *CreateAction) Output() error {
	// do nothing
	return nil
}

func (act *CreateAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Templateid, "templateid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Versionid, "versionid"); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.VersionName, "versionName"); err != nil {
		return err
	}

	if err := common.VerifyMemo(act.req.Memo); err != nil {
		return err
	}

	if err := common.VerifyTemplateContent(act.req.Content); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.Creator, "creator"); err != nil {
		return err
	}

	return nil
}

func (act *CreateAction) reCreateTemplateVersion() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"Versionid":    act.req.Versionid,
		"Memo":         act.req.Memo,
		"Content":      act.req.Content,
		"Creator":      act.req.Creator,
		"LastModifyBy": act.req.Creator,
		"State":        int32(pbcommon.ConfigTemplateVersionState_CTVS_CREATED),
	}

	exec := act.sd.DB().
		Model(&database.ConfigTemplateVersion{}).
		Where(database.ConfigTemplateVersion{
			Bid:         act.req.Bid,
			Templateid:  act.req.Templateid,
			VersionName: act.req.VersionName,
			State:       int32(pbcommon.ConfigTemplateVersionState_CTVS_DELETED),
		}).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate template version failed, no eligible template version"
	}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createTemplateVersion() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplateVersion{})

	st := database.ConfigTemplateVersion{
		Bid:          act.req.Bid,
		Templateid:   act.req.Templateid,
		Versionid:    act.req.Versionid,
		VersionName:  act.req.VersionName,
		Memo:         act.req.Memo,
		Content:      act.req.Content,
		Creator:      act.req.Creator,
		LastModifyBy: act.req.Creator,
		State:        int32(pbcommon.ConfigTemplateVersionState_CTVS_CREATED),
	}

	err := act.sd.DB().
		Where(database.ConfigTemplateVersion{
			Bid:         st.Bid,
			Templateid:  st.Templateid,
			VersionName: st.VersionName,
		}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Versionid = st.Versionid

	if st.Versionid != act.req.Versionid {
		if st.State == int32(pbcommon.ConfigTemplateVersionState_CTVS_CREATED) {
			return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "the template version with target templateid+versionName already exist."
		}
		if errCode, errMsg := act.reCreateTemplateVersion(); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// create config template version
	if errCode, errMsg := act.createTemplateVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
