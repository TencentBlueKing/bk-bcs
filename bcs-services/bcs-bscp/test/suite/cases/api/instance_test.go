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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestInstance(t *testing.T) {

	var (
		cli *api.Client

		appId uint32
		relId uint32 // release id
	)

	Convey("Prepare For Commit Test", t, func() {
		cli = suite.GetClient().ApiClient
		appId, relId = rm.GetAppToRelease()
		So(appId, ShouldNotEqual, uint32(0))
		So(relId, ShouldNotEqual, uint32(0))
	})

	Convey("Publish Instance Test", t, func() {
		Convey("1.publish_instance normal test", func() {
			// test cases
			reqs := make([]pbcs.PublishInstanceReq, 0)

			r := pbcs.PublishInstanceReq{
				BizId:     cases.TBizID,
				AppId:     appId,
				ReleaseId: relId,
			}

			uids := []string{
				// to test: Only English, numbers, underline, and hyphen are allowed
				"English_-123",
				// to test: maximum length is 64
				cases.RandString(64),
				// to test: start and end with English
				cases.TEnglish,
				// to test: start and end with Number
				cases.TNumber,
			}
			// add uid field test case
			for _, uid := range uids {
				r.Uid = uid
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Memo = memo
				r.Uid = uuid.UUID()
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Instance.Publish(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// verify by list
				listReq, err := cases.GenListInstancePublishByIdsReq(cases.TBizID, []uint32{resp.Id})
				So(err, ShouldBeNil)
				ctx, header = cases.GenApiCtxHeader()
				listResp, err := cli.Instance.List(ctx, header, listReq)
				So(err, ShouldBeNil)
				So(listResp, ShouldNotBeNil)
				So(len(listResp.Details), ShouldEqual, 1)

				one := listResp.Details[0]
				So(one, ShouldNotBeNil)
				So(one.Id, ShouldEqual, resp.Id)

				So(one.Spec, ShouldNotBeNil)
				So(one.Spec.Uid, ShouldEqual, req.Uid)
				So(one.Spec.Memo, ShouldEqual, req.Memo)
				So(one.Spec.ReleaseId, ShouldEqual, req.ReleaseId)

				So(one.Attachment, ShouldNotBeNil)
				So(one.Attachment.BizId, ShouldEqual, cases.TBizID)
				So(one.Attachment.AppId, ShouldEqual, appId)
				So(one.Revision, cases.SoCreateRevision)

				rm.AddInstance(appId, resp.Id)
			}
		})

		Convey("2.publish_instance abnormal test", func() {
			// test cases
			reqs := make([]pbcs.PublishInstanceReq, 0)

			r := pbcs.PublishInstanceReq{
				BizId:     cases.TBizID,
				AppId:     appId,
				ReleaseId: relId,
			}

			uids := []string{
				// to test: Only English, numbers, underline, and hyphen are allowed
				cases.WCharacter,
				// to test: maximum length is 64
				cases.RandString(65),
				// to test: start and end with English and Number
				cases.WPrefix,
				// to test: start and end with English and Number
				cases.WTail,
			}
			// add uid field test case
			for _, uid := range uids {
				r.Uid = uid
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Memo = memo
				r.Uid = uuid.UUID()
				reqs = append(reqs, r)
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Instance.Publish(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("List Instance Publish Test", t, func() {
		// The normal list_strategy_publish_history is test by the publish_with_strategy case,
		// so we just test list_strategy_publish_history normal test on count page in here.
		Convey("1.list_strategy_publish_history normal test: count page", func() {
			publishId := rm.GetInstance(appId)
			So(publishId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{publishId})
			So(err, ShouldBeNil)

			req := &pbcs.ListPublishedInstanceReq{
				BizId:  cases.TBizID,
				Filter: filter,
				Page:   cases.CountPage(),
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Instance.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_strategy_publish_history abnormal test", func() {
			publishId := rm.GetInstance(appId)
			So(publishId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{publishId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListPublishedInstanceReq{
				{ // biz_id is invalid
					BizId:  cases.WID,
					Filter: filter,
					Page:   cases.CountPage(),
				},
				{ // filter is invalid
					BizId:  cases.TBizID,
					Filter: nil,
					Page:   cases.CountPage(),
				},
				{ // page is invalid
					BizId:  cases.TBizID,
					Filter: filter,
					Page:   nil,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Instance.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Delete Instance Publish Test", t, func() {
		publishId := rm.GetInstance(appId)

		req := &pbcs.DeletePublishedInstanceReq{
			Id:    publishId,
			BizId: cases.TBizID,
			AppId: appId,
		}

		ctx, header := cases.GenApiCtxHeader()
		resp, err := cli.Instance.Delete(ctx, header, req)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
}
