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

package configs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// RenderAction renders configs base on template.
type RenderAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.RenderReq
	resp *pb.RenderResp

	commit   *pbcommon.Commit
	zones    structs.RuleList
	clusters structs.RuleList
}

// NewRenderAction creates new RenderAction.
func NewRenderAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.RenderReq, resp *pb.RenderResp) *RenderAction {
	action := &RenderAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *RenderAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RenderAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *RenderAction) Output() error {
	// do nothing.
	return nil
}

func (act *RenderAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}
	return nil
}

func (act *RenderAction) queryCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Commitid: act.req.Commitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryCommit[%d]| request to datamanager QueryCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	act.commit = resp.Commit

	return resp.ErrCode, resp.ErrMsg
}

func (act *RenderAction) renderPre() error {
	if len(act.commit.Template) == 0 {
		return errors.New("can't render configs, template of target commit is missing")
	}

	if len(act.commit.TemplateRule) == 0 {
		return errors.New("can't render configs, template rule of target commit is missing")
	}
	return nil
}

func (act *RenderAction) unmarshalTplRules() (pbcommon.ErrCode, string) {
	// all template rules.
	rules := structs.RuleList{}

	// unmarshal json rules string.
	if err := json.Unmarshal([]byte(act.commit.TemplateRule), &rules); err != nil {
		return pbcommon.ErrCode_E_TPL_INVALID_TEMPLATE_RULE, fmt.Sprintf("can't parse template rules, %+v", err)
	}

	// split cluster and zone rules.
	for _, rule := range rules {
		if rule.Type == structs.RuleKeyTypeCluster {
			// cluster template rule.
			act.clusters = append(act.clusters, rule)
		} else if rule.Type == structs.RuleKeyTypeZone {
			// zone template rule.
			act.zones = append(act.zones, rule)
		} else {
			return pbcommon.ErrCode_E_TPL_INVALID_TEMPLATE_RULE_TYPE, "can't render, invalid template rule type"
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

// queryZone query target zone, and get the clusterid and zoneid to create zone level configs.
func (act *RenderAction) queryZone(name string) (string, string, error) {
	r := &pbdatamanager.QueryZoneReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Render[%d]| request to datamanager QueryZone, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryZone(ctx, r)
	if err != nil {
		logger.Error("Render[%d] request to datamanager QueryZone, %+v", act.req.Seq, err)
		return "", "", err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", "", errors.New(resp.ErrMsg)
	}
	return resp.Zone.Clusterid, resp.Zone.Zoneid, nil
}

// queryCluster query target cluster, and get clusterid to create cluster level configs.
func (act *RenderAction) queryCluster(name string) (string, error) {
	r := &pbdatamanager.QueryClusterReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Render[%d]| request to datamanager QueryCluster, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCluster(ctx, r)
	if err != nil {
		logger.Error("Render[%d] request to datamanager QueryCluster, %+v", act.req.Seq, err)
		return "", err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Cluster.Clusterid, nil
}

// createConfigs creates cluster or zone level configs, when it's cluster level, the zoneid should be empty.
func (act *RenderAction) createConfigs(clusterid, zoneid string, configs []byte) error {
	r := &pbdatamanager.CreateConfigsReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: clusterid,
		Zoneid:    zoneid,
		Cfgsetid:  act.req.Cfgsetid,
		Commitid:  act.req.Commitid,
		Cid:       common.SHA256(string(configs)),
		CfgLink:   "",
		Content:   configs,
		Creator:   act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Render[%d]| request to datamanager CreateConfigs, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateConfigs(ctx, r)
	if err != nil {
		logger.Error("Render[%d] request to datamanager CreateConfigs, %+v", act.req.Seq, err)
		return err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// tplExecute executes template, and gen final configs content.
func (act *RenderAction) tplExecute(tpl string, vars map[string]interface{}) ([]byte, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return nil, err
	}

	// the final configs content size may over the limit, don't block it here,
	// it would be checked at datamanager level.
	buffer := bytes.NewBuffer(nil)

	// rendering template.
	if err := t.Execute(buffer, vars); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// renderForZone renders template for zone level.
func (act *RenderAction) renderForZone(zone structs.Rule) error {
	if zone.Type != structs.RuleKeyTypeZone {
		return errors.New("wrong template rule type")
	}

	// rendering template to gen configs.
	configs, err := act.tplExecute(act.commit.Template, zone.Variables)
	if err != nil {
		return err
	}

	// TODO ignore error if not found ?
	clusterid, zoneid, err := act.queryZone(zone.Name)
	if err != nil {
		return err
	}

	if len(clusterid) == 0 {
		return errors.New("can't find the clusterid under target app with cluster name")
	}

	if len(zoneid) == 0 {
		return errors.New("can't find the zoneid under target app with zone name")
	}

	// create zone level configs.
	if err := act.createConfigs(clusterid, zoneid, configs); err != nil {
		return err
	}
	return nil
}

// renderForCluster renders template for cluster level.
func (act *RenderAction) renderForCluster(cluster structs.Rule) error {
	if cluster.Type != structs.RuleKeyTypeCluster {
		return errors.New("wrong template rule type")
	}

	// rendering template to gen configs.
	configs, err := act.tplExecute(act.commit.Template, cluster.Variables)
	if err != nil {
		return err
	}

	// TODO ignore error if not found ?
	clusterid, err := act.queryCluster(cluster.Name)
	if err != nil {
		return err
	}

	if len(clusterid) == 0 {
		return errors.New("can't find the clusterid under target app with cluster name")
	}

	// create cluster level configs.
	if err := act.createConfigs(clusterid, "", configs); err != nil {
		return err
	}
	return nil
}

func (act *RenderAction) render() (pbcommon.ErrCode, string) {
	// all zone levels here.
	for _, zone := range act.zones {
		if err := act.renderForZone(zone); err != nil {
			return pbcommon.ErrCode_E_TPL_RENDER_FAILED, fmt.Sprintf("can't render for zone[%s-%+v], %+v", zone.Name, zone.Type, err)
		}
	}

	// all cluster levels here.
	for _, cluster := range act.clusters {
		if err := act.renderForCluster(cluster); err != nil {
			return pbcommon.ErrCode_E_TPL_RENDER_FAILED, fmt.Sprintf("can't render for cluster[%s-%+v], %+v", cluster.Name, cluster.Type, err)
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *RenderAction) Do() error {
	// query target commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// check base information of target commit for rendering.
	if err := act.renderPre(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_RENDER_CHECK_ERROR, err.Error())
	}

	// unmarshal template rules.
	if errCode, errMsg := act.unmarshalTplRules(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// render configs base on template.
	if errCode, errMsg := act.render(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
