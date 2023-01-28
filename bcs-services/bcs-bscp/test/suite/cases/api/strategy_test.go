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

package api

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	pbcs "bscp.io/pkg/protocol/config-server"
	strategy "bscp.io/pkg/protocol/core/strategy"
	"bscp.io/pkg/tools"
	"bscp.io/test/client/api"
	"bscp.io/test/suite"
	"bscp.io/test/suite/cases"
)

func TestStrategy(t *testing.T) {

	var (
		cli     *api.Client
		ctx     context.Context
		header  http.Header
		preName string

		nmAppId    uint32                  // normal mode's app id
		nmRelId    uint32                  // normal mode's release id
		nmStgSetId uint32                  // normal mode's strategy set id
		nmScope    *strategy.ScopeSelector // normal mode's scope

		nsAppId    uint32                  // namespace mode's app id
		nsRelId    uint32                  // namespace mode's release id
		nsStgSetId uint32                  // namespace mode's strategy set id
		nsScope    *strategy.ScopeSelector // namespace mode's scope
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

		// generate scope
		var err error
		nmScope, err = cases.GenNormalStrategyScope(preName, nmRelId)
		So(err, ShouldBeNil)
		nsScope, err = cases.GenNamespaceStrategyScope(preName, nsRelId)
		So(err, ShouldBeNil)
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
				Scope:         nsScope,
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
				So(resp.Code, ShouldEqual, errf.OK)
				So(resp.Data, ShouldNotBeNil)
				So(resp.Data.Id, ShouldNotEqual, uint32(0))

				// verify by list_config_item
				listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nsAppId, []uint32{resp.Data.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Strategy.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)
				So(listResp.Code, ShouldEqual, errf.OK)
				So(listResp.Data, ShouldNotBeNil)

				So(len(listResp.Data.Details), ShouldEqual, 1)
				one := listResp.Data.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Data.Id)

				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
				So(one.Spec.Namespace, ShouldEqual, req.Namespace)
				So(one.Spec.ReleaseId, ShouldEqual, req.ReleaseId)
				So(one.Spec.AsDefault, ShouldEqual, req.AsDefault)
				So(one.Spec.Mode, ShouldEqual, string(table.Namespace))

				So(one.Revision, cases.SoRevision)
				So(one.Spec.Scope.Selector, ShouldNotBeNil)
				So(one.Spec.Scope.SubStrategy, cases.SoShouldJsonEqual, req.Scope.SubStrategy)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.BizId, ShouldEqual, req.BizId)
				So(one.Attachment.AppId, ShouldEqual, req.AppId)

				So(one.State, ShouldNotBeNil)
				So(one.State.PubState, ShouldNotBeBlank)

				rm.AddStrategy(nsAppId, req.StrategySetId, resp.Data.Id)
			}
		})

		Convey("2.create_strategy normal test in normal mode", func() {
			req := &pbcs.CreateStrategyReq{
				BizId:         cases.TBizID,
				AppId:         nmAppId,
				StrategySetId: nmStgSetId,
				ReleaseId:     nmRelId,
				AsDefault:     false,
				Scope:         nmScope,
				Namespace:     "",
				Name:          cases.RandName(preName),
			}

			ctx, header = cases.GenApiCtxHeader()
			resp, err := cli.Strategy.Create(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, errf.OK)
			So(resp.Data, ShouldNotBeNil)
			So(resp.Data.Id, ShouldNotEqual, uint32(0))

			// verify by list_config_item
			listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nmAppId, []uint32{resp.Data.Id})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.Strategy.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(listResp.Code, ShouldEqual, errf.OK)
			So(listResp.Data, ShouldNotBeNil)

			So(len(listResp.Data.Details), ShouldEqual, 1)
			one := listResp.Data.Details[0]
			So(one, ShouldNotBeNil)
			So(one.Spec.Scope.Selector, cases.SoShouldJsonEqual, req.Scope.Selector)
			So(one.Spec.Scope.SubStrategy, cases.SoShouldJsonEqual, req.Scope.SubStrategy)

			rm.AddStrategy(nmAppId, req.StrategySetId, resp.Data.Id)
		})

		Convey("3.create_strategy normal test: as default is true", func() {
			req := &pbcs.CreateStrategyReq{
				BizId:         cases.TBizID,
				AppId:         nmAppId,
				StrategySetId: nmStgSetId,
				ReleaseId:     nmRelId,
				AsDefault:     true,
				Scope:         nil,
				Namespace:     "",
				Name:          cases.RandName(preName),
			}
			ctx, header = cases.GenApiCtxHeader()
			resp, err := cli.Strategy.Create(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, errf.OK)
			So(resp.Data, ShouldNotBeNil)
			So(resp.Data.Id, ShouldNotEqual, uint32(0))

			// verify by list_config_item
			listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nmAppId, []uint32{resp.Data.Id})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.Strategy.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(listResp.Code, ShouldEqual, errf.OK)
			So(listResp.Data, ShouldNotBeNil)

			So(len(listResp.Data.Details), ShouldEqual, 1)
			one := listResp.Data.Details[0]
			So(one, ShouldNotBeNil)
			So(one.Spec.AsDefault, ShouldBeTrue)

			rm.AddStrategy(nmAppId, req.StrategySetId, resp.Data.Id)

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
				Scope:         nsScope,
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
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldNotEqual, errf.OK)
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
						Scope:         nsScope,
						Namespace:     cases.RandName(preName),
						Name:          cases.RandName(preName),
					}
					ctx, header = cases.GenApiCtxHeader()
					resp, err := cli.Strategy.Create(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Code, ShouldEqual, errf.OK)
				}

				// try to create an out of limit strategy
				req := &pbcs.CreateStrategyReq{
					BizId:         cases.TBizID,
					AppId:         nsAppId,
					StrategySetId: nsStgSetId,
					ReleaseId:     nsRelId,
					AsDefault:     false,
					Scope:         nsScope,
					Namespace:     cases.RandName(preName),
					Name:          cases.RandName(preName),
				}
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Create(ctx, header, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldNotEqual, errf.OK)
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
						Scope:         nmScope,
						Name:          cases.RandName(preName),
					}
					ctx, header = cases.GenApiCtxHeader()
					resp, err := cli.Strategy.Create(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Code, ShouldEqual, errf.OK)
				}

				// try to create an out of limit strategy
				req := &pbcs.CreateStrategyReq{
					BizId:         cases.TBizID,
					AppId:         nmAppId,
					StrategySetId: nmStgSetId,
					ReleaseId:     nmRelId,
					AsDefault:     false,
					Scope:         nmScope,
					Name:          cases.RandName(preName),
				}
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.Strategy.Create(ctx, header, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldNotEqual, errf.OK)
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
			pbSelector, err := selector.MarshalPB()
			So(err, ShouldBeNil)
			newScope, err := cases.GenNamespaceStrategyScope(preName, nsRelId)
			So(err, ShouldBeNil)
			newScope.SubStrategy.Spec.Scope.Selector = pbSelector

			// test cases
			reqs := make([]pbcs.UpdateStrategyReq, 0)

			r := pbcs.UpdateStrategyReq{
				BizId:     cases.TBizID,
				AppId:     nsAppId,
				Id:        stgId,
				ReleaseId: nsRelId,
				AsDefault: false,
				Scope:     nsScope,
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
				So(resp.Code, ShouldEqual, errf.OK)

				// verify by list_config_item
				listReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, nsAppId, []uint32{req.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Strategy.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)
				So(listResp.Code, ShouldEqual, errf.OK)
				So(listResp.Data, ShouldNotBeNil)

				// just verify the filed maybe update
				So(len(listResp.Data.Details), ShouldEqual, 1)
				one := listResp.Data.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
				So(one.Spec.AsDefault, ShouldEqual, req.AsDefault)
				So(one.Spec.Scope.Selector, ShouldNotBeNil)
				So(one.Spec.Scope.SubStrategy, cases.SoShouldJsonEqual, req.Scope.SubStrategy)
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
				Scope:     nsScope,
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
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldNotEqual, errf.OK)
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
			So(resp.Code, ShouldEqual, errf.OK)

			// verify by list_strategy_set
			listReq, err := cases.GenListStrategySetByIdsReq(cases.TBizID, nmAppId, []uint32{stgId})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.StrategySet.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(listResp.Code, ShouldEqual, errf.OK)
			So(listResp.Data, ShouldNotBeNil)
			So(len(listResp.Data.Details), ShouldEqual, 0)
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
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, errf.InvalidParameter)
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
			So(resp.Code, ShouldEqual, errf.OK)
			So(resp.Data.Count, ShouldEqual, uint32(1))
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
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldNotEqual, errf.OK)
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
	So(appResp.Data, ShouldNotBeNil)
	So(appResp.Data.Id, ShouldNotEqual, uint32(0))
	appId = appResp.Data.Id
	rm.AddApp(mode, appId)

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
	}
	ctx, header = cases.GenApiCtxHeader()
	ciResp, err := cli.ConfigItem.Create(ctx, header, ciReq)
	So(err, ShouldBeNil)
	So(ciResp, ShouldNotBeNil)
	So(ciResp.Data, ShouldNotBeNil)
	So(ciResp.Data.Id, ShouldNotEqual, uint32(0))
	rm.AddConfigItem(appId, ciResp.Data.Id)

	// create content
	contReq := &pbcs.CreateContentReq{
		BizId:        cases.TBizID,
		AppId:        appId,
		ConfigItemId: ciResp.Data.Id,
		Sign:         tools.SHA256(name),
		ByteSize:     uint64(len(name)),
	}
	ctx, header = cases.GenApiCtxHeader()
	contResp, err := cli.Content.Create(ctx, header, contReq)
	So(err, ShouldBeNil)
	So(contResp, ShouldNotBeNil)
	So(contResp.Data, ShouldNotBeNil)
	So(contResp.Data.Id, ShouldNotEqual, uint32(0))
	rm.AddContent(ciResp.Data.Id, contResp.Data.Id)

	// create commit
	cmReq := &pbcs.CreateCommitReq{
		BizId:        cases.TBizID,
		AppId:        appId,
		ConfigItemId: ciResp.Data.Id,
		ContentId:    contResp.Data.Id,
	}
	ctx, header = cases.GenApiCtxHeader()
	cmResp, err := cli.Commit.Create(ctx, header, cmReq)
	So(err, ShouldBeNil)
	So(cmResp, ShouldNotBeNil)
	So(cmResp.Data, ShouldNotBeNil)
	So(cmResp.Data.Id, ShouldNotEqual, uint32(0))
	rm.AddCommit(contResp.Data.Id, cmResp.Data.Id)

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
	So(relResp.Data, ShouldNotBeNil)
	So(relResp.Data.Id, ShouldNotEqual, uint32(0))
	releaseId = relResp.Data.Id
	rm.AddCommit(contResp.Data.Id, cmResp.Data.Id)

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
	So(stgSetResp.Data, ShouldNotBeNil)
	So(stgSetResp.Data.Id, ShouldNotEqual, uint32(0))
	strategySetId = stgSetResp.Data.Id
	rm.AddStrategySet(appId, strategySetId)

	return
}
