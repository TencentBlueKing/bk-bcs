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

package api

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestStrategy(t *testing.T) {

	var (
		cli     *api.Client
		ctx     context.Context
		header  http.Header
		preName string

		nmAppId    uint32 // normal mode's app id
		nmRelId    uint32 // normal mode's release id
		nmStgSetId uint32 // normal mode's strategy set id

		nsAppId    uint32 // namespace mode's app id
		nsRelId    uint32 // namespace mode's release id
		nsStgSetId uint32 // namespace mode's strategy set id
	)

	Convey("Prepare For Strategy Test", t, func() {
		cli = suite.GetClient().ApiClient
		preName = "test_stg"

		nmAppId, nmRelId, nmStgSetId = createResource(cli, table.Normal, preName)
		So(nmAppId, ShouldNotEqual, uint32(0))
		So(nmRelId, ShouldNotEqual, uint32(0))
		So(nmStgSetId, ShouldNotEqual, uint32(0))

		nsAppId, nsRelId, nsStgSetId = createResource(cli, table.Namespace, preName)
		So(nsAppId, ShouldNotEqual, uint32(0))
		So(nsRelId, ShouldNotEqual, uint32(0))
		So(nsStgSetId, ShouldNotEqual, uint32(0))
	})

	Convey("Create Strategy Test", t, func() {
		Convey("1.create_strategy normal test in namespace mode", func() {
			// test cases
			reqs := make([]pbcs.CreateStrategyReq, 0)

			r := pbcs.CreateStrategyReq{
				BizId:         cases.TBizID,
				AppId:         nsAppId,
				StrategySetId: nsStgSetId,
				ReleaseId:     nsRelId,
				AsDefault:     false,
			}

			// add name field test case
			names := genNormalNameForCreateTest()
			for _, name := range names {
				r.Name = name
				r.Namespace = cases.RandName(preName)
				reqs = append(reqs, r)
			}

			// add namespace field test case
			namespaces := genNormalNameForCreateTest()
			for _, namespace := range namespaces {
				r.Namespace = namespace
				r.Name = cases.RandName(preName)
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandName(preName)
				r.Memo = memo
				r.Namespace = cases.RandName(preName)
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// verify by list_strategy
				listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nsAppId, []uint32{resp.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Strategy.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Id)

				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
				So(one.Spec.Namespace, ShouldEqual, req.Namespace)
				So(one.Spec.ReleaseId, ShouldEqual, req.ReleaseId)
				So(one.Spec.AsDefault, ShouldEqual, req.AsDefault)
				So(one.Spec.Mode, ShouldEqual, string(table.Namespace))

				So(one.Revision, cases.SoRevision)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.BizId, ShouldEqual, req.BizId)
				So(one.Attachment.AppId, ShouldEqual, req.AppId)

				So(one.State, ShouldNotBeNil)
				So(one.State.PubState, ShouldNotBeBlank)

				rm.AddStrategy(nsAppId, req.StrategySetId, resp.Id)
			}
		})

		Convey("2.create_strategy normal test in normal mode", func() {
			req := &pbcs.CreateStrategyReq{
				BizId:         cases.TBizID,
				AppId:         nmAppId,
				StrategySetId: nmStgSetId,
				ReleaseId:     nmRelId,
				AsDefault:     false,
				Namespace:     "",
				Name:          cases.RandName(preName),
			}

			ctx, header = cases.GenApiCtxHeader()
			resp, err := cli.Strategy.Create(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Id, ShouldNotEqual, uint32(0))

			// verify by list_config_item
			listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nmAppId, []uint32{resp.Id})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.Strategy.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)

			So(len(listResp.Details), ShouldEqual, 1)
			one := listResp.Details[0]
			So(one, ShouldNotBeNil)

			rm.AddStrategy(nmAppId, req.StrategySetId, resp.Id)
		})

		Convey("3.create_strategy normal test: as default is true", func() {
			req := &pbcs.CreateStrategyReq{
				BizId:         cases.TBizID,
				AppId:         nmAppId,
				StrategySetId: nmStgSetId,
				ReleaseId:     nmRelId,
				AsDefault:     true,
				Namespace:     "",
				Name:          cases.RandName(preName),
			}
			ctx, header = cases.GenApiCtxHeader()
			resp, err := cli.Strategy.Create(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Id, ShouldNotEqual, uint32(0))

			// verify by list_config_item
			listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nmAppId, []uint32{resp.Id})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.Strategy.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)

			So(len(listResp.Details), ShouldEqual, 1)
			one := listResp.Details[0]
			So(one, ShouldNotBeNil)
			So(one.Spec.AsDefault, ShouldBeTrue)

			rm.AddStrategy(nmAppId, req.StrategySetId, resp.Id)

		})

		Convey("4.create_strategy abnormal test", func() {
			// test cases
			reqs := make([]pbcs.CreateStrategyReq, 0)

			r := pbcs.CreateStrategyReq{
				BizId:         cases.TBizID,
				AppId:         nsAppId,
				StrategySetId: nsStgSetId,
				ReleaseId:     nsRelId,
				AsDefault:     false,
			}

			// AsDefault is true
			r.Name = cases.RandName(preName)
			r.Namespace = cases.RandName(preName)
			r.AsDefault = true
			reqs = append(reqs, r)
			r.AsDefault = false

			// add name field test case
			names := genAbnormalNameForTest()
			for _, name := range names {
				r.Name = name
				r.Namespace = cases.RandName(preName)
				reqs = append(reqs, r)
			}

			// add namespace field test case
			namespaces := genAbnormalNameForTest()
			for _, namespace := range namespaces {
				r.Namespace = namespace
				r.Name = cases.RandName(preName)
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandName(preName)
				r.Memo = memo
				r.Namespace = cases.RandName(preName)
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Create(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})

		Convey("5.create_strategy abnormal test: the number for strategy in an app is out of limit", func() {
			{ // namespace mode limit 200
				numToCreate := 200 - len(rm.Strategies[nsStgSetId])
				for i := 0; i < numToCreate; i++ {
					req := &pbcs.CreateStrategyReq{
						BizId:         cases.TBizID,
						AppId:         nsAppId,
						StrategySetId: nsStgSetId,
						ReleaseId:     nsRelId,
						AsDefault:     false,
						Namespace:     cases.RandName(preName),
						Name:          cases.RandName(preName),
					}
					ctx, header = cases.GenApiCtxHeader()
					resp, err := cli.Strategy.Create(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
				}

				// try to create an out of limit strategy
				req := &pbcs.CreateStrategyReq{
					BizId:         cases.TBizID,
					AppId:         nsAppId,
					StrategySetId: nsStgSetId,
					ReleaseId:     nsRelId,
					AsDefault:     false,
					Namespace:     cases.RandName(preName),
					Name:          cases.RandName(preName),
				}
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Create(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}

			{ // normal mode limit 5
				numToCreate := 5 - len(rm.Strategies[nmStgSetId])

				for i := 0; i < numToCreate; i++ {
					req := &pbcs.CreateStrategyReq{
						BizId:         cases.TBizID,
						AppId:         nmAppId,
						StrategySetId: nmStgSetId,
						ReleaseId:     nmRelId,
						AsDefault:     false,
						Name:          cases.RandName(preName),
					}
					ctx, header = cases.GenApiCtxHeader()
					resp, err := cli.Strategy.Create(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
				}

				// try to create an out of limit strategy
				req := &pbcs.CreateStrategyReq{
					BizId:         cases.TBizID,
					AppId:         nmAppId,
					StrategySetId: nmStgSetId,
					ReleaseId:     nmRelId,
					AsDefault:     false,
					Name:          cases.RandName(preName),
				}
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Create(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Update Strategy Test", t, func() {
		Convey("1.update_strategy normal test", func() {
			// get strategy id
			stgId := rm.GetStrategy(nsStgSetId)
			So(stgId, ShouldNotEqual, uint32(0))

			selector := cases.GenSubSelector()
			selector.LabelsOr[0].Value = []string{"nanjing", "hangzhou"}

			// test cases
			reqs := make([]pbcs.UpdateStrategyReq, 0)

			r := pbcs.UpdateStrategyReq{
				BizId:     cases.TBizID,
				AppId:     nsAppId,
				Id:        stgId,
				ReleaseId: nsRelId,
				AsDefault: false,
			}

			// add name field test case
			names := genNormalNameForUpdateTest()
			for _, name := range names {
				r.Name = name
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Update(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)

				// verify by list_config_item
				listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nsAppId, []uint32{req.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Strategy.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				// just verify the filed maybe update
				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
				So(one.Spec.AsDefault, ShouldEqual, req.AsDefault)
			}
		})

		Convey("2.update_strategy abnormal test", func() {
			// get strategy id
			stgId := rm.GetStrategy(nsStgSetId)
			So(stgId, ShouldNotEqual, uint32(0))
			// test cases
			reqs := make([]pbcs.UpdateStrategyReq, 0)

			r := pbcs.UpdateStrategyReq{
				BizId:     cases.TBizID,
				AppId:     nsAppId,
				Id:        stgId,
				ReleaseId: nsRelId,
				AsDefault: false,
			}

			// biz_id is invalid
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app id is invalid
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = nsAppId

			// release id is invalid
			r.ReleaseId = cases.WID
			reqs = append(reqs, r)
			r.ReleaseId = nsRelId

			// strategy id is invalid
			r.Id = cases.WID
			reqs = append(reqs, r)
			r.Id = stgId

			// add name field test case
			names := genAbnormalNameForTest()
			for _, name := range names {
				r.Name = name
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Update(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Delete Strategy Test", t, func() {
		Convey("1.delete_strategy normal test", func() {
			stgId := rm.GetStrategy(nmStgSetId)
			So(stgId, ShouldNotEqual, uint32(0))
			defer rm.DeleteStrategy(nmAppId, nmStgSetId, stgId)

			req := &pbcs.DeleteStrategyReq{
				BizId: cases.TBizID,
				AppId: nmAppId,
				Id:    stgId,
			}

			ctx, header = cases.GenApiCtxHeader()
			resp, err := cli.Strategy.Delete(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			// verify by list_strategy_set
			listReq, err := cases.GenListStrategySetByIdsReq(cases.TBizID, nmAppId, []uint32{stgId})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.StrategySet.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(len(listResp.Details), ShouldEqual, 0)
		})

		Convey("2.delete_strategy abnormal test", func() {
			stgId := rm.GetStrategy(nsStgSetId)
			So(stgId, ShouldNotEqual, uint32(0))

			reqs := []*pbcs.DeleteStrategyReq{
				{ // strategy set id is invalid
					Id:    cases.WID,
					BizId: cases.TBizID,
					AppId: nsAppId,
				},
				{ // biz id is invalid
					Id:    stgId,
					BizId: cases.WID,
					AppId: nsAppId,
				},
				{ // app id is invalid
					Id:    stgId,
					BizId: cases.TBizID,
					AppId: cases.WID,
				},
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Delete(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})

	})

	Convey("List Strategy Test", t, func() {
		// The normal list_strategy is test by the create_strategy case,
		// so we just test list_strategy normal test on count page in here
		Convey("1.list_strategy normal test: count page", func() {
			stgId := rm.GetStrategy(nmStgSetId)
			So(stgId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{stgId})
			So(err, ShouldBeNil)

			req := &pbcs.ListStrategiesReq{
				BizId:  cases.TBizID,
				AppId:  nmAppId,
				Filter: filter,
				Page:   cases.CountPage(),
			}

			ctx, header = cases.GenApiCtxHeader()
			resp, err := cli.Strategy.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_config_item abnormal test", func() {
			stgId := rm.GetStrategy(nmStgSetId)
			So(stgId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{stgId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListStrategiesReq{
				{ // biz_id is invalid
					BizId:  cases.WID,
					AppId:  nmAppId,
					Filter: filter,
					Page:   cases.CountPage(),
				},
				{ // app_id is invalid
					BizId:  cases.TBizID,
					AppId:  cases.WID,
					Filter: filter,
					Page:   cases.CountPage(),
				},
				{ // filter is invalid
					BizId:  cases.TBizID,
					AppId:  nmAppId,
					Filter: nil,
					Page:   cases.CountPage(),
				},
				{ // page is invalid
					BizId:  cases.TBizID,
					AppId:  nmAppId,
					Filter: filter,
					Page:   nil,
				},
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}

func createResource(cli *api.Client, mode table.AppMode, name string) (appId, releaseId, strategySetId uint32) {
	// create app
	appReq := &pbcs.CreateAppReq{
		BizId:          cases.TBizID,
		Name:           cases.RandName(name),
		ConfigType:     string(table.File),
		Mode:           string(mode),
		ReloadType:     string(table.ReloadWithFile),
		ReloadFilePath: "/tmp/reload.json",
	}
	ctx, header := cases.GenApiCtxHeader()
	appResp, err := cli.App.Create(ctx, header, appReq)
	So(err, ShouldBeNil)
	So(appResp, ShouldNotBeNil)
	So(appResp.Id, ShouldNotEqual, uint32(0))
	appId = appResp.Id
	rm.AddApp(mode, appId)

	content := "This is content for test"
	signature := tools.SHA256(content)
	size := uint64(len(content))

	// create config item
	ciReq := &pbcs.CreateConfigItemReq{
		BizId:     cases.TBizID,
		AppId:     appId,
		Name:      name + "_config_item",
		Path:      "/etc",
		FileType:  string(table.Xml),
		FileMode:  string(table.Unix),
		User:      "root",
		UserGroup: "root",
		Privilege: "755",
		Sign:      signature,
		ByteSize:  size,
	}
	ctx, header = cases.GenApiCtxHeader()
	ciResp, err := cli.ConfigItem.Create(ctx, header, ciReq)
	So(err, ShouldBeNil)
	So(ciResp, ShouldNotBeNil)
	So(ciResp.Id, ShouldNotEqual, uint32(0))
	rm.AddConfigItem(appId, ciResp.Id)

	// upload content
	ctx, header = cases.GenApiCtxHeader()
	header.Set(constant.ContentIDHeaderKey, signature)
	resp, err := cli.Content.Upload(ctx, header, cases.TBizID, appId, content)
	So(err, ShouldBeNil)
	So(resp, ShouldNotBeNil)

	// create content
	contReq := &pbcs.CreateContentReq{
		BizId:        cases.TBizID,
		AppId:        appId,
		ConfigItemId: ciResp.Id,
		Sign:         signature,
		ByteSize:     size,
	}
	ctx, header = cases.GenApiCtxHeader()
	contResp, err := cli.Content.Create(ctx, header, contReq)
	So(err, ShouldBeNil)
	So(contResp, ShouldNotBeNil)
	So(contResp.Id, ShouldNotEqual, uint32(0))
	rm.AddContent(ciResp.Id, contResp.Id)

	// create commit
	cmReq := &pbcs.CreateCommitReq{
		BizId:        cases.TBizID,
		AppId:        appId,
		ConfigItemId: ciResp.Id,
		ContentId:    contResp.Id,
	}
	ctx, header = cases.GenApiCtxHeader()
	cmResp, err := cli.Commit.Create(ctx, header, cmReq)
	So(err, ShouldBeNil)
	So(cmResp, ShouldNotBeNil)
	So(cmResp.Id, ShouldNotEqual, uint32(0))
	rm.AddCommit(contResp.Id, cmResp.Id)

	// create release
	relReq := &pbcs.CreateReleaseReq{
		BizId: cases.TBizID,
		AppId: appId,
		Name:  cases.RandName(name),
	}
	ctx, header = cases.GenApiCtxHeader()
	relResp, err := cli.Release.Create(ctx, header, relReq)
	So(err, ShouldBeNil)
	So(relResp, ShouldNotBeNil)
	So(relResp.Id, ShouldNotEqual, uint32(0))
	releaseId = relResp.Id
	rm.AddRelease(appId, releaseId)

	// create strategy set
	stgSetReq := &pbcs.CreateStrategySetReq{
		BizId: cases.TBizID,
		AppId: appId,
		Name:  cases.RandName(name),
	}
	ctx, header = cases.GenApiCtxHeader()
	stgSetResp, err := cli.StrategySet.Create(ctx, header, stgSetReq)
	So(err, ShouldBeNil)
	So(stgSetResp, ShouldNotBeNil)
	So(stgSetResp.Id, ShouldNotEqual, uint32(0))
	strategySetId = stgSetResp.Id
	rm.AddStrategySet(appId, strategySetId)

	return appId, releaseId, strategySetId
}
