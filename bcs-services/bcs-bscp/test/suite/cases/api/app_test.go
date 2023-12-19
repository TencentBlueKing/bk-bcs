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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestApplication(t *testing.T) {

	cli := suite.GetClient().ApiClient

	Convey("Create App Test", t, func() {

		preName := "create_app"

		Convey("1.create_app normal test", func() {
			// test cases
			reqs := make([]pbcs.CreateAppReq, 0)

			// mode is normal
			r := pbcs.CreateAppReq{
				BizId:          cases.TBizID,
				Name:           cases.RandName(preName),
				ConfigType:     string(table.File),
				Mode:           string(table.Normal),
				ReloadType:     string(table.ReloadWithFile),
				ReloadFilePath: "/tmp/reload.json",
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

			// mode is namespace
			r.Name = cases.RandName(preName)
			r.Mode = string(table.Namespace)
			reqs = append(reqs, r)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.App.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// verify by list_app
				listReq, err := cases.GenListAppByIdsReq(req.BizId, []uint32{resp.Id})
				So(err, ShouldBeNil)

				ctxList, headerList := cases.GenApiCtxHeader()
				listResp, err := cli.App.List(ctxList, headerList, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(listResp.Count, ShouldEqual, uint32(1))

				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Id)
				So(one.BizId, ShouldEqual, req.BizId)

				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Mode, ShouldEqual, req.Mode)
				So(one.Spec.Memo, ShouldEqual, req.Memo)

				So(one.Revision, cases.SoRevision)

				// save app
				rm.AddApp(table.AppMode(req.Mode), resp.Id)
			}
		})

		Convey("2.create_app abnormal test", func() {
			// test cases
			reqs := make([]pbcs.CreateAppReq, 0)

			r := pbcs.CreateAppReq{
				BizId:          cases.TBizID,
				Name:           cases.RandName(preName),
				ConfigType:     string(table.File),
				Mode:           string(table.Normal),
				ReloadType:     string(table.ReloadWithFile),
				ReloadFilePath: "/tmp/reload.json",
			}

			// biz_id is invalid
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// config_type is invalid
			r.ConfigType = cases.WEnum
			r.Name = cases.RandName(preName)
			reqs = append(reqs, r)
			r.ConfigType = string(table.File)

			// mode is invalid
			r.Mode = cases.WEnum
			r.Name = cases.RandName(preName)
			reqs = append(reqs, r)
			r.Mode = string(table.Normal)

			// reload_type is invalid.
			r.ReloadType = cases.WCharacter
			r.Name = cases.RandName(preName)
			reqs = append(reqs, r)
			r.ReloadType = string(table.ReloadWithFile)

			// reload_file_path not abs path.
			r.ReloadFilePath = "reload.yaml"
			r.Name = cases.RandName(preName)
			reqs = append(reqs, r)
			r.ReloadFilePath = "/tmp/reload.json"

			// add name field test case
			names := genAbnormalNameForTest()
			for _, name := range names {
				r.Name = name
				reqs = append(reqs, r)
			}
			r.Name = cases.RandName(preName)

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandName(preName)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.App.Create(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Update App Test", t, func() {

		preName := "update_app"

		// get an app for test
		appId := rm.GetApp(table.Normal)

		Convey("1.update_app normal test", func() {
			// test cases
			reqs := make([]pbcs.UpdateAppReq, 0)

			// mode is namespace
			r := pbcs.UpdateAppReq{
				BizId: cases.TBizID,
				Id:    appId,
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
				resp, err := cli.App.Update(ctx, header, &req)
				So(err, ShouldBeNil)
				So(*resp, ShouldBeZeroValue)

				// verify by list_app
				listReq, err := cases.GenListAppByIdsReq(req.BizId, []uint32{appId})
				So(err, ShouldBeNil)

				ctxList, headerList := cases.GenApiCtxHeader()
				listResp, err := cli.App.List(ctxList, headerList, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(listResp.Count, ShouldEqual, uint32(1))
				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
			}
		})

		Convey("2.update_app abnormal test", func() {
			// test cases
			reqs := make([]pbcs.UpdateAppReq, 0)

			r := pbcs.UpdateAppReq{
				BizId: cases.TBizID,
				Name:  cases.RandName(preName),
			}

			// biz_id is invalid
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

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
				resp, err := cli.App.Update(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Delete App Test", t, func() {
		Convey("1.delete_app normal test", func() {
			// get an app for test
			appId := rm.GetApp(table.Normal)
			defer rm.DeleteApp(table.Normal, appId)

			req := &pbcs.DeleteAppReq{
				BizId: cases.TBizID,
				Id:    appId,
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.App.Delete(ctx, header, req)
			So(err, ShouldBeNil)
			So(*resp, ShouldBeZeroValue)

			// verify by list_app
			listReq, err := cases.GenListAppByIdsReq(req.BizId, []uint32{appId})
			So(err, ShouldBeNil)

			ctxList, headerList := cases.GenApiCtxHeader()
			listResp, err := cli.App.List(ctxList, headerList, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(len(listResp.Details), ShouldEqual, 0)
		})

		Convey("2.delete_app abnormal test", func() {
			// get an app for test
			appId := rm.GetApp(table.Normal)

			reqs := []*pbcs.DeleteAppReq{
				{ // biz_id is invalid
					BizId: cases.WID,
					Id:    appId,
				},
				{ // app_id is invalid
					BizId: cases.TBizID,
					Id:    cases.WID,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.App.Delete(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})

	})

	Convey("List App Test", t, func() {
		// The normal list_app is test by the first create_app case,
		// so we just test list_app normal test on count page in here.
		Convey("1.list_app normal test: count page", func() {

			appId := rm.GetApp(table.Normal)

			filter, err := cases.GenQueryFilterByIds([]uint32{appId})
			So(err, ShouldBeNil)

			req := &pbcs.ListAppsReq{
				BizId:  cases.TBizID,
				Filter: filter,
				Page:   cases.ListPage(),
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.App.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_app abnormal test: set a invalid parameter", func() {
			appId := rm.GetApp(table.Normal)

			filter, err := cases.GenQueryFilterByIds([]uint32{appId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListAppsReq{
				{ // biz_id is invalid
					BizId:  cases.WID,
					Filter: filter,
					Page:   cases.ListPage(),
				},
				{ // filter is invalid
					BizId:  cases.TBizID,
					Filter: nil,
					Page:   cases.ListPage(),
				},
				{ // page is invalid
					BizId:  cases.TBizID,
					Filter: filter,
					Page:   nil,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.App.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}
