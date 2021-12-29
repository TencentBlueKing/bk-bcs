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

package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PullAction pull release (target or newest) for bcs sidecar.
type PullAction struct {
	viper            *viper.Viper
	gseControllerCli pbgsecontroller.GSEControllerClient
	dataMgrCli       pbdatamanager.DataManagerClient

	req  *pb.GTCMDPullReleaseReq
	resp *pb.GTCMDPullReleaseResp

	release *pbcommon.Release
}

// NewPullAction creates new PullAction.
func NewPullAction(viper *viper.Viper, gseControllerCli pbgsecontroller.GSEControllerClient,
	dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.GTCMDPullReleaseReq, resp *pb.GTCMDPullReleaseResp) *PullAction {

	action := &PullAction{viper: viper, gseControllerCli: gseControllerCli,
		dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PullAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PullAction) Input() error {
	if act.req.Labels == "" {
		act.req.Labels = strategy.EmptySidecarLabels
	}
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PullAction) Output() error {
	// do nothing.
	return nil
}

func (act *PullAction) verify() error {
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

func (act *PullAction) pullRelease() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PullReleaseReq{
		Seq:            act.req.Seq,
		BizId:          act.req.BizId,
		AppId:          act.req.AppId,
		CloudId:        act.req.CloudId,
		Ip:             act.req.Ip,
		Path:           act.req.Path,
		Labels:         act.req.Labels,
		CfgId:          act.req.CfgId,
		LocalReleaseId: act.req.LocalReleaseId,
		ReleaseId:      act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PullRelease[%s]| request to gse-controller, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.PullRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse-controller PullRelease, %+v", err)
	}
	logger.V(4).Infof("PullRelease[%s]| release checking, %+v", act.req.Seq, resp)
	act.release = resp.Release

	return resp.Code, resp.Message
}

func (act *PullAction) queryReleaseConfigContent() (pbcommon.ErrCode, string) {
	sidecarLabels := &strategy.SidecarLabels{}
	if err := json.Unmarshal([]byte(act.req.Labels), sidecarLabels); err != nil {
		return pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN, err.Error()
	}

	r := &pbdatamanager.QueryReleaseConfigContentReq{
		Seq:      act.req.Seq,
		BizId:    act.req.BizId,
		AppId:    act.req.AppId,
		CloudId:  act.req.CloudId,
		Ip:       act.req.Ip,
		Path:     act.req.Path,
		Labels:   sidecarLabels.Labels,
		CfgId:    act.req.CfgId,
		CommitId: act.release.CommitId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PullRelease[%s]| request to datamanager, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryReleaseConfigContent(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryReleaseConfigContent, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	act.resp.ContentId = resp.Data.ContentId
	act.resp.ContentSize = resp.Data.ContentSize

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PullAction) Do() error {
	if errCode, errMsg := act.pullRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release == nil {
		// target release is not need to effect or there
		// is no newest release.
		return nil
	}

	// query release configs.
	if errCode, errMsg := act.queryReleaseConfigContent(); errCode != pbcommon.ErrCode_E_OK {
		// NOTE: return release base info when query release config content failed.
		act.resp.Release = &pbcommon.Release{ReleaseId: act.release.ReleaseId}
		return act.Err(errCode, errMsg)
	}
	act.resp.NeedEffect = true
	act.resp.Release = act.release

	return nil
}
