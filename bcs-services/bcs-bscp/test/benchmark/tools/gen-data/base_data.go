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
	"log"
	"sync"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/criteria/uuid"
	"bscp.io/pkg/dal/table"
	pbcs "bscp.io/pkg/protocol/config-server"
)

// genBaseData 在业务id为 1-2000 的业务下，生成 50 个应用，且都生成 5 个配置项，执行一次 Namespace 策略发布，
// 此外，每个应用进行5次实例发布。
func genBaseData() error {
	concurrence := 50
	wg := sync.WaitGroup{}
	wg.Add(concurrence)

	for i := 0; i < concurrence; i++ {
		go func(i int) {
			for bizID := i*2000/concurrence + 1; bizID < (i+1)*2000/concurrence+1; bizID++ {
				if err := genAppData(uint32(bizID)); err != nil {
					log.Fatalln(err)
				}
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
	return nil
}

func genAppData(bizID uint32) error {
	for i := 0; i < 50; i++ {
		// create app.
		appReq := &pbcs.CreateAppReq{
			BizId:          bizID,
			Name:           randName("app"),
			ConfigType:     string(table.File),
			Mode:           string(table.Namespace),
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
			if err := genCIRelatedData(bizID, appResp.Data.Id); err != nil {
				return err
			}
		}

		// create release.
		rlReq := &pbcs.CreateReleaseReq{
			BizId: bizID,
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
			BizId: bizID,
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
			BizId:         bizID,
			AppId:         appResp.Data.Id,
			StrategySetId: setResp.Data.Id,
			ReleaseId:     rlResp.Data.Id,
			AsDefault:     false,
			Name:          randName("strategy"),
			Namespace:     uuid.UUID(),
			Memo:          memo,
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
			BizId: bizID,
			AppId: appResp.Data.Id,
		}
		rid = RequestID()
		_, err = cli.Publish.PublishWithStrategy(context.Background(), Header(rid), pbReq)
		if err != nil {
			return fmt.Errorf("create strategy publish err, %v, rid: %s", err, rid)
		}

		// publish instance.
		for i := 0; i < 5; i++ {
			ipReq := &pbcs.PublishInstanceReq{
				BizId:     bizID,
				AppId:     appResp.Data.Id,
				Uid:       uuid.UUID(),
				ReleaseId: rlResp.Data.Id,
				Memo:      memo,
			}
			rid = RequestID()
			ipResp, err := cli.Instance.Publish(context.Background(), Header(rid), ipReq)
			if err != nil {
				return fmt.Errorf("create instance publish err, %v, rid: %s", err, rid)
			}
			if ipResp.Code != errf.OK {
				return fmt.Errorf("create instance publish failed, code: %d, msg: %s, rid: %s", ipResp.Code,
					ipResp.Message, rid)
			}
		}
	}

	return nil
}

// genCIRelatedData gen five config item for every app, and create one content and commit for every config item.
func genCIRelatedData(bizID, appID uint32) error {
	// create config item.
	ciReq := &pbcs.CreateConfigItemReq{
		BizId:     bizID,
		AppId:     appID,
		Name:      uuid.UUID() + ".yaml",
		Path:      "/etc",
		FileType:  string(table.Yaml),
		FileMode:  string(table.Unix),
		Memo:      memo,
		User:      "root",
		UserGroup: "root",
		Privilege: "755",
	}
	rid := RequestID()
	ciResp, err := cli.ConfigItem.Create(context.Background(), Header(rid), ciReq)
	if err != nil {
		return fmt.Errorf("create config item err, %v, rid: %s", err, rid)
	}
	if ciResp.Code != errf.OK {
		return fmt.Errorf("create config item failed, code: %d, msg: %s, rid: %s", ciResp.Code, ciResp.Message, rid)
	}

	// create content.
	conReq := &pbcs.CreateContentReq{
		BizId:        bizID,
		AppId:        appID,
		ConfigItemId: ciResp.Data.Id,
		Sign:         "c7d78b78205a2619eb2b80558f85ee18a8836ef5f4f317f8587ee38bc3712a8a",
		ByteSize:     11,
	}
	rid = RequestID()
	conResp, err := cli.Content.Create(context.Background(), Header(rid), conReq)
	if err != nil {
		return fmt.Errorf("create content err, %v, rid: %s", err, rid)
	}
	if conResp.Code != errf.OK {
		return fmt.Errorf("create content failed, code: %d, msg: %s, rid: %s", conResp.Code, conResp.Message, rid)
	}

	// create commit.
	comReq := &pbcs.CreateCommitReq{
		BizId:        bizID,
		AppId:        appID,
		ConfigItemId: ciResp.Data.Id,
		ContentId:    conResp.Data.Id,
		Memo:         memo,
	}
	rid = RequestID()
	comResp, err := cli.Commit.Create(context.Background(), Header(rid), comReq)
	if err != nil {
		return fmt.Errorf("create commit err, %v, rid: %s", err, rid)
	}
	if comResp.Code != errf.OK {
		return fmt.Errorf("create commit failed, code: %d, msg: %s, rid: %s", comResp.Code, comResp.Message, rid)
	}

	return nil
}
