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

package commit

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/integrator"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CommitAction handles commit confirm logic action.
type CommitAction struct {
	viper          *viper.Viper
	businessSvrCli pbbusinessserver.BusinessClient
	md             *structs.IntegrationMetadata

	req  *pb.IntegrateReq
	resp *pb.IntegrateResp

	business *pbcommon.Business

	appid    string
	cfgsetid string
	commitid string

	isCommitConfirmed bool
}

// NewCommitAction creates new CommitAction.
func NewCommitAction(viper *viper.Viper, businessSvrCli pbbusinessserver.BusinessClient, md *structs.IntegrationMetadata,
	req *pb.IntegrateReq, resp *pb.IntegrateResp) *CommitAction {
	action := &CommitAction{viper: viper, businessSvrCli: businessSvrCli, md: md, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CommitAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CommitAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_ITG_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CommitAction) Output() error {
	// do nothing.
	return nil
}

func (act *CommitAction) verify() error {
	length := len(act.md.Spec.BusinessName)
	if length == 0 {
		return errors.New("invalid params, businessName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, businessName too long")
	}

	length = len(act.md.Spec.AppName)
	if length == 0 {
		return errors.New("invalid params, appName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, appName too long")
	}

	length = len(act.md.Spec.ConfigSetName)
	if length == 0 {
		return errors.New("invalid params, configSetName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, configSetName too long")
	}

	act.md.Spec.ConfigSetFpath = common.ParseFpath(act.md.Spec.ConfigSetFpath)
	if len(act.md.Spec.ConfigSetFpath) > database.BSCPCFGSETFPATHLENLIMIT {
		return errors.New("invalid params, fpath too long")
	}

	if len(act.md.Configs) > database.BSCPCONFIGSSIZELIMIT {
		return errors.New("invalid params, configs content too big")
	}
	if len(act.md.Changes) > database.BSCPCHANGESSIZELIMIT {
		return errors.New("invalid params, configs changes too big")
	}

	if len(act.md.Template.Templateid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, templateid too long")
	}
	if len(act.md.Template.Template) > database.BSCPTPLSIZELIMIT {
		return errors.New("invalid params, template size too big")
	}
	if len(act.md.Template.TemplateRule) > database.BSCPTPLRULESSIZELIMIT {
		return errors.New("invalid params, template rules too long")
	}

	if len(act.md.Configs) != 0 && len(act.md.Template.Template) != 0 {
		return errors.New("invalid params, configs and template concurrence")
	}
	if len(act.md.Configs) != 0 && len(act.md.Template.Templateid) != 0 {
		return errors.New("invalid params, configs and templateid concurrence")
	}
	if len(act.md.Template.Template) != 0 && len(act.md.Template.Templateid) != 0 {
		return errors.New("invalid params, template and templateid concurrence")
	}
	if len(act.md.Template.Template) != 0 && len(act.md.Template.TemplateRule) == 0 {
		return errors.New("invalid params, empty template rules")
	}

	if len(act.md.Spec.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CommitAction) queryBusiness() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryBusinessReq{
		Seq:  act.req.Seq,
		Name: act.md.Spec.BusinessName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Commit[%d]| request to businessserver QueryBusiness, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryBusiness, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	act.resp.Bid = resp.Business.Bid
	act.business = resp.Business

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CommitAction) createApp() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateAppReq{
		Seq:        act.req.Seq,
		Bid:        act.business.Bid,
		Name:       act.md.Spec.AppName,
		DeployType: 0,
		Creator:    act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Commit[%d]| request to businessserver CreateApp, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateApp, %+v", err)
	}
	act.appid = resp.Appid
	act.resp.Appid = resp.Appid

	if resp.ErrCode == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *CommitAction) createConfigSet() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateConfigSetReq{
		Seq:     act.req.Seq,
		Bid:     act.business.Bid,
		Appid:   act.appid,
		Name:    act.md.Spec.ConfigSetName,
		Fpath:   act.md.Spec.ConfigSetFpath,
		Memo:    act.md.Spec.Memo,
		Creator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Commit[%d]| request to businessserver CreateConfigSet, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateConfigSet(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateConfigSet, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK && resp.ErrCode != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return resp.ErrCode, resp.ErrMsg
	}
	act.resp.Cfgsetid = resp.Cfgsetid
	act.cfgsetid = resp.Cfgsetid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CommitAction) createCommit() (pbcommon.ErrCode, string) {
	configs, err := base64.StdEncoding.DecodeString(act.md.Configs)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_PARAMS_INVALID, fmt.Sprintf("can't decode configs content from metadata, %+v", err)
	}

	r := &pbbusinessserver.CreateCommitReq{
		Seq:          act.req.Seq,
		Bid:          act.business.Bid,
		Appid:        act.appid,
		Cfgsetid:     act.cfgsetid,
		Operator:     act.req.Operator,
		Templateid:   act.md.Template.Templateid,
		Template:     act.md.Template.Template,
		TemplateRule: act.md.Template.TemplateRule,
		Configs:      configs,
		Changes:      act.md.Changes,
		Memo:         act.md.Spec.Memo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Commit[%d]| request to businessserver CreateCommit, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateCommit, %+v", err)
	}
	act.resp.Commitid = resp.Commitid
	act.commitid = resp.Commitid

	return resp.ErrCode, resp.ErrMsg
}

func (act *CommitAction) confirmCommit() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.ConfirmCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.business.Bid,
		Commitid: act.commitid,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Commit[%d]| request to businessserver ConfirmCommit, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.ConfirmCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver ConfirmCommit, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	act.isCommitConfirmed = true

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CommitAction) cancelCommit() error {
	// already confirmed, do not try to cancel it.
	if act.isCommitConfirmed {
		return nil
	}

	// created one commit with init state, would try to cancel it.
	r := &pbbusinessserver.CancelCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.business.Bid,
		Commitid: act.commitid,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Commit[%d]| request to businessserver CancelCommit, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CancelCommit(ctx, r)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// Do makes the workflows of this action base on input messages.
func (act *CommitAction) Do() error {
	// query busienss information used for publishing.
	if errCode, errMsg := act.queryBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create/query app information used for publishing.
	if errCode, errMsg := act.createApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create configset if no-exist.
	if errCode, errMsg := act.createConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create new commit everytime.
	if errCode, errMsg := act.createCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// cancel unconfirmed commit.
	defer act.cancelCommit()

	// confirm the commit, it would render the template if have one.
	if errCode, errMsg := act.confirmCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
