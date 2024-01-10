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

package feed

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
)

func TestFeed(t *testing.T) {
	SetDefaultFailureMode(FailureHalts)

	var (
		// resource value, they
		nmApp pbapp.App
		//nsApp      pbapp.App
		ConfigItem pbci.ConfigItem
		Content    pbcontent.Content

		// normal mode instance uid
		nmContentId   uint32
		nmReleaseId   uint32
		nmInstanceUid string

		// namespace mode
		//nsContentId uint32
		//nsReleaseId    uint32
		//nsStgNamespace string
	)

	cli := suite.GetClient()

	Convey("Feed Server Suite Test", t, func() {
		Convey("Generate Resource For Feed Test", func() {
			// if the values need to be verified, they need to be defined below
			content := "this is a file content"
			nmInstanceUid = uuid.UUID()
			//nsStgNamespace = "This_Is_NameSpace_Strategy"

			// normal mode app. Its ID needs to be assigned after creation.
			nmApp = pbapp.App{
				Spec: &pbapp.AppSpec{
					ConfigType: string(table.File),
					Mode:       string(table.Normal),
				},
			}

			// namespace mode app. Its ID needs to be assigned after creation.
			//nsApp = pbapp.App{
			//	Spec: &pbapp.AppSpec{
			//		ConfigType: string(table.File),
			//		Mode:       string(table.Namespace),
			//	},
			//}

			// normal mode config item.
			ConfigItem = pbci.ConfigItem{
				Spec: &pbci.ConfigItemSpec{
					Name:     "feed_test_config_item",
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
			Content = pbcontent.Content{
				Spec: &pbcontent.ContentSpec{
					Signature: tools.SHA256(content),
					ByteSize:  uint64(len(content)),
				},
			}

			var err error
			nmName := "normal_mode_feed"

			// create app
			nmAppReq := &pbcs.CreateAppReq{ // the application has release instance
				BizId:          cases.TBizID,
				Name:           cases.RandName(nmName),
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

			// create config item
			nmCIReq := &pbcs.CreateConfigItemReq{
				BizId:     cases.TBizID,
				AppId:     nmAppResp.Id,
				Name:      ConfigItem.Spec.Name,
				Path:      ConfigItem.Spec.Path,
				FileType:  ConfigItem.Spec.FileType,
				FileMode:  ConfigItem.Spec.FileMode,
				User:      ConfigItem.Spec.Permission.User,
				UserGroup: ConfigItem.Spec.Permission.UserGroup,
				Privilege: ConfigItem.Spec.Permission.Privilege,
				Sign:      Content.Spec.Signature,
				ByteSize:  Content.Spec.ByteSize,
			}
			ctx, header = cases.GenApiCtxHeader()
			nmCiResp, err := cli.ApiClient.ConfigItem.Create(ctx, header, nmCIReq)
			So(err, ShouldBeNil)
			So(nmCiResp, ShouldNotBeNil)
			So(nmCiResp.Id, ShouldNotEqual, uint32(0))

			// upload content
			ctx, header = cases.GenApiCtxHeader()
			header.Set(constant.ContentIDHeaderKey, Content.Spec.Signature)
			resp, err := cli.ApiClient.Content.Upload(ctx, header, cases.TBizID, nmAppResp.Id, content)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			// create content
			nmCtReq := &pbcs.CreateContentReq{
				BizId:        cases.TBizID,
				AppId:        nmAppResp.Id,
				ConfigItemId: nmCiResp.Id,
				Sign:         Content.Spec.Signature,
				ByteSize:     Content.Spec.ByteSize,
			}
			ctx, header = cases.GenApiCtxHeader()
			nmCtResp, err := cli.ApiClient.Content.Create(ctx, header, nmCtReq)
			So(err, ShouldBeNil)
			So(nmCtResp, ShouldNotBeNil)
			So(nmCtResp.Id, ShouldNotEqual, uint32(0))
			nmContentId = nmCtResp.Id

			// create commit
			nmCmReq := &pbcs.CreateCommitReq{
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

			// create release
			nmRelReq := &pbcs.CreateReleaseReq{
				BizId: cases.TBizID,
				AppId: nmAppResp.Id,
				Name:  nmName + "_release",
			}
			ctx, header = cases.GenApiCtxHeader()
			nmRelResp, err := cli.ApiClient.Release.Create(ctx, header, nmRelReq)
			So(err, ShouldBeNil)
			So(nmRelResp, ShouldNotBeNil)
			So(nmRelResp.Id, ShouldNotEqual, uint32(0))
			nmReleaseId = nmRelResp.Id

			// publish instance
			nmInsReq := &pbcs.PublishInstanceReq{
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

			// create strategy set
			nmStgSetReq := &pbcs.CreateStrategySetReq{
				BizId: cases.TBizID,
				AppId: nmAppResp.Id,
				Name:  nmName + "_strategy_set",
			}
			ctx, header = cases.GenApiCtxHeader()
			nmStgSetResp, err := cli.ApiClient.StrategySet.Create(ctx, header, nmStgSetReq)
			So(err, ShouldBeNil)
			So(nmStgSetResp, ShouldNotBeNil)
			So(nmStgSetResp.Id, ShouldNotEqual, uint32(0))

			// NOTE: strategy related test depends on group, add group test first
			//// create strategy
			//nmStgReq := &pbcs.CreateStrategyReq{
			//	BizId:         cases.TBizID,
			//	AppId:         nmAppResp.Id,
			//	StrategySetId: nmStgSetResp.Id,
			//	ReleaseId:     nmRelResp.Id,
			//	AsDefault:     false,
			//	Name:          nmName + "_strategy",
			//	Namespace:     "",
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nmStgResp, err := cli.ApiClient.Strategy.Create(ctx, header, nmStgReq)
			//So(err, ShouldBeNil)
			//So(nmStgResp, ShouldNotBeNil)
			//So(nmStgResp.Id, ShouldNotEqual, uint32(0))
			//
			//// publish with strategy
			//nmPubStgReq := &pbcs.PublishReq{
			//	BizId: cases.TBizID,
			//	AppId: nmAppResp.Id,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nmPubStgResp, err := cli.ApiClient.Publish.PublishWithStrategy(ctx, header, nmPubStgReq)
			//So(err, ShouldBeNil)
			//So(nmPubStgResp, ShouldNotBeNil)
			//So(nmPubStgResp.Id, ShouldNotEqual, uint32(0))
			//
			//// finish publish with strategy
			//nmFinishPubStgReq := &pbcs.FinishPublishReq{
			//	BizId: cases.TBizID,
			//	AppId: nmAppResp.Id,
			//	Id:    nmStgResp.Id,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nmFinishPubStgResp, err := cli.ApiClient.Publish.FinishPublishWithStrategy(ctx, header, nmFinishPubStgReq)
			//So(err, ShouldBeNil)
			//So(nmFinishPubStgResp, ShouldNotBeNil)
			//
			//// namespace mode
			//nsName := "namespace_mode_feed"
			//// create namespace mode app
			//nsAppNsReq := &pbcs.CreateAppReq{
			//	BizId:          cases.TBizID,
			//	Name:           cases.RandName(nsName),
			//	ConfigType:     nsApp.Spec.ConfigType,
			//	Mode:           nsApp.Spec.Mode,
			//	ReloadType:     string(table.ReloadWithFile),
			//	ReloadFilePath: "/tmp/reload.json",
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsAppResp, err := cli.ApiClient.App.Create(ctx, header, nsAppNsReq)
			//So(err, ShouldBeNil)
			//So(nsAppResp, ShouldNotBeNil)
			//So(nsAppResp.Id, ShouldNotEqual, uint32(0))
			//nsApp.Id = nsAppResp.Id
			//
			//// create config item
			//nsCiReq := &pbcs.CreateConfigItemReq{
			//	BizId:     cases.TBizID,
			//	AppId:     nsAppResp.Id,
			//	Name:      ConfigItem.Spec.Name,
			//	Path:      ConfigItem.Spec.Path,
			//	FileType:  ConfigItem.Spec.FileType,
			//	FileMode:  ConfigItem.Spec.FileMode,
			//	User:      ConfigItem.Spec.Permission.User,
			//	UserGroup: ConfigItem.Spec.Permission.UserGroup,
			//	Privilege: ConfigItem.Spec.Permission.Privilege,
			//	Sign:      Content.Spec.Signature,
			//	ByteSize:  Content.Spec.ByteSize,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsCiResp, err := cli.ApiClient.ConfigItem.Create(ctx, header, nsCiReq)
			//So(err, ShouldBeNil)
			//So(nsCiResp, ShouldNotBeNil)
			//So(nsCiResp.Id, ShouldNotEqual, uint32(0))
			//
			//// upload content
			//ctx, header = cases.GenApiCtxHeader()
			//header.Set(constant.ContentIDHeaderKey, Content.Spec.Signature)
			//nsUpResp, err := cli.ApiClient.Content.Upload(ctx, header, cases.TBizID, nmAppResp.Id, content)
			//So(err, ShouldBeNil)
			//So(nsUpResp, ShouldNotBeNil)
			//
			//// create content
			//nsCtReq := &pbcs.CreateContentReq{
			//	BizId:        cases.TBizID,
			//	AppId:        nsAppResp.Id,
			//	ConfigItemId: nsCiResp.Id,
			//	Sign:         Content.Spec.Signature,
			//	ByteSize:     Content.Spec.ByteSize,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsCtResp, err := cli.ApiClient.Content.Create(ctx, header, nsCtReq)
			//So(err, ShouldBeNil)
			//So(nsCtResp, ShouldNotBeNil)
			//So(nsCtResp.Id, ShouldNotEqual, uint32(0))
			//nsContentId = nsCtResp.Id
			//
			//// create commit
			//nsCmReq := &pbcs.CreateCommitReq{
			//	BizId:        cases.TBizID,
			//	AppId:        nsAppResp.Id,
			//	ConfigItemId: nsCiResp.Id,
			//	ContentId:    nsContentId,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsCmResp, err := cli.ApiClient.Commit.Create(ctx, header, nsCmReq)
			//So(err, ShouldBeNil)
			//So(nsCmResp, ShouldNotBeNil)
			//So(nsCmResp.Id, ShouldNotEqual, uint32(0))
			//
			//// create release
			//nsRelReq := &pbcs.CreateReleaseReq{
			//	BizId: cases.TBizID,
			//	AppId: nsAppResp.Id,
			//	Name:  nsName + "_release",
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsRelResp, err := cli.ApiClient.Release.Create(ctx, header, nsRelReq)
			//So(err, ShouldBeNil)
			//So(nsRelResp, ShouldNotBeNil)
			//So(nsRelResp.Id, ShouldNotEqual, uint32(0))
			////nsReleaseId = nsRelResp.Id
			//
			//// create strategy set
			//nsStgSetReq := &pbcs.CreateStrategySetReq{
			//	BizId: cases.TBizID,
			//	AppId: nsAppResp.Id,
			//	Name:  nsName + "_strategy_set",
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsStgSetResp, err := cli.ApiClient.StrategySet.Create(ctx, header, nsStgSetReq)
			//So(err, ShouldBeNil)
			//So(nsStgSetResp, ShouldNotBeNil)
			//So(nsStgSetResp.Id, ShouldNotEqual, uint32(0))
			//
			//// create namespace mode strategy scope
			//So(err, ShouldBeNil)
			//nsStgReq := &pbcs.CreateStrategyReq{
			//	BizId:         cases.TBizID,
			//	AppId:         nsAppResp.Id,
			//	StrategySetId: nsStgSetResp.Id,
			//	ReleaseId:     nsRelResp.Id,
			//	AsDefault:     false,
			//	Name:          nsName + "_strategy",
			//	Namespace:     nsStgNamespace,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsStgResp, err := cli.ApiClient.Strategy.Create(ctx, header, nsStgReq)
			//So(err, ShouldBeNil)
			//So(nsStgResp, ShouldNotBeNil)
			//So(nsStgResp.Id, ShouldNotEqual, uint32(0))
			//
			//// publish with strategy
			//nsPubStgReq := &pbcs.PublishReq{
			//	BizId: cases.TBizID,
			//	AppId: nsAppResp.Id,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsPubStgResp, err := cli.ApiClient.Publish.PublishWithStrategy(ctx, header, nsPubStgReq)
			//So(err, ShouldBeNil)
			//So(nsPubStgResp, ShouldNotBeNil)
			//So(nsPubStgResp.Id, ShouldNotEqual, uint32(0))
			//
			//// finish publish with strategy
			//nsFinishPubStgReq := &pbcs.FinishPublishReq{
			//	BizId: cases.TBizID,
			//	AppId: nsAppResp.Id,
			//	Id:    nsStgResp.Id,
			//}
			//ctx, header = cases.GenApiCtxHeader()
			//nsFinishPubStgResp, err := cli.ApiClient.Publish.FinishPublishWithStrategy(ctx, header, nsFinishPubStgReq)
			//So(err, ShouldBeNil)
			//So(nsFinishPubStgResp, ShouldNotBeNil)

			gen := Generator{
				Cli: cli.ApiClient,
			}

			err = gen.GenData(kit.New())
			So(err, ShouldBeNil)

			// wait for cache to flush
			time.Sleep(time.Second * 2)
		})

		// 1. Normal Mode App Base Request Test: in normal mode app under, feed server handle test right and not
		// right request params, return response data is expected.
		Convey("1. Normal Mode App Base Request Test", func() {
			// normal test
			{
				reqs := []*types.ListFileAppLatestReleaseMetaReq{
					// NOTE: strategy related test depends on group, add group test first
					//{ // Publish with strategy, and hit main strategy
					//	BizId: cases.TBizID,
					//	AppId: nmApp.Id,
					//	Uid:   uuid.UUID(),
					//	Labels: map[string]string{
					//		"os": "windows",
					//	},
					//},
					//{ // Publish with strategy, and hit sub strategy
					//	BizId: cases.TBizID,
					//	AppId: nmApp.Id,
					//	Uid:   uuid.UUID(),
					//	Labels: map[string]string{
					//		"area": "south of china",
					//		"city": "shenzhen",
					//	},
					//},
					{ // publish with instance
						BizId:  cases.TBizID,
						AppId:  nmApp.Id,
						Uid:    nmInstanceUid,
						Labels: nil,
					},
				}

				for _, req := range reqs {
					ctx, header := cases.GenApiCtxHeader()
					resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Code, ShouldEqual, errf.OK)

					So(resp.Data, ShouldNotBeNil)
					So(resp.Data.Repository, ShouldNotBeNil)
					So(resp.Data.Repository.Root, ShouldNotBeBlank)
					So(resp.Data.ReleaseId, ShouldEqual, nmReleaseId)

					So(len(resp.Data.ConfigItems), ShouldEqual, 1)
					ci := resp.Data.ConfigItems[0]
					So(ci, ShouldNotBeNil)
					So(ci.RciId, ShouldNotEqual, uint32(0))
					So(ci.RepositorySpec, ShouldNotBeNil)
					So(ci.RepositorySpec.Path, ShouldNotBeBlank)

					ciSpec := ci.ConfigItemSpec
					So(ciSpec, ShouldNotBeNil)
					So(ciSpec.Path, ShouldEqual, ConfigItem.Spec.Path)
					So(ciSpec.Name, ShouldEqual, ConfigItem.Spec.Name)
					So(ciSpec.FileMode, ShouldEqual, ConfigItem.Spec.FileMode)
					So(ciSpec.FileType, ShouldEqual, ConfigItem.Spec.FileType)
					So(ciSpec.Permission, ShouldNotBeNil)
					So(ciSpec.Permission.User, ShouldEqual, ConfigItem.Spec.Permission.User)
					So(ciSpec.Permission.UserGroup, ShouldEqual, ConfigItem.Spec.Permission.UserGroup)
					So(ciSpec.Permission.Privilege, ShouldEqual, ConfigItem.Spec.Permission.Privilege)

					cmSpec := ci.CommitSpec
					So(cmSpec, ShouldNotBeNil)
					So(cmSpec.ContentId, ShouldEqual, nmContentId)
					So(cmSpec.Content, ShouldNotBeNil)
					So(cmSpec.Content.ByteSize, ShouldEqual, Content.Spec.ByteSize)
					So(cmSpec.Content.Signature, ShouldEqual, Content.Spec.Signature)
				}
			}

			// abnormal test
			{
				reqs := []*types.ListFileAppLatestReleaseMetaReq{
					{ // wrong biz id
						BizId: cases.WID,
						AppId: nmApp.Id,
						Uid:   uuid.UUID(),
						Labels: map[string]string{
							"os": "windows",
						},
					},
					{ // wrong app id
						BizId: cases.TBizID,
						AppId: cases.WID,
						Uid:   uuid.UUID(),
						Labels: map[string]string{
							"area": "south of china",
						},
					},
					{ // uid is null
						BizId: cases.TBizID,
						AppId: nmApp.Id,
						Uid:   "",
						Labels: map[string]string{
							"area": "south of china",
						},
					},
					{ // uid is out of limited length
						BizId: cases.TBizID,
						AppId: nmApp.Id,
						Uid:   cases.RandString(65),
						Labels: map[string]string{
							"area": "south of china",
						},
					},
					{ // don't hit main strategy
						BizId: cases.TBizID,
						AppId: nmApp.Id,
						Uid:   uuid.UUID(),
						Labels: map[string]string{
							"os": "unix",
						},
					},
					{ // non-exist instance uid
						BizId: cases.TBizID,
						AppId: nmApp.Id,
						Uid:   uuid.UUID(),
						Labels: map[string]string{
							"test": "test",
						},
					},
					{ // label is null
						BizId:  cases.TBizID,
						AppId:  nmApp.Id,
						Uid:    uuid.UUID(),
						Labels: nil,
					},
				}

				for _, req := range reqs {
					ctx, header := cases.GenApiCtxHeader()
					resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
				}
			}

		})

		// NOTE: strategy related test depends on group, add group test first
		//// 2. Namespace Mode App Base Request Test: in namespace mode app under, feed server handle test right and not
		//// right request params, return response data is expected.
		//Convey("2. Namespace Mode App Base Request Test", func() {
		//	// normal test
		//	{
		//		req := &types.ListFileAppLatestReleaseMetaReq{
		//			BizId: cases.TBizID,
		//			AppId: nsApp.Id,
		//			Uid:   uuid.UUID(),
		//			Labels: map[string]string{
		//				"city": "shanghai",
		//			},
		//			Namespace: nsStgNamespace,
		//		}
		//
		//		ctx, header := cases.GenApiCtxHeader()
		//		resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//		So(err, ShouldBeNil)
		//		So(resp, ShouldNotBeNil)
		//		So(resp.Code, ShouldEqual, errf.OK)
		//
		//		So(resp.Data, ShouldNotBeNil)
		//		So(resp.Data.Repository, ShouldNotBeNil)
		//		So(resp.Data.Repository.Root, ShouldNotBeBlank)
		//		So(resp.Data.ReleaseId, ShouldEqual, nsReleaseId)
		//
		//		So(len(resp.Data.ConfigItems), ShouldEqual, 1)
		//		ci := resp.Data.ConfigItems[0]
		//		So(ci, ShouldNotBeNil)
		//		So(ci.RciId, ShouldNotEqual, uint32(0))
		//		So(ci.RepositorySpec, ShouldNotBeNil)
		//		So(ci.RepositorySpec.Path, ShouldNotBeBlank)
		//
		//		ciSpec := ci.ConfigItemSpec
		//		So(ciSpec, ShouldNotBeNil)
		//		So(ciSpec.Path, ShouldEqual, ConfigItem.Spec.Path)
		//		So(ciSpec.Name, ShouldEqual, ConfigItem.Spec.Name)
		//		So(ciSpec.FileMode, ShouldEqual, ConfigItem.Spec.FileMode)
		//		So(ciSpec.FileType, ShouldEqual, ConfigItem.Spec.FileType)
		//		So(ciSpec.Permission, ShouldNotBeNil)
		//		So(ciSpec.Permission.User, ShouldEqual, ConfigItem.Spec.Permission.User)
		//		So(ciSpec.Permission.UserGroup, ShouldEqual, ConfigItem.Spec.Permission.UserGroup)
		//		So(ciSpec.Permission.Privilege, ShouldEqual, ConfigItem.Spec.Permission.Privilege)
		//
		//		cmSpec := ci.CommitSpec
		//		So(cmSpec, ShouldNotBeNil)
		//		So(cmSpec.ContentId, ShouldEqual, nsContentId)
		//		So(cmSpec.Content, ShouldNotBeNil)
		//		So(cmSpec.Content.ByteSize, ShouldEqual, Content.Spec.ByteSize)
		//		So(cmSpec.Content.Signature, ShouldEqual, Content.Spec.Signature)
		//	}
		//
		//	// abnormal test
		//	{
		//		req := &types.ListFileAppLatestReleaseMetaReq{
		//			BizId: cases.TBizID,
		//			AppId: nsApp.Id,
		//			Uid:   uuid.UUID(),
		//			Labels: map[string]string{
		//				"city": "shanghai",
		//			},
		//			Namespace: "this_is_a_wrong_namespace",
		//		}
		//		ctx, header := cases.GenApiCtxHeader()
		//		resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//		So(err, ShouldBeNil)
		//		So(resp, ShouldNotBeNil)
		//		So(resp.Code, ShouldEqual, errf.RecordNotFound)
		//	}
		//})
		//
		//// Because the previous test case has already verified the return parameters, the ce s in the
		//// following scenario only verifies that the version number returned is correct

		//Convey("3. Normal Mode App Match Default Strategy", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId: cases.TBizID,
		//		AppId: BaseNormalTestAppID,
		//		Uid:   "6512bd43d9caa6e02c990b0a82652dca",
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNMDefaultStrategyReleaseID)
		//})
		//
		//Convey("4. Normal Mode App Match Main Strategy", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId: cases.TBizID,
		//		AppId: BaseNormalTestAppID,
		//		Uid:   "6512bd43d9caa6e02c990b0a82652dca",
		//		Labels: map[string]string{
		//			"os": "windows",
		//		},
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNMMainStrategyReleaseID)
		//})
		//
		//Convey("5. Normal Mode App Match Sub Strategy", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId: cases.TBizID,
		//		AppId: BaseNormalTestAppID,
		//		Uid:   "6512bd43d9caa6e02c990b0a82652dca",
		//		Labels: map[string]string{
		//			"os":   "windows",
		//			"city": "shenzhen",
		//		},
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNMSubStrategyReleaseID)
		//})
		//
		//Convey("6. Normal Mode App Match Instance Publish", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId: cases.TBizID,
		//		AppId: BaseNormalTestAppID,
		//		Uid:   InstanceUID,
		//		Labels: map[string]string{
		//			"os":   "windows",
		//			"city": "shenzhen",
		//		},
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNMInstancePublishReleaseID)
		//})
		//
		//Convey("7. Namespace Mode App Match Default Strategy", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId:     cases.TBizID,
		//		AppId:     BaseNamespaceTestAppID,
		//		Uid:       "6512bd43d9caa6e02c990b0a82652dca",
		//		Namespace: "xxxxxxxxx",
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNSDefaultStrategyReleaseID)
		//})
		//
		//Convey("8. Namespace Mode App Match Namespace", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId:     cases.TBizID,
		//		AppId:     BaseNamespaceTestAppID,
		//		Uid:       "6512bd43d9caa6e02c990b0a82652dca",
		//		Namespace: BNSNamespace,
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNSNamespaceReleaseID)
		//})
		//
		//Convey("9. Namespace Mode App Match Sub Strategy", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId:     cases.TBizID,
		//		AppId:     BaseNamespaceTestAppID,
		//		Uid:       "6512bd43d9caa6e02c990b0a82652dca",
		//		Namespace: BNSNamespace,
		//		Labels: map[string]string{
		//			"city": "beijing",
		//		},
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNSSubStrategyReleaseID)
		//})
		//
		//Convey("10. Namespace Mode App Match Instance Publish", func() {
		//	req := &types.ListFileAppLatestReleaseMetaReq{
		//		BizId:     cases.TBizID,
		//		AppId:     BaseNamespaceTestAppID,
		//		Uid:       InstanceUID,
		//		Namespace: BNSNamespace,
		//		Labels: map[string]string{
		//			"city": "shenzhen",
		//		},
		//	}
		//	ctx, header := cases.GenApiCtxHeader()
		//	resp, err := cli.FeedClient.ListAppFileLatestRelease(ctx, header, req)
		//	So(err, ShouldBeNil)
		//	So(resp.Code, ShouldEqual, errf.OK)
		//	So(resp.Data.ReleaseId, ShouldEqual, BNSInstancePublishReleaseID)
		//})
	})
}
