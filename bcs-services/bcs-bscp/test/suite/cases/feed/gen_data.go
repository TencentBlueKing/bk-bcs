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

package feed

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

const (
	// InstanceUID feed测试所用的uid.
	InstanceUID = "bscp-feed-suite-test-instance"
	// TestDataMemo test data memo content.
	TestDataMemo = "feed suite test data"
	// BNSNamespace base namespace publish used namespace.
	BNSNamespace = "base-namespace"
)

// feed server test need request param.
var (
	BaseNormalTestAppID         uint32
	BNMDefaultStrategyReleaseID uint32
	BNMMainStrategyReleaseID    uint32
	BNMSubStrategyReleaseID     uint32
	BNMInstancePublishReleaseID uint32

	BaseNamespaceTestAppID      uint32
	BNSDefaultStrategyReleaseID uint32
	BNSNamespaceReleaseID       uint32
	BNSSubStrategyReleaseID     uint32
	BNSInstancePublishReleaseID uint32
)

// Generator define data generator.
type Generator struct {
	Cli *api.Client
}

// GenData gen need scene match related data.
func (g *Generator) GenData(kt *kit.Kit) error {
	header := http.Header{}
	header.Set(constant.UserKey, constant.BKUserForTestPrefix+"gen-data")
	header.Set(constant.RidKey, kt.Rid)
	header.Set(constant.AppCodeKey, "test")
	header.Add("Cookie", "bk_token="+constant.BKTokenForTest)

	if err := g.GenBaseNormalData(kt.Ctx, header); err != nil {
		return fmt.Errorf("gen base normal data failed, err: %v, rid: %s", err, kt.Rid)
	}

	if err := g.GenBaseNamespaceData(kt.Ctx, header); err != nil {
		return fmt.Errorf("gen base namespace data failed, err: %v, rid: %s", err, kt.Rid)
	}

	return nil
}

// GenBaseNormalData gen base normal publish way related data.
// Data: 在 app_id = ${BaseNormalTestAppID} 的应用下，进行一次兜底策略发布；进行一次带子策略的Normal策略发布；进行一次实例发布。
// Test Scene:
// 1. 匹配兜底策略
// 2. 匹配主策略
// 3. 匹配子策略
// 5. 匹配实例发布
func (g *Generator) GenBaseNormalData(ctx context.Context, header http.Header) error {
	appSpec := &table.AppSpec{
		Name:       cases.RandName("app"),
		ConfigType: table.File,
		Mode:       table.Normal,
	}
	appID, err := g.genAppData(ctx, header, appSpec)
	if err != nil {
		return err
	}
	BaseNormalTestAppID = appID

	// gen five config item for every app, and create one content and commit for every config item.
	for i := 0; i < 5; i++ {
		if err = g.genCIRelatedData(ctx, header, cases.TBizID, BaseNormalTestAppID); err != nil {
			return fmt.Errorf("gen ci related failed, err: %v", err)
		}
	}

	// NOTE: strategy related test depends on group, add group test first
	//stgSetSpec := &table.StrategySetSpec{
	//	Name: cases.RandName("strategy_set"),
	//}
	//stgSetID, err := g.genStrategySetData(ctx, header, stgSetSpec, BaseNormalTestAppID)
	//if err != nil {
	//	return err
	//}
	//
	//// 1. exec one default strategy publish.
	//releaseID, err := g.defaultStrategyPublish(ctx, header, BaseNormalTestAppID, stgSetID)
	//if err != nil {
	//	return fmt.Errorf("exec default strategy publish failed, err: %v", err)
	//}
	//BNMDefaultStrategyReleaseID = releaseID
	//
	//// 2. exec one normal strategy with sub strategy publish.
	//if err = g.normalStrategyPublish(ctx, header, BaseNormalTestAppID, stgSetID); err != nil {
	//	return fmt.Errorf("exec normal strategy publish failed, err: %v", err)
	//}

	// 3. exec one instance publish.
	releaseID, err := g.instancePublish(ctx, header, BaseNormalTestAppID)
	if err != nil {
		return fmt.Errorf("exec instance publish failed, err: %v", err)
	}
	BNMInstancePublishReleaseID = releaseID

	return nil
}

// GenBaseNamespaceData gen base namespace publish way related data.
// Data: 在 app_id = ${BaseNamespaceTestAppID} 的应用下，进行一次兜底策略发布；
//
//	进行一次带子策略的Namespace策略发布；进行一次实例发布。
//
// Test Scene:
// 1. 匹配兜底策略
// 2. 匹配Namespace
// 3. 匹配子策略
// 5. 匹配实例发布
func (g *Generator) GenBaseNamespaceData(ctx context.Context, header http.Header) error {
	appSpec := &table.AppSpec{
		Name:       cases.RandName("app"),
		ConfigType: table.File,
		Mode:       table.Namespace,
	}
	appID, err := g.genAppData(ctx, header, appSpec)
	if err != nil {
		return err
	}
	BaseNamespaceTestAppID = appID

	// gen five config item for every app, and create one content and commit for every config item.
	for i := 0; i < 5; i++ {
		if err = g.genCIRelatedData(ctx, header, cases.TBizID, BaseNamespaceTestAppID); err != nil {
			return fmt.Errorf("gen ci related failed, err: %v", err)
		}
	}

	// NOTE: strategy related test depends on group, add group test first
	//stgSetSpec := &table.StrategySetSpec{
	//	Name: cases.RandName("strategy_set"),
	//}
	//stgSetID, err := g.genStrategySetData(ctx, header, stgSetSpec, BaseNamespaceTestAppID)
	//if err != nil {
	//	return err
	//}
	//
	//// 1. exec one default strategy publish.
	//releaseID, err := g.defaultStrategyPublish(ctx, header, BaseNamespaceTestAppID, stgSetID)
	//if err != nil {
	//	return fmt.Errorf("exec default strategy publish failed, err: %v", err)
	//}
	//BNSDefaultStrategyReleaseID = releaseID
	//
	//// 2. exec one namespace strategy with sub strategy publish.
	//if err = g.namespaceStrategyPublish(ctx, header, BaseNamespaceTestAppID, stgSetID); err != nil {
	//	return fmt.Errorf("exec namespace strategy publish failed, err: %v", err)
	//}

	// 3. exec one instance publish.
	releaseID, err := g.instancePublish(ctx, header, BaseNamespaceTestAppID)
	if err != nil {
		return fmt.Errorf("exec instance publish failed, err: %v", err)
	}
	BNSInstancePublishReleaseID = releaseID

	return nil
}

func (g *Generator) instancePublish(ctx context.Context, header http.Header, appID uint32) (uint32, error) {
	// create release.
	rlReq := &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  TestDataMemo,
	}
	rlResp, err := g.Cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return 0, fmt.Errorf("create release err, %v", err)
	}

	// publish instance.
	ipReq := &pbcs.PublishInstanceReq{
		BizId:     cases.TBizID,
		AppId:     appID,
		Uid:       InstanceUID,
		ReleaseId: rlResp.Id,
		Memo:      TestDataMemo,
	}
	_, err = g.Cli.Instance.Publish(ctx, header, ipReq)
	if err != nil {
		return 0, fmt.Errorf("create instance publish err, %v", err)
	}

	return rlResp.Id, nil
}

func (g *Generator) namespaceStrategyPublish(ctx context.Context, header http.Header, appID, stgSetID uint32) error {
	// create release.
	rlReq := &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  TestDataMemo,
	}
	rlResp, err := g.Cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v", err)
	}

	BNSNamespaceReleaseID = rlResp.Id

	// create release.
	rlReq = &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  TestDataMemo,
	}
	rlResp, err = g.Cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v", err)
	}

	BNSSubStrategyReleaseID = rlResp.Id

	// create strategy.
	styReq := &pbcs.CreateStrategyReq{
		BizId:         cases.TBizID,
		AppId:         appID,
		StrategySetId: stgSetID,
		Name:          cases.RandName("strategy"),
		Namespace:     BNSNamespace,
		Memo:          TestDataMemo,
		ReleaseId:     BNSNamespaceReleaseID,
	}
	_, err = g.Cli.Strategy.Create(ctx, header, styReq)
	if err != nil {
		return fmt.Errorf("create strategy err, %v", err)
	}

	// publish strategy.
	pbReq := &pbcs.PublishReq{
		BizId: cases.TBizID,
		AppId: appID,
	}
	_, err = g.Cli.Publish.PublishWithStrategy(ctx, header, pbReq)
	if err != nil {
		return fmt.Errorf("create strategy publish err, %v", err)
	}

	return nil
}

func (g *Generator) normalStrategyPublish(ctx context.Context, header http.Header, appID, stgSetID uint32) error {
	// create release.
	rlReq := &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  TestDataMemo,
	}
	rlResp, err := g.Cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v", err)
	}

	BNMMainStrategyReleaseID = rlResp.Id

	// create release.
	rlReq = &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  TestDataMemo,
	}
	rlResp, err = g.Cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return fmt.Errorf("create release err, %v", err)
	}

	BNMSubStrategyReleaseID = rlResp.Id

	// create strategy.
	styReq := &pbcs.CreateStrategyReq{
		BizId:         cases.TBizID,
		AppId:         appID,
		StrategySetId: stgSetID,
		Name:          cases.RandName("strategy"),
		Memo:          TestDataMemo,
		ReleaseId:     BNMMainStrategyReleaseID,
	}
	_, err = g.Cli.Strategy.Create(ctx, header, styReq)
	if err != nil {
		return fmt.Errorf("create strategy err, %v", err)
	}

	// publish strategy.
	pbReq := &pbcs.PublishReq{
		BizId: cases.TBizID,
		AppId: appID,
	}
	_, err = g.Cli.Publish.PublishWithStrategy(ctx, header, pbReq)
	if err != nil {
		return fmt.Errorf("create strategy publish err, %v", err)
	}

	return nil
}

func (g *Generator) defaultStrategyPublish(ctx context.Context, header http.Header, appID,
	stgSetID uint32) (uint32, error) {

	// create release.
	rlReq := &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  cases.RandName("release"),
		Memo:  TestDataMemo,
	}
	rlResp, err := g.Cli.Release.Create(ctx, header, rlReq)
	if err != nil {
		return 0, fmt.Errorf("create release err, %v", err)
	}

	// create strategy.
	styReq := &pbcs.CreateStrategyReq{
		BizId:         cases.TBizID,
		AppId:         appID,
		StrategySetId: stgSetID,
		Name:          cases.RandName("strategy"),
		AsDefault:     true,
		Memo:          TestDataMemo,
		ReleaseId:     rlResp.Id,
	}
	_, err = g.Cli.Strategy.Create(ctx, header, styReq)
	if err != nil {
		return 0, fmt.Errorf("create strategy err, %v", err)
	}

	// publish strategy.
	pbReq := &pbcs.PublishReq{
		BizId: cases.TBizID,
		AppId: appID,
	}
	_, err = g.Cli.Publish.PublishWithStrategy(ctx, header, pbReq)
	if err != nil {
		return 0, fmt.Errorf("create strategy publish err, %v", err)
	}

	return rlResp.Id, nil
}

func (g *Generator) genAppData(ctx context.Context, header http.Header, spec *table.AppSpec) (uint32, error) {
	appReq := &pbcs.CreateAppReq{
		BizId:          cases.TBizID,
		Name:           spec.Name,
		ConfigType:     string(spec.ConfigType),
		Mode:           string(spec.Mode),
		Memo:           TestDataMemo,
		ReloadType:     string(table.ReloadWithFile),
		ReloadFilePath: "/tmp/reload.json",
	}
	resp, err := g.Cli.App.Create(ctx, header, appReq)
	if err != nil {
		return 0, fmt.Errorf("create app err, %v", err)
	}

	return resp.Id, nil
}

func (g *Generator) genStrategySetData(ctx context.Context, header http.Header,
	spec *table.StrategySetSpec, appID uint32) (uint32, error) {

	// create strategy set.
	setReq := &pbcs.CreateStrategySetReq{
		BizId: cases.TBizID,
		AppId: appID,
		Name:  spec.Name,
		Memo:  TestDataMemo,
	}
	resp, err := g.Cli.StrategySet.Create(ctx, header, setReq)
	if err != nil {
		return 0, fmt.Errorf("create strategy set err, %v", err)
	}

	return resp.Id, nil
}

// genCIRelatedData gen five config item for every app, and create one content and commit for every config item.
func (g *Generator) genCIRelatedData(ctx context.Context, header http.Header, bizID, appID uint32) error {
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
		Memo:      TestDataMemo,
		User:      "root",
		UserGroup: "root",
		Privilege: "755",
		Sign:      signature,
		ByteSize:  size,
	}
	ciResp, err := g.Cli.ConfigItem.Create(ctx, header, ciReq)
	if err != nil {
		return fmt.Errorf("create config item err, %v", err)
	}

	// upload content.
	header.Set(constant.ContentIDHeaderKey, signature)
	_, err = g.Cli.Content.Upload(ctx, header, cases.TBizID, appID, content)
	if err != nil {
		return fmt.Errorf("upload content err, %v", err)
	}

	// create content.
	conReq := &pbcs.CreateContentReq{
		BizId:        bizID,
		AppId:        appID,
		ConfigItemId: ciResp.Id,
		Sign:         signature,
		ByteSize:     size,
	}
	conResp, err := g.Cli.Content.Create(ctx, header, conReq)
	if err != nil {
		return fmt.Errorf("create content err, %v", err)
	}

	// create commit.
	comReq := &pbcs.CreateCommitReq{
		BizId:        bizID,
		AppId:        appID,
		ConfigItemId: ciResp.Id,
		ContentId:    conResp.Id,
		Memo:         TestDataMemo,
	}
	_, err = g.Cli.Commit.Create(ctx, header, comReq)
	if err != nil {
		return fmt.Errorf("create commit err, %v", err)
	}

	return nil
}
