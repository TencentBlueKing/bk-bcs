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

package templatebinding

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// CreateAction action for creating templateset
type CreateAction struct {
	viper *viper.Viper

	smgr *dbsharding.ShardingManager
	sd   *dbsharding.ShardingDB

	req  *pb.CreateConfigTemplateBindingReq
	resp *pb.CreateConfigTemplateBindingResp
}

// NewCreateAction create new CreateAction
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateConfigTemplateBindingReq, resp *pb.CreateConfigTemplateBindingResp) *CreateAction {
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

	if err := common.VerifyID(act.req.Appid, "appid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Cfgsetid, "cfgsetid"); err != nil {
		return err
	}

	// when create, commitid can be empty
	if len(act.req.Commitid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}

	if err := common.VerifyTemplateBindingParams(act.req.BindingParams); err != nil {
		return nil
	}

	if err := common.VerifyNormalName(act.req.Creator, "creator"); err != nil {
		return nil
	}

	return nil
}

func (act *CreateAction) createTemplateBinding() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplateBinding{})

	st := database.ConfigTemplateBinding{
		Bid:           act.req.Bid,
		Templateid:    act.req.Templateid,
		Versionid:     act.req.Versionid,
		Appid:         act.req.Appid,
		Cfgsetid:      act.req.Cfgsetid,
		Commitid:      act.req.Commitid,
		BindingParams: act.req.BindingParams,
		Creator:       act.req.Creator,
		LastModifyBy:  act.req.Creator,
		State:         act.req.State,
	}

	err := act.sd.DB().
		Where(database.ConfigTemplateBinding{
			Bid:        st.Bid,
			Templateid: st.Templateid,
			Appid:      st.Appid,
		}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action with input message
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// create config template binding
	if errCode, errMsg := act.createTemplateBinding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
