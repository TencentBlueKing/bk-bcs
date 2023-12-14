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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestContent(t *testing.T) {

	var (
		cli *api.Client

		appId uint32
		ciId  uint32

		content   string // content of config item
		signature string // SHA256 signature of content
		size      uint64 // byte size of content
	)

	Convey("Prepare For Content Test", t, func() {
		cli = suite.GetClient().ApiClient

		appId = rm.GetApp(table.Normal)
		So(appId, ShouldNotEqual, uint32(0))
		ciId = rm.GetConfigItem(appId)
		So(ciId, ShouldNotEqual, uint32(0))

		// define content
		content = "This is content for test"
		signature = tools.SHA256(content)
		size = uint64(len(content))
	})

	Convey("Upload Content Test", t, func() {
		ctx, header := cases.GenApiCtxHeader()
		header.Set(constant.ContentIDHeaderKey, signature)
		resp, err := cli.Content.Upload(ctx, header, cases.TBizID, appId, content)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})

	Convey("Create Content Test", t, func() {
		Convey("1.create_content normal test", func() {
			// create content
			req := &pbcs.CreateContentReq{
				BizId:        cases.TBizID,
				AppId:        appId,
				ConfigItemId: ciId,
				Sign:         signature,
				ByteSize:     size,
			}
			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Content.Create(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Id, ShouldNotEqual, uint32(0))

			// due to byte_size in response is string type, so it will cause unmarshal err, we skip it for a while
			// verify by list content
			var listReq *pbcs.ListContentsReq
			listReq, err = cases.GenListContentByIdsReq(cases.TBizID, appId, []uint32{resp.Id})
			So(err, ShouldBeNil)
			ctx, header = cases.GenApiCtxHeader()
			var listResp *pbcs.ListContentsResp
			listResp, err = cli.Content.List(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(len(listResp.Details), ShouldEqual, 1)

			one := listResp.Details[0]
			So(one, ShouldNotBeNil)
			So(one.Id, ShouldEqual, resp.Id)

			So(one.Spec, ShouldNotBeNil)
			So(one.Spec.ByteSize, ShouldEqual, size)
			So(one.Spec.Signature, ShouldEqual, signature)

			So(one.Attachment, ShouldNotBeNil)
			So(one.Attachment.AppId, ShouldEqual, appId)
			So(one.Attachment.ConfigItemId, ShouldEqual, ciId)
			So(one.Attachment.BizId, ShouldEqual, cases.TBizID)

			So(one.Revision, cases.SoCreateRevision)

			rm.AddContent(ciId, resp.Id)
		})

		Convey("2.create_content abnormal test", func() {
			reqs := []*pbcs.CreateContentReq{
				{ // biz_id is invalid
					BizId:        cases.WID,
					AppId:        appId,
					ConfigItemId: ciId,
					Sign:         signature,
					ByteSize:     size,
				},
				{ // app_id is invalid
					BizId:        cases.TBizID,
					AppId:        cases.WID,
					ConfigItemId: ciId,
					Sign:         signature,
					ByteSize:     size,
				},
				{ // config_item_id is invalid
					BizId:        cases.TBizID,
					AppId:        appId,
					ConfigItemId: cases.WID,
					Sign:         signature,
					ByteSize:     size,
				},
				{ // sign is invalid
					BizId:        cases.TBizID,
					AppId:        appId,
					ConfigItemId: ciId,
					Sign:         "signature",
					ByteSize:     size,
				},
				{ // byte size is invalid
					BizId:        cases.TBizID,
					AppId:        appId,
					ConfigItemId: ciId,
					Sign:         signature,
					ByteSize:     0,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Content.Create(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("List Content Test", t, func() {
		// The normal list_content is test by the first create_content case,
		// so we just test list_content normal test on count page in here.
		Convey("1.list_content normal test: count page", func() {
			ctId := rm.GetContent(ciId)
			So(ctId, ShouldNotEqual, uint32(0))

			req, err := cases.GenListContentByIdsReq(cases.TBizID, appId, []uint32{ctId})
			So(err, ShouldBeNil)
			req.Page = cases.CountPage()

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Content.List(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, 1)
		})

		Convey("2.list_content abnormal test", func() {
			ctId := rm.GetContent(ciId)
			So(ctId, ShouldNotEqual, uint32(0))

			filter, err := cases.GenQueryFilterByIds([]uint32{ctId})
			So(err, ShouldBeNil)

			reqs := []*pbcs.ListContentsReq{
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
				resp, err := cli.Content.List(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}
