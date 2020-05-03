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
	"bk-bscp/pkg/logger"
)

// PreviewAction previews template rendering results.
type PreviewAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.PreviewRenderingReq
	resp *pb.PreviewRenderingResp

	rule *structs.Rule
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

	if len(act.req.Template) > database.BSCPTPLSIZELIMIT {
		return errors.New("invalid params, template size too big")
	}

	if len(act.req.TemplateRule) > database.BSCPTPLRULESSIZELIMIT {
		return errors.New("invalid params, template rules too long")
	}
	return nil
}

func (act *PreviewAction) unmarshalTplRule() error {
	rule := &structs.Rule{}
	if err := json.Unmarshal([]byte(act.req.TemplateRule), rule); err != nil {
		return err
	}
	act.rule = rule
	return nil
}

// queryZone query target zone.
func (act *PreviewAction) queryZone(name string) error {
	r := &pbdatamanager.QueryZoneReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PreviewRendering[%d]| request to datamanager QueryZone, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryZone(ctx, r)
	if err != nil {
		logger.Error("PreviewRendering[%d] request to datamanager QueryZone, %+v", act.req.Seq, err)
		return err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// queryCluster query target cluster.
func (act *PreviewAction) queryCluster(name string) error {
	r := &pbdatamanager.QueryClusterReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PreviewRendering[%d]| request to datamanager QueryCluster, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCluster(ctx, r)
	if err != nil {
		logger.Error("PreviewRendering[%d] request to datamanager QueryCluster, %+v", act.req.Seq, err)
		return err
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

func (act *PreviewAction) previewCheck() (pbcommon.ErrCode, string) {
	if act.rule.Type == structs.RuleKeyTypeCluster {
		if err := act.queryCluster(act.rule.Name); err != nil {
			return pbcommon.ErrCode_E_TPL_NO_CLUSTER_TO_RENDER, "can't preview rendering, cluster not found"
		}
	} else if act.rule.Type == structs.RuleKeyTypeZone {
		if err := act.queryZone(act.rule.Name); err != nil {
			return pbcommon.ErrCode_E_TPL_NO_ZONE_TO_RENDER, "can't preview rendering, zone not found"
		}
	} else {
		return pbcommon.ErrCode_E_TPL_INVALID_TEMPLATE_RULE_TYPE, "can't preview rendering, invalid template rule type"
	}
	return pbcommon.ErrCode_E_OK, ""
}

// tplExecute executes template, and gen final configs content.
func (act *PreviewAction) tplExecute(tpl string, vars map[string]interface{}) ([]byte, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	if err := t.Execute(buffer, vars); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (act *PreviewAction) preview() (pbcommon.ErrCode, string) {
	configs, err := act.tplExecute(act.req.Template, act.rule.Variables)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_RENDER_FAILED, fmt.Sprintf("can't preview rendering, render template failed, %+v", err)
	}
	act.resp.Content = configs

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PreviewAction) Do() error {
	if err := act.unmarshalTplRule(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_INVALID_TEMPLATE_RULE, fmt.Sprintf("can't parse template rule, %+v", err))
	}

	if errCode, errMsg := act.previewCheck(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if errCode, errMsg := act.preview(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
