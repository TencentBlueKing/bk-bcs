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

// UpdateAction update a template binding object
type UpdateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.SyncConfigTemplateBindingReq
	resp *pb.SyncConfigTemplateBindingResp

	templateRules structs.RuleList

	binding     *pbcommon.ConfigTemplateBinding
	cfgset      *pbcommon.ConfigSet
	version     *pbcommon.TemplateVersion
	app         *pbcommon.App
	newCommitid string
}

// NewUpdateAction creates new UpdateAction
func NewUpdateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.SyncConfigTemplateBindingReq, resp *pb.SyncConfigTemplateBindingResp) *UpdateAction {
	action := &UpdateAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *UpdateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *UpdateAction) Output() error {

	return nil
}

func (act *UpdateAction) unmarshalBindingParams() error {
	var rules structs.RuleList
	if err := json.Unmarshal([]byte(act.req.BindingParams), &rules); err != nil {
		return fmt.Errorf("invalid param, decode bind parameter failed, err %s", err.Error())
	}
	act.templateRules = rules
	return nil
}

func (act *UpdateAction) verify() error {
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

func (act *UpdateAction) queryBinding() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateBindingReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.req.Templateid,
		Appid:      act.req.Appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryConfigTemplateBinding(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.binding = resp.ConfigTemplateBinding

	return pbcommon.ErrCode_E_OK, ""
}

func (act *UpdateAction) queryVersion() (pbcommon.ErrCode, string) {
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

func (act *UpdateAction) queryConfigSet() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigSetReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Cfgsetid: act.binding.Cfgsetid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryConfigSet(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_TPL_NO_CFGSET_TO_SYNC, err.Error()
	}
	act.cfgset = resp.ConfigSet
	return pbcommon.ErrCode_E_OK, ""
}

func (act *UpdateAction) queryCluster(cluster string, clusterLabels map[string]string) (pbcommon.ErrCode, string) {
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

func (act *UpdateAction) queryZone(zone string) (pbcommon.ErrCode, string) {
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

func (act *UpdateAction) genCommitID() error {
	id, err := common.GenCommitid()
	if err != nil {
		return err
	}
	act.newCommitid = id
	return nil
}

func (act *UpdateAction) createCommit() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.CreateCommitReq{
		Seq:          act.req.Seq,
		Bid:          act.req.Bid,
		Appid:        act.req.Appid,
		Cfgsetid:     act.cfgset.Cfgsetid,
		Commitid:     act.newCommitid,
		Op:           0,
		Operator:     act.req.Operator,
		Templateid:   act.req.Templateid,
		Template:     "",
		TemplateRule: act.req.BindingParams,
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

func (act *UpdateAction) updateBinding() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.UpdateConfigTemplateBindingReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		Templateid:    act.req.Templateid,
		Appid:         act.req.Appid,
		Versionid:     act.req.Versionid,
		Commitid:      act.newCommitid,
		BindingParams: act.req.BindingParams,
		State:         0,
		Operator:      act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.UpdateConfigTemplateBinding(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.resp.Commitid = act.newCommitid

	return pbcommon.ErrCode_E_OK, ""
}

// Do do action
func (act *UpdateAction) Do() error {

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

	// query binding
	if errCode, errMsg := act.queryBinding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, fmt.Sprintf("query binding failed when sync binding, %s", errMsg))
	}

	// query config set
	if errCode, errMsg := act.queryConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, fmt.Sprintf("query configset failed when sync binding, %s", errMsg))
	}

	// query version
	if errCode, errMsg := act.queryVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, fmt.Sprintf("query template version failed when sync binding, %s", errMsg))
	}

	if err := act.genCommitID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error())
	}

	// create commit
	if errCode, errMsg := act.createCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, fmt.Sprintf("create commit failed when sync binding, %s", errMsg))
	}

	// update binding
	if errCode, errMsg := act.updateBinding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
