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
	"testing"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/tools"
	"bscp.io/test/client/api"
	"bscp.io/test/suite"
	"bscp.io/test/suite/cases"
)

func TestRelease(t *testing.T) {

	var (
		cli *api.Client

		preName string
		appId   uint32
	)

	Convey("Prepare For ReleaseID Test", t, func() {
		cli = suite.GetClient().ApiClient
		preName = "release_test"

		// the before test apps have un-commit config item,
		// which don't meet the condition of create_release.
		// So, we should create a new app, whose config item
		// is committed.

		// create app
		appReq := &pbcs.CreateAppReq{
			BizId:          cases.TBizID,
			Name:           cases.RandName(preName),
			ConfigType:     string(table.File),
			Mode:           string(table.Normal),
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

		// create config item
		ciReq := &pbcs.CreateConfigItemReq{
			BizId:     cases.TBizID,
			AppId:     appId,
			Name:      preName + "_config_item",
			Path:      "/etc",
			FileType:  string(table.Json),
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

		// create content
		content := "test_release_content"
		ctReq := &pbcs.CreateContentReq{
			BizId:        cases.TBizID,
			AppId:        appId,
			ConfigItemId: ciResp.Data.Id,
			Sign:         tools.SHA256(content),
			ByteSize:     uint64(len(content)),
		}
		ctx, header = cases.GenApiCtxHeader()
		ctResp, err := cli.Content.Create(ctx, header, ctReq)
		So(err, ShouldBeNil)
		So(ctResp, ShouldNotBeNil)
		So(ctResp.Data, ShouldNotBeNil)
		So(ctResp.Data.Id, ShouldNotEqual, uint32(0))

		// create commit
		cmReq := &pbcs.CreateCommitReq{
			BizId:        cases.TBizID,
			AppId:        appId,
			ConfigItemId: ciResp.Data.Id,
			ContentId:    ctResp.Data.Id,
		}
		ctx, header = cases.GenApiCtxHeader()
		cmResp, err := cli.Commit.Create(ctx, header, cmReq)
		So(err, ShouldBeNil)
		So(cmResp, ShouldNotBeNil)
		So(cmResp.Data, ShouldNotBeNil)
		So(cmResp.Data.Id, ShouldNotEqual, uint32(0))
	})

	Convey("Create ReleaseID Test", t, func() {

		Convey("1.create_release normal test", func() {
			// test cases
			reqs := make([]pbcs.CreateReleaseReq, 0)

			r := pbcs.CreateReleaseReq{
				BizId: cases.TBizID,
				AppId: appId,
			}

			// add preName field test case
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

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Release.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, errf.OK)

				// verify by list_config_item
				listReq, err := cases.GenListReleaseByIdsReq(cases.TBizID, appId, []uint32{resp.Data.Id})
				So(err, ShouldBeNil)

				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Release.List(ctx, header, listReq)
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

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.BizId, ShouldEqual, req.BizId)
				So(one.Attachment.AppId, ShouldEqual, req.AppId)
				So(one.Revision, cases.SoCreateRevision)

				rm.AddRelease(appId, resp.Data.Id)
			}

		})

		Convey("2.create_release abnormal test", func() {
			// test cases
			reqs := make([]pbcs.CreateReleaseReq, 0)

			r := pbcs.CreateReleaseReq{
				BizId: cases.TBizID,
				AppId: appId,
			}

			// biz id is invalid
			r.Name = cases.RandNameN(preName, 128)
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app id is invalid
			r.Name = cases.RandNameN(preName, 128)
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
				r.Name = cases.RandNameN(preName, 128)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Release.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, errf.InvalidParameter)
			}
		})
	})

	Convey("List ReleaseID Test", t, func() {
		// The normal list_release is test by the create_release case,
		// so we just test list_release normal test on count page in here.
		Convey("1.list_release normal test", func() {
			releaseId := rm.GetRelease(appId)
			So(releaseId, ShouldNotEqual, uint32(0))

			req, err := cases.GenListReleaseByIdsReq(cases.TBizID, appId, []uint32{releaseId})
			So(err, ShouldBeNil)
			req.Page = cases.CountPage()

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Release.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, errf.OK)
			So(resp.Data, ShouldNotBeNil)
			So(resp.Data.Count, ShouldEqual, 1)
		})

		Convey("2.list_release abnormal test", func() {
			releaseId := rm.GetRelease(appId)
			So(releaseId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{releaseId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListReleasesReq{
				{ // biz_id is invalid
					BizId:  cases.WID,
					AppId:  appId,
					Filter: filter,
					Page:   cases.ListPage(),
				},
				{ // app_id is invalid
					BizId:  cases.TBizID,
					AppId:  cases.WID,
					Filter: filter,
					Page:   cases.ListPage(),
				},
				{ // filter is invalid
					BizId:  cases.TBizID,
					AppId:  appId,
					Filter: nil,
					Page:   cases.ListPage(),
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
				resp, err := cli.Release.List(ctx, header, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldNotEqual, errf.OK)
			}
		})
	})
}
