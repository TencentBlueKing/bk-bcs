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
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestCommit(t *testing.T) {

	var (
		cli *api.Client

		appId  uint32
		ciId   uint32 // config item id
		ctId   uint32 // content id
		ctSign string // content signature
		ctSize uint64 // content byte size
	)

	Convey("Prepare For Commit Test", t, func() {
		cli = suite.GetClient().ApiClient
		appId = rm.GetApp(table.Normal)
		So(appId, ShouldNotEqual, uint32(0))
		ciId = rm.GetConfigItem(appId)
		So(ciId, ShouldNotEqual, uint32(0))
		ctId = rm.GetContent(ciId)
		So(ctId, ShouldNotEqual, uint32(0))

		// get content from db
		req, err := cases.GenListContentByIdsReq(cases.TBizID, appId, []uint32{ctId})
		So(err, ShouldBeNil)
		ctx, header := cases.GenApiCtxHeader()
		resp, err := cli.Content.List(ctx, header, req)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(len(resp.Details), ShouldEqual, 1)

		one := resp.Details[0]
		So(one.Spec, ShouldNotBeNil)
		ctSign = one.Spec.Signature
		ctSize = one.Spec.ByteSize
	})

	Convey("Create Commit Test", t, func() {
		Convey("1.create_commit normal test", func() {
			// test cases
			reqs := make([]pbcs.CreateCommitReq, 0)

			r := pbcs.CreateCommitReq{
				BizId:        cases.TBizID,
				AppId:        appId,
				ConfigItemId: ciId,
				ContentId:    ctId,
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Memo = memo
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Commit.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// verify by list_commit
				var listReq *pbcs.ListCommitsReq
				listReq, err = cases.GenListCommitByIdsReq(cases.TBizID, appId, []uint32{resp.Id})
				So(err, ShouldBeNil)
				ctx, header = cases.GenApiCtxHeader()
				var listResp *pbcs.ListCommitsResp
				listResp, err = cli.Commit.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)
				So(len(listResp.Details), ShouldEqual, 1)

				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Id)

				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
				So(one.Spec.ContentId, ShouldEqual, ctId)

				So(one.Spec.Content, ShouldNotBeNil)
				So(one.Spec.Content.ByteSize, ShouldEqual, ctSize)
				So(one.Spec.Content.Signature, ShouldEqual, ctSign)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.AppId, ShouldEqual, appId)
				So(one.Attachment.ConfigItemId, ShouldEqual, ciId)
				So(one.Attachment.BizId, ShouldEqual, cases.TBizID)
				So(one.Revision, cases.SoCreateRevision)

				rm.AddCommit(ctId, resp.Id)
			}
		})

		Convey("2.create_commit abnormal test", func() {
			// test cases
			reqs := make([]pbcs.CreateCommitReq, 0)

			r := pbcs.CreateCommitReq{
				BizId:        cases.TBizID,
				AppId:        appId,
				ConfigItemId: ciId,
				ContentId:    ctId,
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Memo = memo
				reqs = append(reqs, r)
			}

			// biz_id is invalid
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app_id is invalid
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// config_item_id is invalid
			r.ConfigItemId = cases.WID
			reqs = append(reqs, r)
			r.ConfigItemId = ciId

			// content_id is invalid
			r.ContentId = cases.WID
			reqs = append(reqs, r)
			r.ContentId = ctId

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Commit.Create(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("List Commit Test", t, func() {
		// The normal list_commit is test by the first create_commit case,
		// so we just test list_commit normal test on count page in here.

		// get a commit for list
		cmId := rm.GetCommit(ctId)
		So(cmId, ShouldNotEqual, uint32(0))

		Convey("1.list_commit normal test: test count page", func() {
			req, err := cases.GenListCommitByIdsReq(cases.TBizID, appId, []uint32{cmId})
			So(err, ShouldBeNil)
			req.Page = cases.CountPage()

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Commit.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_commit abnormal test", func() {
			filter, err := cases.GenQueryFilterByIds([]uint32{cmId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListCommitsReq{
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
				resp, err := cli.Commit.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}
