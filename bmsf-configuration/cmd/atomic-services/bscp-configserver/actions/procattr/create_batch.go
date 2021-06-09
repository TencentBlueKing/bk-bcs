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

package procattr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/audit"
	"bk-bscp/internal/authorization"
	"bk-bscp/internal/database"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// CreateBatchAction creates a procattr object.
type CreateBatchAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateProcAttrBatchReq
	resp *pb.CreateProcAttrBatchResp

	app *pbcommon.App
}

// NewCreateBatchAction creates new CreateBatchAction.
func NewCreateBatchAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateProcAttrBatchReq, resp *pb.CreateProcAttrBatchResp) *CreateBatchAction {

	action := &CreateBatchAction{
		kit:        kit,
		viper:      viper,
		authSvrCli: authSvrCli,
		dataMgrCli: dataMgrCli,
		req:        req,
		resp:       resp,
	}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateBatchAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateBatchAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *CreateBatchAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *CreateBatchAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateBatchAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if err = common.ValidateInt("data", len(act.req.Data), 0,
		database.BSCPCREATEBATCHLIMIT); err != nil {
		return err
	}

	for _, data := range act.req.Data {
		if err = common.ValidateString("cloud_id", data.CloudId,
			database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("ip", data.Ip,
			database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
			return err
		}

		data.Path = filepath.Clean(data.Path)
		if err = common.ValidateString("path", data.Path,
			database.BSCPNOTEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
			return err
		}
		if data.Labels == nil {
			data.Labels = make(map[string]string)
		}
		if err = common.ValidateString("memo", data.Memo,
			database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
			return err
		}
	}
	return nil
}

func (act *CreateBatchAction) authorize() (pbcommon.ErrCode, string) {
	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// check resource authorization.
	isAuthorized, err := authorization.Authorize(act.kit, act.req.AppId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "not authorized"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateBatchAction) queryApp() (pbcommon.ErrCode, string) {
	if act.app != nil {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateProcAttrBatch[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.Data

	return resp.Code, resp.Message
}

func (act *CreateBatchAction) create(data *pb.CreateProcAttrBatchReq_ReqData) (pbcommon.ErrCode, string) {
	labels, err := json.Marshal(data.Labels)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error()
	}

	r := &pbdatamanager.CreateProcAttrReq{
		Seq:      act.kit.Rid,
		CloudId:  data.CloudId,
		Ip:       data.Ip,
		BizId:    act.req.BizId,
		AppId:    act.req.AppId,
		Path:     data.Path,
		Labels:   string(labels),
		Memo:     data.Memo,
		Creator:  act.kit.User,
		Override: true,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateProcAttrBatch[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateProcAttr(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CreateProcAttr, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CreateBatchAction) createInBatchMode() (pbcommon.ErrCode, string) {
	respData := &pb.CreateProcAttrBatchResp_RespData{
		Failed: []*pb.CreateProcAttrBatchResp_FailedInfo{},
	}

	for _, data := range act.req.Data {
		errCode, errMsg := act.create(data)
		if errCode != pbcommon.ErrCode_E_OK && errCode != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {

			failedInfo := &pb.CreateProcAttrBatchResp_FailedInfo{
				Info: &pb.CreateProcAttrBatchResp_ProcAttrInfo{
					CloudId: data.CloudId,
					Ip:      data.Ip,
					Labels:  data.Labels,
					Path:    data.Path,
					Memo:    data.Memo,
				},
				Code:    errCode,
				Message: errMsg,
			}

			respData.Failed = append(respData.Failed, failedInfo)
			continue
		}

		if errCode == pbcommon.ErrCode_E_OK {
			// audit here on new procattr created.
			audit.Audit(int32(pbcommon.SourceType_ST_PROC_ATTR), int32(pbcommon.SourceOpType_SOT_CREATE),
				act.req.BizId, act.req.AppId, act.kit.User, data.Memo)
		}
	}
	act.resp.Data = respData

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateBatchAction) Do() error {
	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create procattrs.
	if errCode, errMsg := act.createInBatchMode(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
