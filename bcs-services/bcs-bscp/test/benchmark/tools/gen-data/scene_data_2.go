/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	pbcs "bscp.io/pkg/protocol/config-server"
)

// genSceneData2 在biz_id=2001，app_id=100002的应用下，创建5个配置项，执行一次兜底策略发布。
func genSceneData2() error {
	appReq := &pbcs.CreateAppReq{
		BizId:          stressBizId,
		Name:           randName("app"),
		ConfigType:     string(table.File),
		Mode:           string(table.Normal),
		Memo:           memo,
		ReloadType:     string(table.ReloadWithFile),
		ReloadFilePath: "/tmp/reload.json",
	}
	rid := RequestID()
	appResp, err := cli.App.Create(context.Background(), Header(rid), appReq)
	if err != nil {
		return fmt.Errorf("create app err, %v, rid: %s", err, rid)
	}
	if appResp.Code != errf.OK {
		return fmt.Errorf("create app failed, code: %d, msg: %s, rid: %s", appResp.Code, appResp.Message, rid)
	}

	// gen five config item for every app, and create one content and commit for every config item.
	for i := 0; i < 5; i++ {
		if err := genCIRelatedData(stressBizId, appResp.Data.Id); err != nil {
			return err
		}
	}

	// create release.
	rlReq := &pbcs.CreateReleaseReq{
		BizId: stressBizId,
		AppId: appResp.Data.Id,
		Name:  randName("release"),
		Memo:  memo,
	}
	rid = RequestID()
	rlResp, err := cli.Release.Create(context.Background(), Header(rid), rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v, rid: %s", err, rid)
	}
	if rlResp.Code != errf.OK {
		return fmt.Errorf("create release failed, code: %d, msg: %s, rid: %s", rlResp.Code, rlResp.Message, rid)
	}

	// create strategy set.
	setReq := &pbcs.CreateStrategySetReq{
		BizId: stressBizId,
		AppId: appResp.Data.Id,
		Name:  randName("strategy_set"),
		Memo:  memo,
	}
	rid = RequestID()
	setResp, err := cli.StrategySet.Create(context.Background(), Header(rid), setReq)
	if err != nil {
		return fmt.Errorf("create strategy set err, %v, rid: %s", err, rid)
	}
	if setResp.Code != errf.OK {
		return fmt.Errorf("create strategy set failed, code: %d, msg: %s, rid: %s", setResp.Code,
			setResp.Message, rid)
	}

	// create strategy.
	styReq := &pbcs.CreateStrategyReq{
		BizId:         stressBizId,
		AppId:         appResp.Data.Id,
		StrategySetId: setResp.Data.Id,
		Name:          randName("strategy"),
		AsDefault:     true,
		Memo:          memo,
		ReleaseId:     rlResp.Data.Id,
	}
	rid = RequestID()
	styResp, err := cli.Strategy.Create(context.Background(), Header(rid), styReq)
	if err != nil {
		return fmt.Errorf("create strategy err, %v, rid: %s", err, rid)
	}
	if styResp.Code != errf.OK {
		return fmt.Errorf("create strategy failed, code: %d, msg: %s, rid: %s", styResp.Code, styResp.Message, rid)
	}

	// publish strategy.
	pbReq := &pbcs.PublishReq{
		BizId: stressBizId,
		AppId: appResp.Data.Id,
	}
	rid = RequestID()
	pbResp, err := cli.Publish.PublishWithStrategy(context.Background(), Header(rid), pbReq)
	if err != nil {
		return fmt.Errorf("create strategy publish err, %v, rid: %s", err, rid)
	}
	if pbResp.Code != errf.OK {
		return fmt.Errorf("create strategy publish failed, code: %d, msg: %s, rid: %s", pbResp.Code,
			pbResp.Message, rid)
	}

	return nil
}
