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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestConfigItem(t *testing.T) {

	var (
		cli *api.Client

		preName string
		appId   uint32

		content   string // content of config item
		signature string // SHA256 signature of content
		size      uint64 // byte size of content
	)

	Convey("Prepare For Config Item Test", t, func() {
		cli = suite.GetClient().ApiClient
		appId = rm.GetApp(table.Normal)
		So(appId, ShouldNotEqual, uint32(0))
		preName = "config_item"

		// define content
		content = "This is content for test"
		signature = tools.SHA256(content)
		size = uint64(len(content))
	})

	Convey("Create Config Item Test", t, func() {

		Convey("1.create_config_item normal test", func() {
			// test cases
			reqs := make([]pbcs.CreateConfigItemReq, 0)

			r := pbcs.CreateConfigItemReq{
				BizId:     cases.TBizID,
				AppId:     appId,
				Name:      cases.RandName(preName),
				Path:      "/etc",
				FileType:  string(table.Xml),
				FileMode:  string(table.Unix),
				User:      "root",
				UserGroup: "root",
				Privilege: "755",
				Sign:      signature,
				ByteSize:  size,
			}

			// add preName field test case
			names := []string{
				cases.TEnNumUnderHyphenDot,
				cases.RandNameN(preName, 64),
				cases.TEnglish,
				cases.TNumber,
			}
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandNameN(preName, 64)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			// file type is yaml
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = string(table.Yaml)
			reqs = append(reqs, r)

			// file type is binary
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = string(table.Binary)
			reqs = append(reqs, r)

			// file type is json
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = string(table.Json)
			reqs = append(reqs, r)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.ConfigItem.Create(ctx, header, &req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldNotEqual, uint32(0))

				// unsupported list query by self-defined filter using ids
				// verify by list_config_item
				//listReq, err := cases.GenListConfigItemByIdsReq(cases.TBizID, appId, []uint32{resp.Id})
				//So(err, ShouldBeNil)
				//
				//ctx, header = cases.GenApiCtxHeader()
				//listResp, err := cli.ConfigItem.List(ctx, header, listReq)
				//So(err, ShouldBeNil)
				//So(listResp, ShouldNotBeNil)
				//
				//So(len(listResp.Details), ShouldEqual, 1)
				//one := listResp.Details[0]
				//So(one, ShouldNotBeNil)
				//So(one.Id, ShouldEqual, resp.Id)
				//So(one.Spec, ShouldNotBeNil)
				//So(one.Spec.Path, ShouldEqual, req.Path)
				//So(one.Spec.Name, ShouldEqual, req.Name)
				//So(one.Spec.FileType, ShouldEqual, req.FileType)
				//So(one.Spec.FileMode, ShouldEqual, req.FileMode)
				//So(one.Spec.Permission, ShouldNotBeNil)
				//So(one.Spec.Permission.User, ShouldEqual, req.User)
				//So(one.Spec.Permission.UserGroup, ShouldEqual, req.UserGroup)
				//So(one.Spec.Permission.Privilege, ShouldEqual, req.Privilege)
				//
				//So(one.Attachment, ShouldNotBeNil)
				//So(one.Attachment.AppId, ShouldEqual, appId)
				//So(one.Attachment.BizId, ShouldEqual, cases.TBizID)

				rm.AddConfigItem(appId, resp.Id)
			}
		})

		Convey("2.create_config_item abnormal test.", func() {
			// test cases
			reqs := make([]pbcs.CreateConfigItemReq, 0)

			r := pbcs.CreateConfigItemReq{
				BizId:     cases.TBizID,
				AppId:     appId,
				Name:      cases.RandName(preName),
				Path:      "/etc",
				FileType:  string(table.Xml),
				FileMode:  string(table.Unix),
				User:      "root",
				UserGroup: "root",
				Privilege: "755",
				Sign:      signature,
				ByteSize:  size,
			}

			// add preName field test case
			names := genAbnormalNameForTest()
			// to test: maximum length is 65
			names[1] = cases.RandNameN(preName, 65)
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandNameN(preName, 64)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			// biz id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// FileType is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = cases.WEnum
			reqs = append(reqs, r)
			r.FileType = string(table.Xml)

			// FileMode id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.FileMode = cases.WEnum
			reqs = append(reqs, r)
			r.FileMode = string(table.Unix)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.ConfigItem.Create(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})

	Convey("Update Config Item Test", t, func() {
		ciId := rm.GetConfigItem(appId)
		So(ciId, ShouldNotEqual, uint32(0))

		Convey("1.update_config_item normal test", func() {
			// test cases
			reqs := make([]pbcs.UpdateConfigItemReq, 0)

			r := pbcs.UpdateConfigItemReq{
				BizId:     cases.TBizID,
				AppId:     appId,
				Id:        ciId,
				Name:      cases.RandName(preName),
				Path:      "/etc",
				FileType:  string(table.Xml),
				FileMode:  string(table.Unix),
				User:      "root",
				UserGroup: "root",
				Privilege: "755",
				Sign:      signature,
				ByteSize:  size,
			}

			// add preName field test case
			names := []string{
				cases.TEnNumUnderHyphenDot + preName,
				cases.RandNameN(preName, 64),
				cases.TEnglish + preName,
				cases.TNumber + "6789",
			}
			for _, n := range names {
				r.Name = n
				reqs = append(reqs, r)
			}

			// add memo field test case
			memos := genNormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandNameN(preName, 64)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			// file type is yaml
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = string(table.Yaml)
			reqs = append(reqs, r)

			// file type is binary
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = string(table.Binary)
			reqs = append(reqs, r)

			// file type is json
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = string(table.Json)
			reqs = append(reqs, r)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.ConfigItem.Update(ctx, header, &req)
				So(err, ShouldBeNil)
				So(*resp, ShouldBeZeroValue)

				// unsupported list query by self-defined filter using ids
				// verify by list_config_item
				//listReq, err := cases.GenListConfigItemByIdsReq(cases.TBizID, appId, []uint32{ciId})
				//So(err, ShouldBeNil)
				//
				//ctx, header = cases.GenApiCtxHeader()
				//listResp, err := cli.ConfigItem.List(ctx, header, listReq)
				//So(err, ShouldBeNil)
				//So(listResp, ShouldNotBeNil)
				//
				//So(len(listResp.Details), ShouldEqual, 1)
				//one := listResp.Details[0]
				//So(one, ShouldNotBeNil)
				//So(one.Id, ShouldEqual, ciId)
				//So(one.Spec, ShouldNotBeNil)
				//So(one.Spec.Path, ShouldEqual, req.Path)
				//So(one.Spec.Name, ShouldEqual, req.Name)
				//So(one.Spec.FileType, ShouldEqual, req.FileType)
				//So(one.Spec.FileMode, ShouldEqual, req.FileMode)
				//So(one.Spec.Permission, ShouldNotBeNil)
				//So(one.Spec.Permission.User, ShouldEqual, req.User)
				//So(one.Spec.Permission.UserGroup, ShouldEqual, req.UserGroup)
				//So(one.Spec.Permission.Privilege, ShouldEqual, req.Privilege)
				//
				//So(one.Attachment, ShouldNotBeNil)
				//So(one.Attachment.AppId, ShouldEqual, appId)
				//So(one.Attachment.BizId, ShouldEqual, cases.TBizID)
			}
		})

		Convey("2.update_config_item abnormal test", func() {
			// test cases
			reqs := make([]pbcs.UpdateConfigItemReq, 0)

			r := pbcs.UpdateConfigItemReq{
				BizId:     cases.TBizID,
				AppId:     appId,
				Name:      cases.RandName(preName),
				Path:      "/etc",
				FileType:  string(table.Xml),
				FileMode:  string(table.Unix),
				User:      "root",
				UserGroup: "root",
				Privilege: "755",
				Sign:      signature,
				ByteSize:  size,
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

			// add memo field test case
			memos := genAbnormalMemoForTest()
			for _, memo := range memos {
				r.Name = cases.RandNameN(preName, 64)
				r.Memo = memo
				reqs = append(reqs, r)
			}

			// biz id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.BizId = cases.WID
			reqs = append(reqs, r)
			r.BizId = cases.TBizID

			// app id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.AppId = cases.WID
			reqs = append(reqs, r)
			r.AppId = appId

			// FileType is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.FileType = cases.WEnum
			reqs = append(reqs, r)
			r.FileType = string(table.Xml)

			// FileMode id is invalid
			r.Name = cases.RandNameN(preName, 64)
			r.FileMode = cases.WEnum
			reqs = append(reqs, r)
			r.FileMode = string(table.Unix)

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.ConfigItem.Update(ctx, header, &req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
	// unsupported list query by self-defined filter using ids
	//Convey("List Config Item Test", t, func() {
	//	// The normal list_config_item is test by the first create_config_item case,
	//	// so we just test list_config_item normal test on count page in here
	//	Convey("1.list_config_item normal test: count page", func() {
	//		// get a config item for list
	//		ciId := rm.GetConfigItem(appId)
	//		So(ciId, ShouldNotEqual, uint32(0))
	//		filter, err := cases.GenQueryFilterByIds([]uint32{ciId})
	//		So(err, ShouldBeNil)
	//
	//		req := &pbcs.ListConfigItemsReq{
	//			BizId:  cases.TBizID,
	//			AppId:  appId,
	//			Filter: filter,
	//			Page:   cases.CountPage(),
	//		}
	//
	//		ctx, header := cases.GenApiCtxHeader()
	//		resp, err := cli.ConfigItem.List(ctx, header, req)
	//		So(err, ShouldBeNil)
	//		So(resp, ShouldNotBeNil)
	//		So(resp.Count, ShouldEqual, uint32(1))
	//	})
	//
	//	Convey("2.list_config_item abnormal test", func() {
	//		// get a config item for list
	//		ciId := rm.GetConfigItem(appId)
	//		So(ciId, ShouldNotEqual, uint32(0))
	//		filter, err := cases.GenQueryFilterByIds([]uint32{ciId})
	//		So(err, ShouldBeNil)
	//
	//		reqs := []*pbcs.ListConfigItemsReq{
	//			{ // biz_id is invalid
	//				BizId:  cases.WID,
	//				AppId:  appId,
	//				Filter: filter,
	//				Page:   cases.CountPage(),
	//			},
	//			{ // app_id is invalid
	//				BizId:  cases.TBizID,
	//				AppId:  cases.WID,
	//				Filter: filter,
	//				Page:   cases.CountPage(),
	//			},
	//			{ // filter is invalid
	//				BizId:  cases.TBizID,
	//				AppId:  appId,
	//				Filter: nil,
	//				Page:   cases.CountPage(),
	//			},
	//			{ // page is invalid
	//				BizId:  cases.TBizID,
	//				AppId:  appId,
	//				Filter: filter,
	//				Page:   nil,
	//			},
	//		}
	//
	//		for _, req := range reqs {
	//			ctx, header := cases.GenApiCtxHeader()
	//			resp, err := cli.ConfigItem.List(ctx, header, req)
	//			So(err, ShouldBeNil)
	//			So(resp, ShouldNotBeNil)
	//		}
	//	})
	//})

	Convey("Delete Config Item Test", t, func() {
		Convey("1.delete_config_item normal test", func() {
			// get a config item for delete
			ciId := rm.GetConfigItem(appId)
			So(ciId, ShouldNotEqual, uint32(0))
			defer rm.DeleteConfigItem(appId, ciId)

			req := &pbcs.DeleteConfigItemReq{
				Id:    ciId,
				BizId: cases.TBizID,
				AppId: appId,
			}

			ctx, header := cases.GenApiCtxHeader()
			resp, err := cli.ConfigItem.Delete(ctx, header, req)
			So(err, ShouldBeNil)
			So(*resp, ShouldBeZeroValue)

			// verify by list
			//listReq, err := cases.GenListConfigItemByIdsReq(cases.TBizID, appId, []uint32{ciId})
			//So(err, ShouldBeNil)
			//ctx, header = cases.GenApiCtxHeader()
			//listResp, err := cli.ConfigItem.List(ctx, header, listReq)
			//So(err, ShouldBeNil)
			//So(listResp, ShouldNotBeNil)
			//So(len(listResp.Details), ShouldEqual, 0)
		})

		Convey("2.delete_config_item abnormal test", func() {
			// get a config item for delete
			ciId := rm.GetConfigItem(appId)

			reqs := []*pbcs.DeleteConfigItemReq{
				{ // config item id is invalid
					Id:    cases.WID,
					BizId: cases.TBizID,
					AppId: appId,
				},
				{ // biz id is invalid
					Id:    ciId,
					BizId: cases.WID,
					AppId: appId,
				},
				{ // app id is invalid
					Id:    ciId,
					BizId: cases.TBizID,
					AppId: cases.WID,
				},
			}

			for _, req := range reqs {
				ctx, header := cases.GenApiCtxHeader()
				resp, err := cli.ConfigItem.Delete(ctx, header, req)
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			}
		})
	})
}
