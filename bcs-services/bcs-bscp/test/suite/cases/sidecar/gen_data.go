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

package sidecar

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbci "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	sfs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

const (
	// testBizID is sidecar test use biz id.
	testBizID = math.MaxUint32 / 2
	// testDataMemo test data memo content.
	testDataMemo = "sidecar suite test data"
	// testContent test content's content.
	testContent = "IP: localhost"
	// testInstanceUID sidecar instance uid for sidecar suite test.
	testInstanceUID = "bscp-sidecar-suite-test-uid"
	// testNamespace sidecar namespace for sidecar suite test.
	testNamespace = "sidecar-suite-test-namespace"
)

var (
	testApp1ID            uint32
	testApp2ID            uint32
	testApp3ID            uint32
	testApp3StrategySetID uint32
)

// AppReleaseMeta define app release meta struct.
type AppReleaseMeta map[uint32][]*ReleaseMeta

// ReleaseMeta defines release meta struct.
type ReleaseMeta struct {
	releaseID         uint32
	instancePublishID uint32
	ciMeta            []*sfs.ConfigItemMetaV1
}

// generator define data generator.
type generator struct {
	cli  *api.Client
	data AppReleaseMeta
}

// InitData gen need init scene related data.
func (g *generator) InitData(kt *kit.Kit) error {
	header := http.Header{}
	header.Set(constant.UserKey, constant.BKUserForTestPrefix+"gen-data")
	header.Set(constant.RidKey, kt.Rid)
	header.Set(constant.AppCodeKey, "test")
	header.Add("Cookie", "bk_token="+constant.BKTokenForTest)

	g.data = make(AppReleaseMeta, 0)

	//if err := g.initApp1(kt.Ctx, header); err != nil {
	//	return err
	//}

	if err := g.initApp2(kt.Ctx, header); err != nil {
		return err
	}

	//if err := g.initApp3(kt.Ctx, header); err != nil {
	//	return err
	//}

	return nil
}

// SimulationData simulation app publish related operation base on init data.
func (g *generator) SimulationData(kt *kit.Kit) error {

	header := http.Header{}
	header.Set(constant.UserKey, constant.BKUserForTestPrefix+"gen-data")
	header.Set(constant.RidKey, kt.Rid)
	header.Set(constant.AppCodeKey, "test")
	header.Add("Cookie", "bk_token="+constant.BKTokenForTest)

	if err := g.simulationApp1(kt.Ctx, header); err != nil {
		return err
	}

	if err := g.simulationApp2(kt.Ctx, header); err != nil {
		return err
	}

	// NOTE: strategy related test depends on group, add group test first
	//if err := g.simulationApp3(kt.Ctx, header); err != nil {
	//	return err
	//}

	return nil
}

func (g *generator) initApp1(ctx context.Context, header http.Header) error {
	appSpec := &table.AppSpec{
		Name:       cases.RandName("app"),
		ConfigType: table.File,
		Mode:       table.Normal,
	}
	appID, err := g.genAppData(ctx, header, appSpec)
	if err != nil {
		return err
	}

	testApp1ID = appID
	g.data[appID] = make([]*ReleaseMeta, 0)

	ciMeta, err := g.genCIRelatedData(ctx, header, appID)
	if err != nil {
		return err
	}

	stgSetSpec := &table.StrategySetSpec{
		Name: cases.RandName("strategy_set"),
	}
	stgSetID, err := g.genStrategySetData(ctx, header, stgSetSpec, appID)
	if err != nil {
		return err
	}

	rlReq := &pbcs.CreateReleaseReq{
		BizId: testBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  testDataMemo,
	}
	rlResp, err := g.cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v", err)
	}

	// create strategy.
	styReq := &pbcs.CreateStrategyReq{
		BizId:         testBizID,
		AppId:         appID,
		StrategySetId: stgSetID,
		Name:          cases.RandName("strategy"),
		AsDefault:     true,
		Memo:          testDataMemo,
		ReleaseId:     rlResp.Id,
	}
	_, err = g.cli.Strategy.Create(ctx, header, styReq)
	if err != nil {
		return fmt.Errorf("create strategy err, %v", err)
	}

	// publish strategy.
	pbReq := &pbcs.PublishReq{
		BizId: testBizID,
		AppId: appID,
	}
	_, err = g.cli.Publish.PublishWithStrategy(ctx, header, pbReq)
	if err != nil {
		return fmt.Errorf("create strategy publish err, %v", err)
	}

	g.data[appID] = append(g.data[appID], &ReleaseMeta{
		releaseID: rlResp.Id,
		ciMeta:    ciMeta,
	})

	return nil
}

func (g *generator) initApp2(ctx context.Context, header http.Header) error {
	appSpec := &table.AppSpec{
		Name:       cases.RandName("app"),
		ConfigType: table.File,
		Mode:       table.Normal,
	}
	appID, err := g.genAppData(ctx, header, appSpec)
	if err != nil {
		return err
	}

	testApp2ID = appID
	g.data[appID] = make([]*ReleaseMeta, 0)

	ciMeta, err := g.genCIRelatedData(ctx, header, appID)
	if err != nil {
		return err
	}

	// NOTE: strategy related test depends on group, add group test first
	//stgSetSpec := &table.StrategySetSpec{
	//	Name: cases.RandName("strategy_set"),
	//}
	//stgSetID, err := g.genStrategySetData(ctx, header, stgSetSpec, appID)
	//if err != nil {
	//	return err
	//}
	//
	//rlReq := &pbcs.CreateReleaseReq{
	//	BizId: testBizID,
	//	AppId: appID,
	//	Name:  cases.RandName("release"),
	//	Memo:  testDataMemo,
	//}
	//rlResp, err := g.cli.Release.Create(ctx, header, rlReq)
	//if err != nil {
	//	return fmt.Errorf("create release err, %v", err)
	//}
	//
	//
	//// create strategy.
	//styReq := &pbcs.CreateStrategyReq{
	//	BizId:         testBizID,
	//	AppId:         appID,
	//	StrategySetId: stgSetID,
	//	Name:          cases.RandName("strategy"),
	//	AsDefault:     true,
	//	Memo:          testDataMemo,
	//	ReleaseId:     rlResp.Id,
	//}
	//_, err = g.cli.Strategy.Create(ctx, header, styReq)
	//if err != nil {
	//	return fmt.Errorf("create strategy err, %v", err)
	//}
	//
	//// publish strategy.
	//pbReq := &pbcs.PublishReq{
	//	BizId: testBizID,
	//	AppId: appID,
	//}
	//_, err = g.cli.Publish.PublishWithStrategy(ctx, header, pbReq)
	//if err != nil {
	//	return fmt.Errorf("create strategy publish err, %v", err)
	//}
	//
	//// record sidecar can match release info.
	//g.data[appID] = append(g.data[appID], &ReleaseMeta{
	//	releaseID: rlResp.Id,
	//	ciMeta:    ciMeta,
	//})

	for i := 0; i < 5; i++ {
		rlReq := &pbcs.CreateReleaseReq{
			BizId: testBizID,
			AppId: appID,
			Name:  cases.RandName("release"),
			Memo:  testDataMemo,
		}
		rlResp, err := g.cli.Release.Create(ctx, header, rlReq)
		if err != nil {
			return fmt.Errorf("create release err, %v", err)
		}

		ipReq := &pbcs.PublishInstanceReq{
			BizId:     testBizID,
			AppId:     appID,
			Uid:       testInstanceUID,
			ReleaseId: rlResp.Id,
			Memo:      testDataMemo,
		}
		if i != 0 {
			ipReq.Uid = testInstanceUID + strconv.Itoa(i)
		}

		ipResp, err := g.cli.Instance.Publish(ctx, header, ipReq)
		if err != nil {
			return fmt.Errorf("create instance publish err, %v", err)
		}

		if i == 0 {
			// record sidecar can match release info.
			g.data[appID] = append(g.data[appID], &ReleaseMeta{
				releaseID:         rlResp.Id,
				instancePublishID: ipResp.Id,
				ciMeta:            ciMeta,
			})
		}
	}

	return nil
}

func (g *generator) initApp3(ctx context.Context, header http.Header) error {
	appSpec := &table.AppSpec{
		Name:       cases.RandName("app"),
		ConfigType: table.File,
		Mode:       table.Namespace,
	}
	appID, err := g.genAppData(ctx, header, appSpec)
	if err != nil {
		return err
	}

	testApp3ID = appID
	g.data[appID] = make([]*ReleaseMeta, 0)

	ciMeta, err := g.genCIRelatedData(ctx, header, appID)
	if err != nil {
		return err
	}

	stgSetSpec := &table.StrategySetSpec{
		Name: cases.RandName("strategy_set"),
	}
	stgSetID, err := g.genStrategySetData(ctx, header, stgSetSpec, appID)
	if err != nil {
		return err
	}

	testApp3StrategySetID = stgSetID

	for i := 0; i < 5; i++ {
		rlReq := &pbcs.CreateReleaseReq{
			BizId: testBizID,
			AppId: appID,
			Name:  cases.RandName("release"),
			Memo:  testDataMemo,
		}
		rlResp, err := g.cli.Release.Create(ctx, header, rlReq)
		if err != nil {
			return fmt.Errorf("create release err, %v", err)
		}

		// create strategy.
		styReq := &pbcs.CreateStrategyReq{
			BizId:         testBizID,
			AppId:         appID,
			StrategySetId: stgSetID,
			Name:          cases.RandName("strategy"),
			Namespace:     testNamespace,
			Memo:          testDataMemo,
			ReleaseId:     rlResp.Id,
		}
		if i != 0 {
			styReq.Namespace = testNamespace + strconv.Itoa(i)
		}
		_, err = g.cli.Strategy.Create(ctx, header, styReq)
		if err != nil {
			return fmt.Errorf("create strategy err, %v", err)
		}

		// publish strategy.
		pbReq := &pbcs.PublishReq{
			BizId: testBizID,
			AppId: appID,
		}
		_, err = g.cli.Publish.PublishWithStrategy(ctx, header, pbReq)
		if err != nil {
			return fmt.Errorf("create strategy publish err, %v", err)
		}

		if i == 0 {
			// record sidecar can match release info.
			g.data[appID] = append(g.data[appID], &ReleaseMeta{
				releaseID: rlResp.Id,
				ciMeta:    ciMeta,
			})
		}
	}
	return nil
}

func (g *generator) genAppData(ctx context.Context, header http.Header, spec *table.AppSpec) (uint32, error) {
	appReq := &pbcs.CreateAppReq{
		BizId:          testBizID,
		Name:           spec.Name,
		ConfigType:     string(spec.ConfigType),
		Mode:           string(spec.Mode),
		Memo:           testDataMemo,
		ReloadType:     string(table.ReloadWithFile),
		ReloadFilePath: "/tmp/reload.json",
	}
	resp, err := g.cli.App.Create(ctx, header, appReq)
	if err != nil {
		return 0, fmt.Errorf("create app err, %v", err)
	}

	return resp.Id, nil
}

func (g *generator) genCIRelatedData(ctx context.Context, header http.Header, appID uint32) ([]*sfs.ConfigItemMetaV1,
	error) {

	sign := tools.SHA256(testContent)
	size := uint64(len(testContent))
	result := make([]*sfs.ConfigItemMetaV1, 0)
	for i := 0; i < 2; i++ {
		// create config item.
		ciReq := &pbcs.CreateConfigItemReq{
			BizId:     testBizID,
			AppId:     appID,
			Name:      tools.Itoa(appID) + "-" + strconv.Itoa(i) + ".yaml",
			Path:      "/etc",
			FileType:  string(table.Yaml),
			FileMode:  string(table.Unix),
			Memo:      testDataMemo,
			User:      "root",
			UserGroup: "root",
			Privilege: "755",
			Sign:      sign,
			ByteSize:  size,
		}
		_, err := g.cli.ConfigItem.Create(ctx, header, ciReq)
		if err != nil {
			return nil, fmt.Errorf("create config item err, %v", err)
		}

		header.Set(constant.ContentIDHeaderKey, sign)
		uploadResp, err := g.cli.Content.Upload(context.Background(), header, testBizID, appID, testContent)
		if err != nil {
			return nil, fmt.Errorf("upload content failed, err: %v", err)
		}
		if uploadResp.Code != errf.OK {
			return nil, fmt.Errorf("upload content failed, code: %d, msg: %s", uploadResp.Code, uploadResp.Message)
		}

		// create ConfigItem will create content and commit too, so no need to create them
		//// create content.
		//conReq := &pbcs.CreateContentReq{
		//	BizId:        testBizID,
		//	AppId:        appID,
		//	ConfigItemId: ciResp.Id,
		//	Sign:         sign,
		//	ByteSize:     size,
		//}
		//conResp, err := g.cli.Content.Create(ctx, header, conReq)
		//if err != nil {
		//	return nil, fmt.Errorf("create content err, %v", err)
		//}
		//
		//// create commit.
		//comReq := &pbcs.CreateCommitReq{
		//	BizId:        testBizID,
		//	AppId:        appID,
		//	ConfigItemId: ciResp.Id,
		//	ContentId:    conResp.Id,
		//	Memo:         testDataMemo,
		//}
		//_, err = g.cli.Commit.Create(ctx, header, comReq)
		//if err != nil {
		//	return nil, fmt.Errorf("create commit err, %v", err)
		//}

		result = append(result, &sfs.ConfigItemMetaV1{
			ContentSpec: &pbcontent.ContentSpec{
				Signature: sign,
				ByteSize:  uint64(len(testContent)),
			},
			ConfigItemSpec: &pbci.ConfigItemSpec{
				Name:     ciReq.Name,
				Path:     ciReq.Path,
				FileType: ciReq.FileType,
				FileMode: ciReq.FileMode,
				Memo:     ciReq.Memo,
				Permission: &pbci.FilePermission{
					User:      ciReq.User,
					UserGroup: ciReq.UserGroup,
					Privilege: ciReq.Privilege,
				},
			},
		})
	}

	return result, nil
}

func (g *generator) genStrategySetData(ctx context.Context, header http.Header, spec *table.StrategySetSpec,
	appID uint32) (uint32, error) {

	// create strategy set.
	setReq := &pbcs.CreateStrategySetReq{
		BizId: testBizID,
		AppId: appID,
		Name:  spec.Name,
		Memo:  testDataMemo,
	}
	resp, err := g.cli.StrategySet.Create(ctx, header, setReq)
	if err != nil {
		return 0, fmt.Errorf("create strategy set err, %v", err)
	}

	return resp.Id, nil
}

func (g *generator) simulationApp1(ctx context.Context, header http.Header) error {
	rlReq := &pbcs.CreateReleaseReq{
		BizId: testBizID,
		AppId: testApp1ID,
		Name:  cases.RandName("release"),
		Memo:  testDataMemo,
	}
	rlResp, err := g.cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v", err)
	}

	ipReq := &pbcs.PublishInstanceReq{
		BizId:     testBizID,
		AppId:     testApp1ID,
		Uid:       testInstanceUID,
		ReleaseId: rlResp.Id,
		Memo:      testDataMemo,
	}

	_, err = g.cli.Instance.Publish(ctx, header, ipReq)
	if err != nil {
		return fmt.Errorf("create instance publish err, %v", err)
	}

	// every app release's ci detail is same.
	g.data[testApp1ID] = append(g.data[testApp1ID], &ReleaseMeta{
		releaseID: rlResp.Id,
		ciMeta:    g.data[testApp1ID][len(g.data[testApp1ID])-1].ciMeta,
	})

	return nil
}

func (g *generator) simulationApp2(ctx context.Context, header http.Header) error {
	req := &pbcs.DeletePublishedInstanceReq{
		Id:    g.data[testApp2ID][len(g.data[testApp2ID])-1].instancePublishID,
		BizId: testBizID,
		AppId: testApp2ID,
	}
	_, err := g.cli.Instance.Delete(ctx, header, req)
	if err != nil {
		return fmt.Errorf("delete instance publish err, %v", err)
	}

	g.data[testApp2ID] = g.data[testApp2ID][0:(len(g.data[testApp2ID]) - 1)]

	return nil
}

func (g *generator) simulationApp3(ctx context.Context, header http.Header) error {
	for i := 0; i < 5; i++ {
		rlReq := &pbcs.CreateReleaseReq{
			BizId: testBizID,
			AppId: testApp3ID,
			Name:  cases.RandName("release"),
			Memo:  testDataMemo,
		}
		rlResp, err := g.cli.Release.Create(ctx, header, rlReq)
		if err != nil {
			return fmt.Errorf("create release err, %v", err)
		}

		// create strategy.
		styReq := &pbcs.CreateStrategyReq{
			BizId:         testBizID,
			AppId:         testApp3ID,
			StrategySetId: testApp3StrategySetID,
			Name:          cases.RandName("strategy"),
			Namespace:     testNamespace + "-" + strconv.Itoa(i),
			Memo:          testDataMemo,
			ReleaseId:     rlResp.Id,
		}
		_, err = g.cli.Strategy.Create(ctx, header, styReq)
		if err != nil {
			return fmt.Errorf("create strategy err, %v", err)
		}

		// publish strategy.
		pbReq := &pbcs.PublishReq{
			BizId: testBizID,
			AppId: testApp3ID,
		}
		_, err = g.cli.Publish.PublishWithStrategy(ctx, header, pbReq)
		if err != nil {
			return fmt.Errorf("create strategy publish err, %v", err)
		}
	}

	return nil
}
