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
	"bk-bscp/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/common"
)

// CreateAction create a template binding object
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateConfigTemplateBindingReq
	resp *pb.CreateConfigTemplateBindingResp

	templateRules structs.RuleList

	template    *pbcommon.ConfigTemplate
	version     *pbcommon.TemplateVersion
	app         *pbcommon.App
	newCfgSetid string
	newCommitid string
}

// NewCreateAction creates new CreateAction
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateConfigTemplateBindingReq, resp *pb.CreateConfigTemplateBindingResp) *CreateAction {
	action := &CreateAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {

	return nil
}

func (act *CreateAction) unmarshalBindingParams() error {
	var rules structs.RuleList
	if err := json.Unmarshal([]byte(act.req.BindingParams), &rules); err != nil {
		return fmt.Errorf("invalid param, decode bind parameter failed, err %s", err.Error())
	}
	act.templateRules = rules
	return nil
}

func (act *CreateAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Templateid, "templateid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Appid, "appid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Versionid, "versionid"); err != nil {
		return err
	}

	if len(act.req.BindingParams) == 0 {
		return errors.New("invalid params, missing bindingParams")
	}

	if err := act.unmarshalBindingParams(); err != nil {
		return err
	}

	for _, rule := range act.templateRules {
		if err := common.VerifyNormalName(rule.Cluster, "cluster"); err != nil {
			return fmt.Errorf("in binding params, %s", err.Error())
		}
		for _, zone := range rule.Zones {
			if err := common.VerifyNormalName(zone.Zone, "zone"); err != nil {
				return fmt.Errorf("in binding params, %s", err.Error())
			}
			for _, instance := range zone.Instances {
				if err := common.VerifyNormalName(instance.Index, "index"); err != nil {
					return fmt.Errorf("in instance of binding params, %s", err.Error())
				}
				for key := range instance.Variables {
					if err := common.VerifyVarKey(key); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (act *CreateAction) queryTemplate() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.req.Templateid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryConfigTemplate(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_NO_TEMPLATE_TO_BIND, resp.ErrMsg
	}
	act.template = resp.ConfigTemplate
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryVersion() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryTemplateVersionReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Versionid: act.req.Versionid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryTemplateVersion(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_NO_VERSION_TO_BIND, resp.ErrMsg
	}
	act.version = resp.TemplateVersion
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryApp() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryAppReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryApp(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_NO_APP_TO_BIND, resp.ErrMsg
	}
	act.app = resp.App
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *CreateAction) queryCluster(cluster string, clusterLabels map[string]string) (pbcommon.ErrCode, string) {

	clusterLabelsStr, err := json.Marshal(clusterLabels)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}

	req := &pbdatamanager.QueryClusterReq{
		Seq:    act.req.Seq,
		Bid:    act.req.Bid,
		Appid:  act.req.Appid,
		Name:   cluster,
		Labels: string(clusterLabelsStr),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryCluster(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_NO_CLUSTER_TO_BIND, fmt.Sprintf("can't query cluster[%v], %s", cluster, resp.ErrMsg)
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryZone(zone string) (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryZoneReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  zone,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryZone(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_NO_ZONE_TO_BIND, fmt.Sprintf("can't query zone[%v], %s", zone, resp.ErrMsg)
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) genConfigSetID() error {
	id, err := common.GenCfgsetid()
	if err != nil {
		return err
	}
	act.newCfgSetid = id
	return nil
}

func (act *CreateAction) createConfitSet() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.CreateConfigSetReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Appid:    act.req.Appid,
		Cfgsetid: act.newCfgSetid,
		Name:     act.template.Name,
		Creator:  act.req.Creator,
		Memo:     act.template.Memo,
		Fpath:    act.template.Fpath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.CreateConfigSet(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_CREATE_CFGSET_FAILED, fmt.Sprintf("can't create configset, %s", resp.ErrMsg)
	}

	act.newCfgSetid = resp.Cfgsetid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) deleteConfigSet() {
	req := &pbdatamanager.DeleteConfigSetReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Cfgsetid: act.newCfgSetid,
		Operator: act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.DeleteConfigSet(ctx, req)
	if err != nil {
		logger.Warnf("delete configSet[%v] failed, err %s", act.newCfgSetid, err.Error())
		return
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		logger.Warnf("delete configSet[%v] not ok, errCode %v, errMsg %v", act.newCfgSetid, resp.ErrCode, resp.ErrMsg)
		return
	}
}

func (act *CreateAction) genCommitID() error {
	id, err := common.GenCommitid()
	if err != nil {
		return err
	}
	act.newCommitid = id
	return nil
}

func (act *CreateAction) createCommit() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.CreateCommitReq{
		Seq:          act.req.Seq,
		Bid:          act.req.Bid,
		Commitid:     act.newCommitid,
		Appid:        act.req.Appid,
		Cfgsetid:     act.newCfgSetid,
		Op:           0,
		Operator:     act.req.Creator,
		Templateid:   act.req.Templateid,
		TemplateRule: act.req.BindingParams,
		Changes:      "",
		Memo:         act.version.Memo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.CreateCommit(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_CREATE_COMMIT_FAILED, fmt.Sprintf("can't create commit, %s", resp.ErrMsg)
	}

	act.newCommitid = resp.Commitid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) deleteCommit() {
	req := &pbdatamanager.CancelCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Commitid: act.newCommitid,
		Operator: act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.CancelCommit(ctx, req)
	if err != nil {
		logger.Warnf("cancel commit[%v] failed, err %s", act.newCfgSetid, err.Error())
		return
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		logger.Warnf("cancel commit[%v] not ok, errCode %v, errMsg %v", act.newCfgSetid, resp.ErrCode, resp.ErrMsg)
		return
	}
}

func (act *CreateAction) createTemplateBinding() (pbcommon.ErrCode, string) {

	bindingParamsBytes, err := json.Marshal(act.req.BindingParams)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("can't encoding binding params, %s", err.Error())
	}

	req := &pbdatamanager.CreateConfigTemplateBindingReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		Templateid:    act.req.Templateid,
		Appid:         act.req.Appid,
		Versionid:     act.req.Versionid,
		Cfgsetid:      act.newCfgSetid,
		Commitid:      act.newCommitid,
		BindingParams: string(bindingParamsBytes),
		Creator:       act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.CreateConfigTemplateBinding(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.resp.Cfgsetid = act.newCfgSetid
	act.resp.Commitid = act.newCommitid

	return pbcommon.ErrCode_E_OK, ""
}

// Do do action
func (act *CreateAction) Do() error {

	if errCode, errMsg := act.queryTemplate(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if errCode, errMsg := act.queryVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	for _, rule := range act.templateRules {
		if errCode, errMsg := act.queryCluster(rule.Cluster, rule.ClusterLabels); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
		for _, zone := range rule.Zones {
			if errCode, errMsg := act.queryZone(zone.Zone); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}
		}

	}

	if err := act.genConfigSetID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error())
	}

	if errCode, errMsg := act.createConfitSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if err := act.genCommitID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error())
	}

	if errCode, errMsg := act.createCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.deleteCommit()
		return act.Err(errCode, errMsg)
	}

	if errCode, errMsg := act.createTemplateBinding(); errCode != pbcommon.ErrCode_E_OK {
		act.deleteCommit()
		act.deleteConfigSet()
		return act.Err(errCode, errMsg)
	}

	return nil
}
