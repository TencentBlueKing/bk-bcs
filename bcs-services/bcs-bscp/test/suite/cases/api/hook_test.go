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

func TestHook(t *testing.T) {

	var (
		cli *api.Client

		preName string
		appId   uint32

		preType  string
		preHook  string
		postType string
		postHook string
	)

	Convey("Prepare For Hook Test", t, func() {
		cli = suite.GetClient().ApiClient
		appId = rm.GetApp(table.Normal)
		So(appId, ShouldNotEqual, uint32(0))
		preName = "hook"

		preType = "shell"
		preHook = "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n"
		postType = "python"
		postHook = "from datetime import datetime\nprint(\"hello, end at\", datetime.now())\n"
	})

	Convey("Create Hook Test", t, func() {

		Convey("1.create_hook normal test", func() {
			// test cases
			reqs := make([]pbcs.CreateHookReq, 0)

			r := pbcs.CreateHookReq{
				AppId:    appId,
				Name:     cases.RandName(preName),
				PreType:  preType,
				PreHook:  preHook,
				PostType: postType,
				PostHook: postHook,
			}

			// add name field test case
			names := genNormalNameForCreateTest()
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// hook type is shell
			r.Name = cases.RandNameN(preName, 64)
			r.PreType = string(table.Shell)
			reqs = append(reqs, r)

			// hook type is python
			r.Name = cases.RandNameN(preName, 64)
			r.PreType = string(table.Python)
			reqs = append(reqs, r)

			// An application can only create one hook when the releaseID is specified, so we must ensure that the
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
				resp, err := cli.Hook.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// verify by list_hook
				listReq, err := cases.GenListHookByIdsReq(appId, []uint32{resp.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Hook.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Id)
				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.PreType, ShouldEqual, req.PreType)
				So(one.Spec.PreHook, ShouldEqual, req.PreHook)
				So(one.Spec.PostType, ShouldEqual, req.PostType)
				So(one.Spec.PostHook, ShouldEqual, req.PostHook)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.AppId, ShouldEqual, appId)
				So(one.Attachment.BizId, ShouldEqual, cases.TBizID)
				So(one.Attachment.ReleaseId, ShouldEqual, req.ReleaseId)

				rm.AddHook(appId, resp.Id)
			}
		})

		Convey("2.create_hook abnormal test.", func() {
			// test cases
			reqs := make([]pbcs.CreateHookReq, 0)

			r := pbcs.CreateHookReq{
				AppId:    appId,
				Name:     cases.RandName(preName),
				PreType:  preType,
				PreHook:  preHook,
				PostType: postType,
				PostHook: postHook,
			}

			// add preName field test case
			names := genAbnormalNameForTest()
			// to test: maximum length is 65
			names[1] = cases.RandNameN(preName, 65)
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// app id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// HookType is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.PreType = cases.WEnum
			reqs = append(reqs, r)
			r.PreType = string(table.Shell)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Hook.Create(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Update Hook Test", t, func() {
		hId := rm.GetHook(appId)
		So(hId, ShouldNotEqual, uint32(0))

		Convey("1.update_hook normal test", func() {
			// test cases
			reqs := make([]pbcs.UpdateHookReq, 0)

			r := pbcs.UpdateHookReq{
				AppId:    appId,
				HookId:   hId,
				Name:     cases.RandName(preName),
				PreType:  preType,
				PreHook:  preHook,
				PostType: postType,
				PostHook: postHook,
			}

			// add name field test case
			names := genNormalNameForUpdateTest()
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// hook type is shell
			r.Name = cases.RandNameN(preName, 64)
			r.PreType = string(table.Shell)
			reqs = append(reqs, r)

			// hook type is python
			r.Name = cases.RandNameN(preName, 64)
			r.PreType = string(table.Python)
			reqs = append(reqs, r)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Hook.Update(ctx, header, &req)
				So(err, ShouldBeNil)
				So(*resp, ShouldBeZeroValue)

				// verify by list_hook
				listReq, err := cases.GenListHookByIdsReq(appId, []uint32{hId})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Hook.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)

				So(len(listResp.Details), ShouldEqual, 1)
				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, hId)
				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Name, ShouldEqual, req.Name)
				So(one.Spec.PreType, ShouldEqual, req.PreType)
				So(one.Spec.PreHook, ShouldEqual, req.PreHook)
				So(one.Spec.PostType, ShouldEqual, req.PostType)
				So(one.Spec.PostHook, ShouldEqual, req.PostHook)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.AppId, ShouldEqual, appId)
				So(one.Attachment.BizId, ShouldEqual, cases.TBizID)
				So(one.Attachment.ReleaseId, ShouldEqual, req.ReleaseId)
			}
		})

		Convey("2.update_hook abnormal test", func() {
			// test cases
			reqs := make([]pbcs.UpdateHookReq, 0)

			r := pbcs.UpdateHookReq{
				AppId:    appId,
				Name:     cases.RandName(preName),
				PreType:  preType,
				PreHook:  preHook,
				PostType: postType,
				PostHook: postHook,
			}

			// add name field test case
			names := []string{
				// to test: 仅允许使用英文、数字、下划线、中划线
				cases.WCharacter,
				// to test: maximum length is 65
				cases.RandNameN(preName, 64),
				// to test: 必须以英文、数字开头
				cases.WPrefix,
				// to test: 必须以英文、数字结尾
				cases.WTail,
			}
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// app id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// HookType is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.PreType = cases.WEnum
			reqs = append(reqs, r)
			r.PreType = string(table.Shell)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Hook.Update(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
	Convey("List Hook Test", t, func() {
		// The normal list_hook is test by the first create_hook case,
		// so we just test list_hook normal test on count page in here
		Convey("1.list_hook normal test: count page", func() {
			// get a hook for list
			hId := rm.GetHook(appId)
			So(hId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{hId})
			So(err, ShouldBeNil)

			req := &pbcs.ListHooksReq{
				AppId:  appId,
				Filter: filter,
				Page:   cases.ListPage(),
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Hook.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_hook abnormal test", func() {
			// get a hook for list
			hId := rm.GetHook(appId)
			So(hId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{hId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListHooksReq{
				{ // app_id is invalid
					AppId:  cases.WID,
					Filter: filter,
					Page:   cases.ListPage(),
				},
				{ // filter is invalid
					AppId:  appId,
					Filter: nil,
					Page:   cases.ListPage(),
				},
				{ // page is invalid
					AppId:  appId,
					Filter: filter,
					Page:   nil,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Hook.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Delete Hook Test", t, func() {
		Convey("1.delete_hook normal test", func() {
			// get a hook for delete
			hId := rm.GetHook(appId)
			So(hId, ShouldNotEqual, uint32(0))
			defer rm.DeleteHook(appId, hId)

			req := &pbcs.DeleteHookReq{
				HookId: hId,
				AppId:  appId,
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Hook.Delete(ctx, header, req)
			So(err, ShouldBeNil)
			So(*resp, ShouldBeZeroValue)

			// verify by list
			listReq, err := cases.GenListHookByIdsReq(appId, []uint32{hId})
			So(err, ShouldBeNil)
			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.Hook.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(len(listResp.Details), ShouldEqual, 0)
		})

		Convey("2.delete_hook abnormal test", func() {
			// get a hook for delete
			hId := rm.GetHook(appId)

			reqs := []*pbcs.DeleteHookReq{
				{ // hook id is invalid
					HookId: cases.WID,
					AppId:  appId,
				},
				{ // app id is invalid
					HookId: hId,
					AppId:  cases.WID,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Hook.Delete(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}
