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

package cache

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbcache "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbconfig "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestCache(t *testing.T) {

	SetDefaultFailureMode(FailureHalts)

	var (
		cli *client.Client

		// The following values are used to generate data and
		// verify data in the test process.
		// They need to be defined before they can be used.
		name         string
		content      string // content of a config item file
		nmApp        pbapp.App
		nmConfigItem pbci.ConfigItem
		nmContent    pbcontent.Content
		// NOTE: strategy related test depends on group, add group test first
		//nmStrategy   pbstrategy.Strategy

		// The following ids need to be assigned after generating resources.
		// They are used for verification in the test process.
		nmCommitId    uint32 // normal mode commit id
		nmReleaseId   uint32 // normal mode release id
		nmInstanceUid string // normal mode instance uid
	)

	Convey("Prepare Job For Cache Test", t, func() {
		cli = suite.GetClient()

		Convey("Define The Resource", func() {
			// defined the resource below
			name = "cache_test"
			content = "this is a file content"
			nmInstanceUid = uuid.UUID()

			// normal mode app. Its ID needs to be assigned after creation.
			nmApp = pbapp.App{
				Spec: &pbapp.AppSpec{
					ConfigType: string(table.File),
					Mode:       string(table.Normal),
				},
			}

			// normal mode config item.
			nmConfigItem = pbci.ConfigItem{
				Spec: &pbci.ConfigItemSpec{
					Name:     name + "_config_item",
					Path:     "/etc",
					FileType: string(table.Xml),
					FileMode: string(table.Unix),
					Permission: &pbci.FilePermission{
						User:      "root",
						UserGroup: "root",
						Privilege: "777",
					},
				},
			}

			// normal mode content.
			nmContent = pbcontent.Content{
				Spec: &pbcontent.ContentSpec{
					Signature: tools.SHA256(content),
					ByteSize:  uint64(len(content)),
				},
			}

			// NOTE: strategy related test depends on group, add group test first
			// normal mode strategy.
			//nmStrategy = pbstrategy.Strategy{
			//	Spec: &pbstrategy.StrategySpec{
			//		AsDefault: false,
			//		Name:      name + "_strategy",
			//		Namespace: "",
			//	},
			//}
		})

		Convey("Generate Resource For Cache Test", func() {

			var err error
			// create resource for test
			// create app: the application has release instance
			nmAppReq := &pbconfig.CreateAppReq{
				BizId:          cases.TBizID,
				Name:           cases.RandName(name),
				ConfigType:     nmApp.Spec.ConfigType,
				Mode:           nmApp.Spec.Mode,
				ReloadType:     string(table.ReloadWithFile),
				ReloadFilePath: "/tmp/reload.json",
			}
			ctx, header := cases.GenApiCtxHeader()
			nmAppResp, err := cli.ApiClient.App.Create(ctx, header, nmAppReq)
			So(err, ShouldBeNil)
			So(nmAppResp, ShouldNotBeNil)
			So(nmAppResp.Id, ShouldNotEqual, uint32(0))
			nmApp.Id = nmAppResp.Id

			// create config item in normal mode
			nmCIReq := &pbconfig.CreateConfigItemReq{
				BizId:     cases.TBizID,
				AppId:     nmAppResp.Id,
				Name:      nmConfigItem.Spec.Name,
				Path:      nmConfigItem.Spec.Path,
				FileType:  nmConfigItem.Spec.FileType,
				FileMode:  nmConfigItem.Spec.FileMode,
				User:      nmConfigItem.Spec.Permission.User,
				UserGroup: nmConfigItem.Spec.Permission.UserGroup,
				Privilege: nmConfigItem.Spec.Permission.Privilege,
				Sign:      nmContent.Spec.Signature,
				ByteSize:  nmContent.Spec.ByteSize,
			}
			ctx, header = cases.GenApiCtxHeader()
			nmCiResp, err := cli.ApiClient.ConfigItem.Create(ctx, header, nmCIReq)
			So(err, ShouldBeNil)
			So(nmCiResp, ShouldNotBeNil)
			So(nmCiResp.Id, ShouldNotEqual, uint32(0))

			// upload content in normal mode
			ctx, header = cases.GenApiCtxHeader()
			header.Set(constant.ContentIDHeaderKey, nmContent.Spec.Signature)
			resp, err := cli.ApiClient.Content.Upload(ctx, header, cases.TBizID, nmAppResp.Id, content)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			// create content in normal mode
			nmCtReq := &pbconfig.CreateContentReq{
				BizId:        cases.TBizID,
				AppId:        nmAppResp.Id,
				ConfigItemId: nmCiResp.Id,
				Sign:         nmContent.Spec.Signature,
				ByteSize:     nmContent.Spec.ByteSize,
			}
			ctx, header = cases.GenApiCtxHeader()
			nmCtResp, err := cli.ApiClient.Content.Create(ctx, header, nmCtReq)
			So(err, ShouldBeNil)
			So(nmCtResp, ShouldNotBeNil)
			So(nmCtResp.Id, ShouldNotEqual, uint32(0))
			nmContent.Id = nmCtResp.Id

			// create commit in normal mode
			nmCmReq := &pbconfig.CreateCommitReq{
				BizId:        cases.TBizID,
				AppId:        nmAppResp.Id,
				ConfigItemId: nmCiResp.Id,
				ContentId:    nmCtResp.Id,
			}
			ctx, header = cases.GenApiCtxHeader()
			nmCmResp, err := cli.ApiClient.Commit.Create(ctx, header, nmCmReq)
			So(err, ShouldBeNil)
			So(nmCmResp, ShouldNotBeNil)
			So(nmCmResp.Id, ShouldNotEqual, uint32(0))
			nmCommitId = nmCmResp.Id

			// create release in normal mode
			nmRelReq := &pbconfig.CreateReleaseReq{
				BizId: cases.TBizID,
				AppId: nmAppResp.Id,
				Name:  name + "_release",
			}
			ctx, header = cases.GenApiCtxHeader()
			nmRelResp, err := cli.ApiClient.Release.Create(ctx, header, nmRelReq)
			So(err, ShouldBeNil)
			So(nmRelResp, ShouldNotBeNil)
			So(nmRelResp.Id, ShouldNotEqual, uint32(0))
			nmReleaseId = nmRelResp.Id

			// publish instance in normal mode
			nmInsReq := &pbconfig.PublishInstanceReq{
				BizId:     cases.TBizID,
				AppId:     nmAppResp.Id,
				Uid:       nmInstanceUid,
				ReleaseId: nmRelResp.Id,
			}
			ctx, header = cases.GenApiCtxHeader()
			nmInsResp, err := cli.ApiClient.Instance.Publish(ctx, header, nmInsReq)
			So(err, ShouldBeNil)
			So(nmInsResp, ShouldNotBeNil)
			So(nmInsResp.Id, ShouldNotEqual, uint32(0))

			// create strategy set in normal mode
			nmStgSetReq := &pbconfig.CreateStrategySetReq{
				BizId: cases.TBizID,
				AppId: nmAppResp.Id,
				Name:  name + "_strategy_set",
			}
			ctx, header = cases.GenApiCtxHeader()
			nmStgSetResp, err := cli.ApiClient.StrategySet.Create(ctx, header, nmStgSetReq)
			So(err, ShouldBeNil)
			So(nmStgSetResp, ShouldNotBeNil)
			So(nmStgSetResp.Id, ShouldNotEqual, uint32(0))

			// NOTE: strategy related test depends on group, add group test first
			//// create strategy in normal mode
			//nmStgReq := &pbconfig.CreateStrategyReq{
			//	BizId:         cases.TBizID,
			//	AppId:         nmAppResp.Id,
			//	StrategySetId: nmStgSetResp.Id,
			//	ReleaseId:     nmRelResp.Id,
			//	AsDefault:     nmStrategy.Spec.AsDefault,
			//	Name:          name + "_strategy",
			//	Namespace:     "",
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nmStgResp, err := cli.ApiClient.Strategy.Create(ctx, header, nmStgReq)
			//So(err, ShouldBeNil)
			//So(nmStgResp, ShouldNotBeNil)
			//So(nmStgResp.Id, ShouldNotEqual, uint32(0))
			//
			//// publish with strategy in normal mode
			//pubStgReq := &pbconfig.PublishReq{
			//	BizId: cases.TBizID,
			//	AppId: nmAppResp.Id,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nmPubStgResp, err := cli.ApiClient.Publish.PublishWithStrategy(ctx, header, pubStgReq)
			//So(err, ShouldBeNil)
			//So(nmPubStgResp, ShouldNotBeNil)
			//So(nmPubStgResp.Id, ShouldNotEqual, uint32(0))
			//
			//// finish publish with strategy in normal mode
			//nmFinishPubStgReq := &pbconfig.FinishPublishReq{
			//	BizId: cases.TBizID,
			//	AppId: nmAppResp.Id,
			//	Id:    nmStgResp.Id,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//finishPubStgResp, err := cli.ApiClient.Publish.FinishPublishWithStrategy(ctx, header, nmFinishPubStgReq)
			//So(err, ShouldBeNil)
			//So(*finishPubStgResp, ShouldBeZeroValue)
			//// the application doesn't have release instance
			//nsAppReq := &pbconfig.CreateAppReq{
			//	BizId:          cases.TBizID,
			//	Name:           cases.RandName(name),
			//	ConfigType:     string(table.File),
			//	Mode:           string(table.Normal),
			//	ReloadType:     string(table.ReloadWithFile),
			//	ReloadFilePath: "/tmp/reload.json",
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsAppResp, err := cli.ApiClient.App.Create(ctx, header, nsAppReq)
			//So(err, ShouldBeNil)
			//So(nsAppResp, ShouldNotBeNil)
			//So(nsAppResp.Id, ShouldNotEqual, uint32(0))

			// wait for cache to flush
			time.Sleep(time.Second * 2)
		})
	})

	Convey("Cache Service Suite Test", t, func() {

		Convey("1.Get Application Meta Test", func() {
			// normal test
			{
				req := &pbcache.GetAppMetaReq{
					BizId: cases.TBizID,
					AppId: nmApp.Id,
				}

				ctx := cases.GenCacheContext()
				resp, err := cli.CacheClient.GetAppMeta(ctx, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Mode, ShouldEqual, nmApp.Spec.Mode)
				So(resp.ConfigType, ShouldEqual, nmApp.Spec.ConfigType)
				So(resp.Reload.ReloadType, ShouldEqual, table.ReloadWithFile)
				So(resp.Reload.FileReloadSpec.ReloadFilePath, ShouldEqual, "/tmp/reload.json")
			}

			// abnormal test
			{
				reqs := []*pbcache.GetAppMetaReq{
					{
						BizId: cases.WID,
						AppId: nmApp.Id,
					},
					{
						BizId: cases.TBizID,
						AppId: cases.WID,
					},
				}

				for _, req := range reqs {
					ctx := cases.GenCacheContext()
					_, err := cli.CacheClient.GetAppMeta(ctx, req)
					So(err, ShouldNotBeNil)
				}
			}

		})

		Convey("2.Get Released Config Item Test", func() {
			// normal test
			{
				req := &pbcache.GetReleasedCIReq{
					BizId:     cases.TBizID,
					ReleaseId: nmReleaseId,
				}

				ctx := cases.GenCacheContext()
				resp, err := cli.CacheClient.GetReleasedCI(ctx, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(len(resp), ShouldEqual, 1)
				So(resp[0], ShouldNotBeNil)

				So(resp[0].ID, ShouldNotEqual, uint32(0))
				So(resp[0].CommitID, ShouldEqual, nmCommitId)
				So(resp[0].ReleaseID, ShouldEqual, nmReleaseId)

				cmSpec := resp[0].CommitSpec
				So(cmSpec, ShouldNotBeNil)
				So(cmSpec.ContentID, ShouldEqual, nmContent.Id)
				So(cmSpec.ByteSize, ShouldEqual, nmContent.Spec.ByteSize)
				So(cmSpec.Signature, ShouldEqual, nmContent.Spec.Signature)

				ciSpec := resp[0].ConfigItemSpec
				So(ciSpec, ShouldNotBeNil)
				So(ciSpec.Name, ShouldEqual, nmConfigItem.Spec.Name)
				So(ciSpec.Path, ShouldEqual, nmConfigItem.Spec.Path)
				So(ciSpec.FileType, ShouldEqual, nmConfigItem.Spec.FileType)
				So(ciSpec.FileMode, ShouldEqual, nmConfigItem.Spec.FileMode)

				So(ciSpec.Permission, ShouldNotBeNil)
				permission := nmConfigItem.Spec.Permission
				So(ciSpec.Permission.User, ShouldEqual, permission.User)
				So(ciSpec.Permission.UserGroup, ShouldEqual, permission.UserGroup)
				So(ciSpec.Permission.Privilege, ShouldEqual, permission.Privilege)
			}

			// abnormal test
			{
				reqs := []*pbcache.GetReleasedCIReq{
					{
						BizId:     cases.WID,
						ReleaseId: nmReleaseId,
					},
					{
						BizId:     cases.TBizID,
						ReleaseId: cases.WID,
					},
					{
						BizId:     cases.TBizID - 1,
						ReleaseId: nmReleaseId,
					},
					{
						BizId:     cases.TBizID,
						ReleaseId: cases.TBizID - 1,
					},
				}

				for _, cacheReq := range reqs {
					ctx := cases.GenCacheContext()
					_, err := cli.CacheClient.GetReleasedCI(ctx, cacheReq)
					So(err, ShouldNotBeNil)
				}
			}
		})

		Convey("4.Get Application Instance ReleaseID Test", func() {
			// normal test: application has release instance.
			{
				req := &pbcache.GetAppInstanceReleaseReq{
					BizId: cases.TBizID,
					AppId: nmApp.Id,
					Uid:   nmInstanceUid,
				}

				ctx := cases.GenCacheContext()
				resp, err := cli.CacheClient.GetAppInstanceRelease(ctx, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.ReleaseId, ShouldEqual, nmReleaseId)
			}

			// normal test: the uid is not exist.
			{
				req := &pbcache.GetAppInstanceReleaseReq{
					BizId: cases.TBizID,
					AppId: nmApp.Id,
					Uid:   uuid.UUID(),
				}

				ctx := cases.GenCacheContext()
				resp, err := cli.CacheClient.GetAppInstanceRelease(ctx, req)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.ReleaseId, ShouldEqual, uint32(0))
			}

			// abnormal test: set a invalid value.
			{
				reqs := []*pbcache.GetAppInstanceReleaseReq{
					{
						BizId: cases.WID,
						AppId: nmApp.Id,
						Uid:   uuid.UUID(),
					},
					{
						BizId: cases.TBizID,
						AppId: cases.WID,
						Uid:   uuid.UUID(),
					},
				}

				for _, req := range reqs {
					ctx := cases.GenCacheContext()
					_, err := cli.CacheClient.GetAppInstanceRelease(ctx, req)
					So(err, ShouldNotBeNil)
				}
			}
		})
	})
}
