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

package effect

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// ReportAction report sidecar release information.
type ReportAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req *pb.GTCMDEffectReport
}

// NewReportAction creates new ReportAction.
func NewReportAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.GTCMDEffectReport) *ReportAction {
	action := &ReportAction{viper: viper, dataMgrCli: dataMgrCli, req: req}
	return action
}

// Err setup error code message in response and return the error.
func (act *ReportAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReportAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReportAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReportAction) verify() error {
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
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("ip", act.req.Ip,
		database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("labels", act.req.Labels,
		database.BSCPEMPTY, database.BSCPLABELSSIZELIMIT); err != nil {
		return err
	}
	if len(act.req.Labels) == 0 {
		act.req.Labels = strategy.EmptySidecarLabels
	}
	act.req.Path = filepath.Clean(act.req.Path)
	if err = common.ValidateString("path", act.req.Path,
		database.BSCPNOTEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateInt("infos", len(act.req.Infos),
		database.BSCPNOTEMPTY, math.MaxInt32); err != nil {
		return err
	}
	return nil
}

func (act *ReportAction) report() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateAppInstanceReleaseReq{
		Seq:     act.req.Seq,
		BizId:   act.req.BizId,
		AppId:   act.req.AppId,
		CloudId: act.req.CloudId,
		Ip:      act.req.Ip,
		Path:    act.req.Path,
		Labels:  act.req.Labels,
		Infos:   act.req.Infos,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Report[%s]| request to datamanager, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateAppInstanceRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to  datamanager CreateAppInstanceRelease, %+v", err)
	}
	return resp.Code, resp.Message
}

// Do makes the workflows of this action base on input messages.
func (act *ReportAction) Do() error {
	if errCode, errMsg := act.report(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
