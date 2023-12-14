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
	"log"
	"sync"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

var (
	// 协程并发数
	concurrence = 10
	// 业务总数，可多次运行，分别调整该值为10、100、1000
	// 对比不同业务、应用量级下的性能情况(10业务-500应用；100业务-5000应用；1000业务-50000应用)
	bizCnt = 10
	// 单个业务下的应用数
	appCnt = 50
	// 单个应用下的配置项数
	cfgItemCnt = 5
	// 实例发布次数
	publishInstCnt = 5
)

// genBaseData 在业务id为 1-10 的业务下，生成 50 个应用，且都生成 5 个配置项，执行一次 Namespace 策略发布，
// 此外，每个应用进行5次实例发布。
func genBaseData() error {
	wg := sync.WaitGroup{}
	wg.Add(concurrence)

	for i := 0; i < concurrence; i++ {
		go func(i int) {
			for bizID := i*bizCnt/concurrence + 1; bizID < (i+1)*bizCnt/concurrence+1; bizID++ {
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
	for i := 0; i < appCnt; i++ {
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

		// gen five config item for every app, and create one content and commit for every config item.
		for i := 0; i < cfgItemCnt; i++ {
			if err := genCIRelatedData(bizID, appResp.Id); err != nil {
				return err
			}
		}

		// create release.
		rlReq := &pbcs.CreateReleaseReq{
			BizId: bizID,
			AppId: appResp.Id,
			Name:  randName("release"),
			Memo:  memo,
		}
		rid = RequestID()
		rlResp, err := cli.Release.Create(context.Background(), Header(rid), rlReq)
		if err != nil {
			return fmt.Errorf("create release err, %v, rid: %s", err, rid)
		}

		// NOTE: strategy related test depends on group, add group test first
		//// create strategy set.
		//setReq := &pbcs.CreateStrategySetReq{
		//	BizId: bizID,
		//	AppId: appResp.Id,
		//	Name:  randName("strategy_set"),
		//	Memo:  memo,
		//}
		//rid = RequestID()
		//setResp, err := cli.StrategySet.Create(context.Background(), Header(rid), setReq)
		//if err != nil {
		//	return fmt.Errorf("create strategy set err, %v, rid: %s", err, rid)
		//}
		//
		//// create strategy.
		//styReq := &pbcs.CreateStrategyReq{
		//	BizId:         bizID,
		//	AppId:         appResp.Id,
		//	StrategySetId: setResp.Id,
		//	ReleaseId:     rlResp.Id,
		//	AsDefault:     false,
		//	Name:          randName("strategy"),
		//	Namespace:     uuid.UUID(),
		//	Memo:          memo,
		//}
		//rid = RequestID()
		//_, err = cli.Strategy.Create(context.Background(), Header(rid), styReq)
		//if err != nil {
		//	return fmt.Errorf("create strategy err, %v, rid: %s", err, rid)
		//}
		//
		//// publish strategy.
		//pbReq := &pbcs.PublishReq{
		//	BizId: bizID,
		//	AppId: appResp.Id,
		//}
		//rid = RequestID()
		//_, err = cli.Publish.PublishWithStrategy(context.Background(), Header(rid), pbReq)
		//if err != nil {
		//	return fmt.Errorf("create strategy publish err, %v, rid: %s", err, rid)
		//}

		// publish instance.
		for i := 0; i < publishInstCnt; i++ {
			ipReq := &pbcs.PublishInstanceReq{
				BizId:     bizID,
				AppId:     appResp.Id,
				Uid:       uuid.UUID(),
				ReleaseId: rlResp.Id,
				Memo:      memo,
			}
			rid = RequestID()
			_, err = cli.Instance.Publish(context.Background(), Header(rid), ipReq)
			if err != nil {
				return fmt.Errorf("create instance publish err, %v, rid: %s", err, rid)
			}
		}
	}

	return nil
}

// genCIRelatedData gen five config item for every app, and create one content and commit for every config item.
func genCIRelatedData(bizID, appID uint32) error {
	content := "This is content for test"
	signature := tools.SHA256(content)
	size := uint64(len(content))

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
		Sign:      signature,
		ByteSize:  size,
	}
	rid := RequestID()
	_, err := cli.ConfigItem.Create(context.Background(), Header(rid), ciReq)
	if err != nil {
		return fmt.Errorf("create config item err, %v, rid: %s", err, rid)
	}

	// create ConfigItem will create content and commit too, so no need to create them
	//// upload content.
	//rid = RequestID()
	//header := Header(rid)
	//header.Set(constant.ContentIDHeaderKey, signature)
	//_, err = cli.Content.Upload(context.Background(), header, bizID, appID, content)
	//if err != nil {
	//	return fmt.Errorf("upload content err, %v", err)
	//}
	//
	//// create content.
	//conReq := &pbcs.CreateContentReq{
	//	BizId:        bizID,
	//	AppId:        appID,
	//	ConfigItemId: ciResp.Id,
	//	Sign:         signature,
	//	ByteSize:     size,
	//}
	//rid = RequestID()
	//conResp, err := cli.Content.Create(context.Background(), Header(rid), conReq)
	//if err != nil {
	//	return fmt.Errorf("create content err, %v, rid: %s", err, rid)
	//}
	//
	//// create commit.
	//comReq := &pbcs.CreateCommitReq{
	//	BizId:        bizID,
	//	AppId:        appID,
	//	ConfigItemId: ciResp.Id,
	//	ContentId:    conResp.Id,
	//	Memo:         memo,
	//}
	//rid = RequestID()
	//_, err = cli.Commit.Create(context.Background(), Header(rid), comReq)
	//if err != nil {
	//	return fmt.Errorf("create commit err, %v, rid: %s", err, rid)
	//}

	return nil
}
