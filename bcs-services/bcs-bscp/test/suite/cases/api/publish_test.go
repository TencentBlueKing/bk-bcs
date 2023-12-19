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

	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestPublish(t *testing.T) {

	var (
		cli *api.Client

		appId uint32
		stgId uint32 // strategy id
	)

	Convey("Prepare For Publish Test", t, func() {
		cli = suite.GetClient().ApiClient
		appId, stgId = rm.GetAppToStrategy()
		So(appId, ShouldNotEqual, uint32(0))
		So(stgId, ShouldNotEqual, uint32(0))
	})

	Convey("Publish With Strategy And Finish Publish Strategy Test", t, func() {
		Convey("1.publish_with_strategy and finish_publish_strategy normal test", func() {
			// start publishing with strategy
			req := &pbcs.PublishReq{
				BizId: cases.TBizID,
				AppId: appId,
			}
			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Publish.PublishWithStrategy(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Id, ShouldNotEqual, uint32(0))

			// finish publishing with strategy
			finishReq := &pbcs.FinishPublishReq{
				BizId: cases.TBizID,
				AppId: appId,
				Id:    stgId,
			}
			ctx, header = cases.GenApiCtxHeader()
			finishResp, err := cli.Publish.FinishPublishWithStrategy(ctx, header, finishReq)
			So(err, ShouldBeNil)
			So(finishResp, ShouldNotBeNil)

			// verify by list
			listReq, err := cases.GenListStrategyPublishByIdsReq(cases.TBizID, appId, []uint32{resp.Id})
			So(err, ShouldBeNil)
			ctx, header = cases.GenApiCtxHeader()
			listResp, err := cli.Publish.ListStrategyPublishHistory(ctx, header, listReq)
			So(err, ShouldBeNil)
			So(listResp, ShouldNotBeNil)
			So(len(listResp.Details), ShouldEqual, 1)
			oneHistory := listResp.Details[0]

			// get strategy detail
			listStgReq, err := cases.GenListStrategyByIdsReq(cases.TBizID, appId, []uint32{stgId})
			So(err, ShouldBeNil)
			So(err, ShouldBeNil)
			ctx, header = cases.GenApiCtxHeader()
			listStgResp, err := cli.Strategy.List(ctx, header, listStgReq)
			So(err, ShouldBeNil)
			So(listStgResp, ShouldNotBeNil)
			So(len(listStgResp.Details), ShouldEqual, 1)
			oneStg := listStgResp.Details[0]

			So(oneHistory.State, ShouldNotBeNil)
			So(oneHistory.State.PubState, ShouldBeBlank)

			So(oneHistory.Id, ShouldEqual, resp.Id)
			So(oneHistory.StrategyId, ShouldEqual, stgId)
			So(oneHistory.Spec, ShouldNotBeNil)
			So(oneHistory.Spec.Name, ShouldEqual, oneStg.Spec.Name)
			So(oneHistory.Spec.Memo, ShouldEqual, oneStg.Spec.Memo)
			So(oneHistory.Spec.Namespace, ShouldEqual, oneStg.Spec.Namespace)
			So(oneHistory.Spec.ReleaseId, ShouldEqual, oneStg.Spec.ReleaseId)
			So(oneHistory.Spec.AsDefault, ShouldEqual, oneStg.Spec.AsDefault)

			So(oneHistory.Revision, cases.SoCreateRevision)
			So(oneHistory.Attachment, ShouldNotBeNil)
			So(oneHistory.Attachment.AppId, ShouldEqual, appId)
			So(oneHistory.Attachment.BizId, ShouldEqual, cases.TBizID)
			So(oneHistory.Attachment.StrategySetId, ShouldEqual, oneStg.Attachment.StrategySetId)

			So(oneHistory.StrategyId, ShouldNotBeNil)

			rm.AddPublish(appId, resp.Id)
		})

		Convey("2.publish_with_strategy abnormal test", func() {
			// test cases
			reqs := []*pbcs.PublishReq{
				{ // biz_id is invalid
					BizId: cases.WID,
					AppId: appId,
				},
				{ // app_id is invalid
					BizId: cases.TBizID,
					AppId: cases.WID,
				},
				{ // strategy_id is invalid
					BizId: cases.TBizID,
					AppId: appId,
				},
				{ // strategy_id is not exist
					BizId: cases.TBizID,
					AppId: appId,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.Publish.PublishWithStrategy(ctx, header, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
			}
		})

		Convey("3.publish_with_strategy abnormal test: don't finish a publish and start other publish", func() {
			req := &pbcs.PublishReq{
				BizId: cases.TBizID,
				AppId: appId,
			}

			// start first publish
			ctx, header := cases.GenApiCtxHeader()
			firstResp, err := cli.Publish.PublishWithStrategy(ctx, header, req)
			So(err, ShouldBeNil)
			So(firstResp, ShouldNotBeNil)
			So(firstResp.Id, ShouldNotEqual, uint32(0))

			// start second publish
			ctx, header = cases.GenApiCtxHeader()
			secondResp, err := cli.Publish.PublishWithStrategy(ctx, header, req)
			So(err, ShouldBeNil)
			So(secondResp, ShouldNotBeNil)

			// finish publish with strategy
			finishReq := &pbcs.FinishPublishReq{
				BizId: cases.TBizID,
				AppId: appId,
				Id:    stgId,
			}
			ctx, header = cases.GenApiCtxHeader()
			finishResp, err := cli.Publish.FinishPublishWithStrategy(ctx, header, finishReq)
			So(err, ShouldBeNil)
			So(finishResp, ShouldNotBeNil)

			// try a new strategy publish after finishing
			ctx, header = cases.GenApiCtxHeader()
			secondResp, err = cli.Publish.PublishWithStrategy(ctx, header, req)
			So(err, ShouldBeNil)
			So(secondResp, ShouldNotBeNil)
			So(firstResp.Id, ShouldNotEqual, uint32(0))

			// finish upon publish
			ctx, header = cases.GenApiCtxHeader()
			finishResp, err = cli.Publish.FinishPublishWithStrategy(ctx, header, finishReq)
			So(err, ShouldBeNil)
			So(finishResp, ShouldNotBeNil)

		})

		Convey("4.finish_publish_strategy abnormal test", func() {
			// create a publish_with_strategy for test
			pubReq := &pbcs.PublishReq{
				BizId: cases.TBizID,
				AppId: appId,
			}
			ctx, header := cases.GenApiCtxHeader()
			pubResp, err := cli.Publish.PublishWithStrategy(ctx, header, pubReq)
			So(err, ShouldBeNil)
			So(pubResp, ShouldNotBeNil)
			So(pubResp.Id, ShouldNotEqual, uint32(0))

			reqs := []*pbcs.FinishPublishReq{
				{ // biz_id is invalid
					BizId: cases.WID,
					AppId: appId,
					Id:    stgId,
				},
				{ // app_id is invalid
					BizId: cases.TBizID,
					AppId: cases.WID,
					Id:    stgId,
				},
				{ // strategy_id is invalid
					BizId: cases.TBizID,
					AppId: appId,
					Id:    cases.WID,
				},
			}

			for _, req := range reqs {
				ctx, header = cases.GenApiCtxHeader()
				finishResp, err := cli.Publish.FinishPublishWithStrategy(ctx, header, req)
				So(err, ShouldBeNil)
				So(finishResp, ShouldNotBeNil)
			}
		})
	})

	Convey("List Strategy Publish History Test", t, func() {
		// The normal list_strategy_publish_history is test by the publish_with_strategy case,
		// so we just test list_strategy_publish_history normal test on count page in here.
		Convey("1.list_strategy_publish_history normal test: count page", func() {
			publishId := rm.GetPublish(appId)
			So(publishId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{publishId})
			So(err, ShouldBeNil)

			req := &pbcs.ListPubStrategyHistoriesReq{
				BizId:  cases.TBizID,
				AppId:  appId,
				Filter: filter,
				Page:   cases.CountPage(),
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.Publish.ListStrategyPublishHistory(ctx, header, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Count, ShouldEqual, uint32(1))
		})

		Convey("2.list_strategy_publish_history abnormal test", func() {
			publishId := rm.GetPublish(appId)
			So(publishId, ShouldNotEqual, uint32(0))
			filter, err := cases.GenQueryFilterByIds([]uint32{publishId})
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
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
			}
		})
	})
}
