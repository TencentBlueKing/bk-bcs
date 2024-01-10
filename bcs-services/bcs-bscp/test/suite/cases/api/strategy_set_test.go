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
	"testing"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestStrategySet(t *testing.T) {

	var (
		cli *api.Client

		preName string
	)

	Convey("Prepare For Strategy Set Test", t, func() {
		cli = suite.GetClient().ApiClient
		preName = "test_stg_set"
	})

	Convey("Create Strategy Set Test", t, func() {
		Convey("1.create_strategy_set normal test", func() {
			// test cases
			reqs := make([]pbcs.CreateStrategySetReq, 0)

			r := pbcs.CreateStrategySetReq{
				BizId: cases.TBizID,
				Name:  cases.RandName(preName),
			}

			// add name field test case
			names := genNormalNameForCreateTest()
			for _, name := range names {
				r.Name = name
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandName(preName)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			// An application can only create one strategy set, so we must ensure that the
			// number of applications created is greater than or equal to the number of
			// test cases.If it is not satisfied with the condition, we should create enough
			// number of applications for test cases.
			numApp := len(rm.App[table.Normal])
			numCases := len(reqs)
			if numApp < numCases {
				numToCreate := numCases - numApp
				for i := 0; i < numToCreate; i++ {
					req := &pbcs.CreateAppReq{
						BizId:          cases.TBizID,
						Name:           cases.RandNameN(preName, 128),
						ConfigType:     string(table.File),
						Mode:           string(table.Normal),
						ReloadType:     string(table.ReloadWithFile),
						ReloadFilePath: "/tmp/reload.json",
					}
					ctx, header := cases.GenApiCtxHeader()
					resp, err := cli.App.Create(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Id, ShouldNotEqual, uint32(0))
					rm.AddApp(table.Normal, resp.Id)
				}
			}

			for index, req := range reqs {
				// get an app and set to test case
				appId := rm.App[table.Normal][index]
				So(appId, ShouldNotEqual, uint32(0))
				req.AppId = appId

				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.StrategySet.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// verify by list_strategy_set
				listReq, err := cases.GenListStrategySetByIdsReq(cases.TBizID, appId, []uint32{resp.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.StrategySet.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Id)

				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.BizId, ShouldEqual, req.BizId)
				So(one.Attachment.AppId, ShouldEqual, req.AppId)

				So(one.Revision, cases.SoRevision)

				So(one.State, ShouldNotBeNil)
				So(one.State.Status, ShouldEqual, string(table.Enabled))

				rm.AddStrategySet(appId, resp.Id)
			}
		})

		Convey("2.create_strategy_set abnormal test: one app must create only one strategy set", func() {
			// get an app id which has strategy set
			appId, _ := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))

			req := &pbcs.CreateStrategySetReq{
				BizId: cases.TBizID,
				AppId: appId,
				Name:  cases.RandName(preName),
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.StrategySet.Create(ctx, header, req)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})

		Convey("3.create_strategy_set abnormal test", func() {
			// create an app id which doesn't have strategy set
			appReq := &pbcs.CreateAppReq{
				BizId:          cases.TBizID,
				Name:           cases.RandName(preName),
				ConfigType:     string(table.File),
				Mode:           string(table.Normal),
				Memo:           cases.TNumber,
				ReloadType:     string(table.ReloadWithFile),
				ReloadFilePath: "/tmp/reload.json",
			}

			ctx, header := cases.GenApiCtxHeader()
			appResp, err := cli.App.Create(ctx, header, appReq)
			So(err, ShouldBeNil)
			So(appResp, ShouldNotBeNil)
			So(appResp.Id, ShouldNotEqual, uint32(0))
			appId := appResp.Id

			// test cases
			reqs := make([]pbcs.CreateStrategySetReq, 0)

			// mode is namespace
			r := pbcs.CreateStrategySetReq{
				BizId: cases.TBizID,
				AppId: appId,
				Name:  cases.RandName(preName),
			}

			// biz_id is invalid
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app id is invalid
			r.Name = cases.RandName(preName)
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// add name field test case
			names := genAbnormalNameForTest()
			for _, name := range names {
				r.Name = name
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandName(preName)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				resp, err := cli.StrategySet.Create(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Update Strategy Set Test", t, func() {
		Convey("1.update_strategy_set normal test", func() {
			appId, stgSetId := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))
			So(stgSetId, ShouldNotEqual, uint32(0))

			// test cases
			reqs := make([]pbcs.UpdateStrategySetReq, 0)

			r := pbcs.UpdateStrategySetReq{
				BizId: cases.TBizID,
				AppId: appId,
				Id:    stgSetId,
				Name:  cases.RandName(preName),
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
				r.Name = cases.RandName(preName)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.StrategySet.Update(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)

				// verify by list_strategy_set
				listReq, err := cases.GenListStrategySetByIdsReq(cases.TBizID, appId, []uint32{req.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.StrategySet.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
			}
		})

		Convey("2.update_strategy_set abnormal test", func() {
			appId, stgSetId := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))
			So(stgSetId, ShouldNotEqual, uint32(0))

			// test cases
			reqs := make([]pbcs.UpdateStrategySetReq, 0)

			// mode is namespace
			r := pbcs.UpdateStrategySetReq{
				BizId: cases.TBizID,
				AppId: appId,
				Id:    stgSetId,
				Name:  cases.RandName(preName),
			}

			// biz_id is invalid
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app id is invalid
			r.Name = cases.RandName(preName)
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// strategy set id is invalid
			r.Name = cases.RandName(preName)
			r.Id = cases.WID
			reqs = append(reqs, r)
			r.Id = stgSetId

			// add name field test case
			names := genAbnormalNameForTest()
			for _, name := range names {
				r.Name = name
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandName(preName)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.StrategySet.Update(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Delete Strategy Set Test", t, func() {
		Convey("1.delete_strategy_set normal test", func() {
			appId, stgSetId := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))
			So(stgSetId, ShouldNotEqual, uint32(0))
			defer rm.DeleteStrategySet(appId)

			req := &pbcs.DeleteStrategySetReq{
				BizId: cases.TBizID,
				AppId: appId,
				Id:    stgSetId,
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.StrategySet.Delete(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			// verify by list_strategy_set
			listReq, err := cases.GenListStrategySetByIdsReq(cases.TBizID, appId, []uint32{stgSetId})
			So(err, ShouldBeNil)

			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.StrategySet.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(len(listResp.Details), ShouldEqual, 0)
		})

		Convey("2.delete_strategy_set abnormal test", func() {
			appId, stgSetId := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))
			So(stgSetId, ShouldNotEqual, uint32(0))

			reqs := []*pbcs.DeleteStrategySetReq{
				{ // strategy set id is invalid
					Id:    cases.WID,
					BizId: cases.TBizID,
					AppId: appId,
				},
				{ // biz id is invalid
					Id:    stgSetId,
					BizId: cases.WID,
					AppId: appId,
				},
				{ // app id is invalid
					Id:    stgSetId,
					BizId: cases.TBizID,
					AppId: cases.WID,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.StrategySet.Delete(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})

	})

	Convey("List Strategy Set Test", t, func() {
		// The normal list_strategy_set is test by the create_strategy_set case,
		// so we just test list_strategy_set normal test on count page in here
		Convey("1.list_strategy_set normal test: count page", func() {
			appId, stgSetId := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))
			So(stgSetId, ShouldNotEqual, uint32(0))

			filter, err := cases.GenQueryFilterByIds([]uint32{stgSetId})
			So(err, ShouldBeNil)

			req := &pbcs.ListStrategySetsReq{
				BizId:  cases.TBizID,
				AppId:  appId,
				Filter: filter,
				Page:   cases.CountPage(),
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.StrategySet.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_strategy_set abnormal test", func() {
			appId, stgSetId := rm.GetAppToStrategySet()
			So(appId, ShouldNotEqual, uint32(0))
			So(stgSetId, ShouldNotEqual, uint32(0))

			filter, err := cases.GenQueryFilterByIds([]uint32{stgSetId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListStrategySetsReq{
				{ // biz_id is invalid
					BizId:  cases.WID,
					AppId:  appId,
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
					AppId:  appId,
					Filter: nil,
					Page:   cases.CountPage(),
				},
				{ // page is invalid
					BizId:  cases.TBizID,
					AppId:  appId,
					Filter: filter,
					Page:   nil,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.StrategySet.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}
