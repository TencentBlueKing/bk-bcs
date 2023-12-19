/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

// genSceneData5 在biz_id=2001，app_id=100005的应用下，创建5个配置项，执行一次兜底策略发布、四次Normal策略发布、
// 且第一条Normal策略包含子策略、200次实例发布。
func genSceneData5() error {
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

	// // gen five config item for every app, and create one content and commit for every config item.
	for i := 0; i < 5; i++ {
		if err := genCIRelatedData(stressBizId, appResp.Id); err != nil {
			return err
		}
	}

	// create release.
	rlReq := &pbcs.CreateReleaseReq{
		BizId: stressBizId,
		AppId: appResp.Id,
		Name:  randName("release"),
		Memo:  memo,
	}
	rid = RequestID()
	rlResp, err := cli.Release.Create(context.Background(), Header(rid), rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v, rid: %s", err, rid)
	}

	// create strategy set.
	setReq := &pbcs.CreateStrategySetReq{
		BizId: stressBizId,
		AppId: appResp.Id,
		Name:  randName("strategy_set"),
		Memo:  memo,
	}
	rid = RequestID()
	setResp, err := cli.StrategySet.Create(context.Background(), Header(rid), setReq)
	if err != nil {
		return fmt.Errorf("create strategy set err, %v, rid: %s", err, rid)
	}

	opt := &pubOption{
		bizID:         stressBizId,
		appID:         appResp.Id,
		strategySetID: setResp.Id,
		releaseID:     rlResp.Id,
	}
	if err := genScene5PublishData(opt); err != nil {
		return err
	}

	return nil
}

// pubOption gen publish data option.
type pubOption struct {
	bizID         uint32
	appID         uint32
	strategySetID uint32
	releaseID     uint32
}

func genScene5PublishData(opt *pubOption) error {
	// create default strategy.
	styReq := &pbcs.CreateStrategyReq{
		BizId:         opt.bizID,
		AppId:         opt.appID,
		StrategySetId: opt.strategySetID,
		Name:          randName("strategy"),
		AsDefault:     true,
		Memo:          memo,
		ReleaseId:     opt.releaseID,
	}
	rid := RequestID()
	_, err := cli.Strategy.Create(context.Background(), Header(rid), styReq)
	if err != nil {
		return fmt.Errorf("create strategy err, %v, rid: %s", err, rid)
	}

	// publish strategy.
	pbReq := &pbcs.PublishReq{
		BizId: opt.bizID,
		AppId: opt.appID,
	}
	rid = RequestID()
	_, err = cli.Publish.PublishWithStrategy(context.Background(), Header(rid), pbReq)
	if err != nil {
		return fmt.Errorf("create strategy publish err, %v, rid: %s", err, rid)
	}

	// generate four different scope for testing
	for i := 0; i < 4; i++ {
		sl := make([]selector.Element, 0)
		for j := 0; j <= i; j++ {
			sl = append(sl, elements[j])
		}

		styReq := &pbcs.CreateStrategyReq{
			BizId:         opt.bizID,
			AppId:         opt.appID,
			StrategySetId: opt.strategySetID,
			Name:          randName("strategy"),
			Memo:          memo,
			ReleaseId:     opt.releaseID,
		}

		rid = RequestID()
		_, err = cli.Strategy.Create(context.Background(), Header(rid), styReq)
		if err != nil {
			return fmt.Errorf("create strategy err, %v, rid: %s", err, rid)
		}

		// publish strategy.
		pbReq := &pbcs.PublishReq{
			BizId: opt.bizID,
			AppId: opt.appID,
		}
		rid = RequestID()
		_, err = cli.Publish.PublishWithStrategy(context.Background(), Header(rid), pbReq)
		if err != nil {
			return fmt.Errorf("create strategy publish err, %v, rid: %s", err, rid)
		}
	}

	return nil
}
