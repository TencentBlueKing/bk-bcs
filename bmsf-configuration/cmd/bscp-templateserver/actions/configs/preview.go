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

package configs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/renderengine"
)

// PreviewAction previews template rendering results.
type PreviewAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.PreviewRenderingReq
	resp *pb.PreviewRenderingResp

	pluginName      string
	templateBinding *pbcommon.ConfigTemplateBinding
	template        *pbcommon.ConfigTemplate
	versionid       string
	version         *pbcommon.TemplateVersion
	commit          *pbcommon.Commit
	rules           structs.RuleList
}

// NewPreviewAction creates new PreviewAction.
func NewPreviewAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.PreviewRenderingReq, resp *pb.PreviewRenderingResp) *PreviewAction {
	action := &PreviewAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PreviewAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PreviewAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PreviewAction) Output() error {
	// do nothing.
	return nil
}

func (act *PreviewAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Appid, "appid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Commitid, "commitid"); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.Operator, "operator"); err != nil {
		return err
	}

	return nil
}

func (act *PreviewAction) queryCommit() (pbcommon.ErrCode, string) {
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

func (act *PreviewAction) queryTemplateBinding() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateBindingReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.commit.Templateid,
		Appid:      act.req.Appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryConfigTemplateBinding[%d]| request to datamanger QueryConfigTemplateBinding, %+v", act.req.Seq, req)

	resp, err := act.dataMgrCli.QueryConfigTemplateBinding(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryConfigTemplateBinding, %+v", err)
	}
	if resp.ErrCode == pbcommon.ErrCode_E_OK {
		act.templateBinding = resp.ConfigTemplateBinding
		act.versionid = act.templateBinding.Versionid
	}

	return resp.ErrCode, resp.ErrMsg
}

func (act *PreviewAction) queryTemplate() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.commit.Templateid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryConfigTemplate[%d]| request to datamanager QueryConfigTemplate, %+v", act.req.Seq, req)

	resp, err := act.dataMgrCli.QueryConfigTemplate(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanger QueryConfigTemplate, %+v", err)
	}
	if resp.ErrCode == pbcommon.ErrCode_E_OK {
		act.template = resp.ConfigTemplate
		plugin, ok := renderengine.EngineTypeMap[act.template.EngineType]
		if !ok {
			return pbcommon.ErrCode_E_TPL_RENDER_CHECK_ERROR, fmt.Sprintf("invalid engine type")
		}
		act.pluginName = plugin
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PreviewAction) queryVersion() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryTemplateVersionReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Versionid: act.versionid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryTemplateVersion[%d]| request to datamanager QueryTemplateVersion, %+v", act.req.Seq, req)

	resp, err := act.dataMgrCli.QueryTemplateVersion(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanger QueryTemplateVersion, %+v", err)
	}
	if resp.ErrCode == pbcommon.ErrCode_E_OK {
		act.version = resp.TemplateVersion
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PreviewAction) unmarshalTplRules() (pbcommon.ErrCode, string) {

	// unmarshal json rules string.
	if err := json.Unmarshal([]byte(act.commit.TemplateRule), &act.rules); err != nil {
		return pbcommon.ErrCode_E_TPL_INVALID_TEMPLATE_RULE, fmt.Sprintf("can't parse template rules, %+v", err)
	}

	return pbcommon.ErrCode_E_OK, ""
}

// queryZone query target zone, and get the zoneid to create zone level configs.
func (act *PreviewAction) queryZone(name string) (string, error) {
	r := &pbdatamanager.QueryZoneReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Preview[%d]| request to datamanager QueryZone, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryZone(ctx, r)
	if err != nil {
		logger.Error("Preview[%d] request to datamanager QueryZone, %+v", act.req.Seq, err)
		return "", err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Zone.Zoneid, nil
}

// queryCluster query target cluster, and get clusterid to create cluster level configs.
func (act *PreviewAction) queryCluster(name, labels string) (string, error) {

	r := &pbdatamanager.QueryClusterReq{
		Seq:    act.req.Seq,
		Bid:    act.req.Bid,
		Appid:  act.req.Appid,
		Name:   name,
		Labels: labels,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Preview[%d]| request to datamanager QueryCluster, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCluster(ctx, r)
	if err != nil {
		logger.Error("Preview[%d] request to datamanager QueryCluster, %+v", act.req.Seq, err)
		return "", err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Cluster.Clusterid, nil
}

// innerTemplateExecute executes template, and gen final configs content.
func (act *PreviewAction) innerTemplateExecute(conf *renderengine.RenderInConf) ([]*renderengine.RenderOutInstance, error) {
	t, err := renderengine.NewEngine(act.viper.GetString("templateplugin.bindir"))
	if err != nil {
		return nil, fmt.Errorf("create render engine failed, err %s", err.Error())
	}
	err = t.FindPlugin(act.pluginName)
	if err != nil {
		return nil, fmt.Errorf("find plugin failed, err %s", err.Error())
	}

	out, err := t.Execute(conf, act.pluginName)
	if err != nil {
		return nil, fmt.Errorf("execute render failed, err %s", err.Error())
	}
	if out == nil {
		return nil, fmt.Errorf("execute render return nil")
	}
	if out.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, fmt.Errorf("execute render err, code %d msg %s", out.ErrCode, out.ErrMsg)
	}

	return out.Instances, nil
}

func (act *PreviewAction) listGlobalVars() (map[string]interface{}, error) {
	globalVars := make(map[string]interface{})

	req := &pbdatamanager.QueryVariableListReq{
		Bid:   act.req.Bid,
		Type:  int32(pbcommon.VariableType_VT_GLOBAL),
		Limit: database.BSCPQUERYLIMIT,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryVariableList[%d]| request to datamanager QueryVariableList, %+v", act.req.Seq, req)

	resp, err := act.dataMgrCli.QueryVariableList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list global variables failed, err %s", err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, errors.New(resp.ErrMsg)
	}
	for _, v := range resp.Vars {
		globalVars[v.Key] = v.Value
	}
	return globalVars, nil
}

func (act *PreviewAction) listClusterVars(cluster, clusterLabels string) (map[string]interface{}, error) {
	clusterVars := make(map[string]interface{})

	req := &pbdatamanager.QueryVariableListReq{
		Bid:           act.req.Bid,
		Cluster:       cluster,
		ClusterLabels: clusterLabels,
		Type:          int32(pbcommon.VariableType_VT_CLUSTER),
		Limit:         database.BSCPQUERYLIMIT,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryVariableList[%d]| request to datamanager QueryVariableList, %+v", act.req.Seq, req)

	resp, err := act.dataMgrCli.QueryVariableList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list cluster variables failed, err %s", err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, errors.New(resp.ErrMsg)
	}
	for _, v := range resp.Vars {
		clusterVars[v.Key] = v.Value
	}
	return clusterVars, nil
}

func (act *PreviewAction) listZoneVars(cluster, clusterLabels, zone string) (map[string]interface{}, error) {
	zoneVars := make(map[string]interface{})

	req := &pbdatamanager.QueryVariableListReq{
		Bid:           act.req.Bid,
		Cluster:       cluster,
		ClusterLabels: clusterLabels,
		Zone:          zone,
		Type:          int32(pbcommon.VariableType_VT_ZONE),
		Limit:         database.BSCPQUERYLIMIT,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryVariableList[%d]| request to datamanager QueryVariableList, %+v", act.req.Seq, req)

	resp, err := act.dataMgrCli.QueryVariableList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list zone variables failed, err %s", err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, errors.New(resp.ErrMsg)
	}
	for _, v := range resp.Vars {
		zoneVars[v.Key] = v.Value
	}
	return zoneVars, nil
}

func (act *PreviewAction) render() (pbcommon.ErrCode, string) {

	globalVars, err := act.listGlobalVars()
	if err != nil {
		return pbcommon.ErrCode_E_TPL_GET_VARS_FAILED, err.Error()
	}

	renderInConf := &renderengine.RenderInConf{
		Vars:     globalVars,
		Operator: act.req.Operator,
	}

	if len(act.commit.Templateid) == 0 {
		renderInConf.Template = act.commit.Template
	} else {
		renderInConf.Template = act.version.Content
	}

	clusterMap := make(map[string]string)
	zoneMap := make(map[string]string)

	for _, rule := range act.rules {
		clusterLabelsBytes, err := json.Marshal(rule.ClusterLabels)
		if err != nil {
			return pbcommon.ErrCode_E_TS_PARAMS_INVALID, err.Error()
		}

		clusterid, err := act.queryCluster(rule.Cluster, string(clusterLabelsBytes))
		if err != nil {
			return pbcommon.ErrCode_E_TPL_NO_CLUSTER_TO_RENDER, err.Error()
		}
		clusterMap[genKeyForCluster(rule.Cluster, rule.ClusterLabels)] = clusterid

		clusterVars, err := act.listClusterVars(rule.Cluster, string(clusterLabelsBytes))
		if err != nil {
			return pbcommon.ErrCode_E_TPL_GET_VARS_FAILED, err.Error()
		}
		clusterConf := &renderengine.RenderInCluster{
			Cluster:       rule.Cluster,
			ClusterLabels: rule.ClusterLabels,
			Vars:          clusterVars,
		}
		for _, zone := range rule.Zones {
			zoneid, err := act.queryZone(zone.Zone)
			if err != nil {
				return pbcommon.ErrCode_E_TPL_NO_ZONE_TO_RENDER, err.Error()
			}
			zoneMap[zone.Zone] = zoneid

			zoneVars, err := act.listZoneVars(rule.Cluster, string(clusterLabelsBytes), zone.Zone)
			if err != nil {
				return pbcommon.ErrCode_E_TPL_GET_VARS_FAILED, err.Error()
			}
			zoneConf := &renderengine.RenderInZone{
				Zone: zone.Zone,
				Vars: zoneVars,
			}

			for _, ins := range zone.Instances {
				insConf := &renderengine.RenderInInstance{
					Index: ins.Index,
					Vars:  ins.Variables,
				}
				zoneConf.Instances = append(zoneConf.Instances, insConf)
			}

			clusterConf.Zones = append(clusterConf.Zones, zoneConf)
		}
		renderInConf.Clusters = append(renderInConf.Clusters, clusterConf)
	}

	renderOutInstances, err := act.innerTemplateExecute(renderInConf)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_RENDER_FAILED, err.Error()
	}

	for _, ins := range renderOutInstances {
		act.resp.Cfgslist = append(act.resp.Cfgslist, &pbcommon.Configs{
			Cfgsetid:  act.templateBinding.Cfgsetid,
			Appid:     act.templateBinding.Appid,
			Bid:       act.req.Bid,
			Clusterid: clusterMap[genKeyForCluster(ins.Cluster, ins.ClusterLabels)],
			Zoneid:    zoneMap[ins.Zone],
			Commitid:  act.templateBinding.Commitid,
			Cid:       common.SHA256(ins.Content),
			Content:   []byte(ins.Content),
			Creator:   act.req.Operator,
			Index:     ins.Index,
		})
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PreviewAction) Do() error {
	// query target commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if len(act.commit.Templateid) != 0 {
		// query target template
		if errCode, errMsg := act.queryTemplate(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
		// query target template binding
		if errCode, errMsg := act.queryTemplateBinding(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// query target template version
		if errCode, errMsg := act.queryVersion(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
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
