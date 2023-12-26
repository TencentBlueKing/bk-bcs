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

// Package release NOTES
package release

import (
	"fmt"
	"time"

	"golang.org/x/time/rate"

	clientset "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/eventc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/lcache"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcommit "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbci "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbhook "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/hook"
	pbkv "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
)

// New initialize the release service instance.
func New(cs *clientset.ClientSet, cache *lcache.Cache, w eventc.Watcher) (*ReleasedService, error) {
	provider, err := repository.NewProvider(cc.FeedServer().Repository)
	if err != nil {
		return nil, fmt.Errorf("init repository provider failed, err: %v", err)
	}

	limiter := cc.FeedServer().MRLimiter
	return &ReleasedService{
		cs:                   cs,
		cache:                cache,
		provider:             provider,
		watcher:              w,
		wait:                 initWait(),
		limiter:              rate.NewLimiter(rate.Limit(limiter.QPS), int(limiter.Burst)),
		matchReleaseWaitTime: time.Duration(limiter.WaitTimeMil) * time.Millisecond,
	}, nil
}

// ReleasedService defines release related operations.
type ReleasedService struct {
	cs                   *clientset.ClientSet
	cache                *lcache.Cache
	provider             repository.Provider
	watcher              eventc.Watcher
	wait                 *waitShutdown
	limiter              *rate.Limiter
	matchReleaseWaitTime time.Duration
}

// ListAppLatestReleaseMeta list a app's latest release metadata
func (rs *ReleasedService) ListAppLatestReleaseMeta(kt *kit.Kit, opts *types.AppInstanceMeta) (
	*types.AppLatestReleaseMeta, error) {

	releaseID, err := rs.GetMatchedRelease(kt, opts)
	if err != nil {
		return nil, err
	}

	rci, err := rs.cache.ReleasedCI.Get(kt, opts.BizID, releaseID)
	if err != nil {
		return nil, err
	}

	pre, post, err := rs.cache.ReleasedHook.Get(kt, opts.BizID, releaseID)
	if err != nil {
		return nil, err
	}

	uriDec := rs.provider.URIDecorator(opts.BizID)
	meta := &types.AppLatestReleaseMeta{
		ReleaseId: releaseID,
		Repository: &types.Repository{
			Root: uriDec.Root(),
		},
	}
	if pre != nil {
		meta.PreHook = &pbhook.HookSpec{
			Type:    pre.Type.String(),
			Content: pre.Content,
		}
	}
	if post != nil {
		meta.PostHook = &pbhook.HookSpec{
			Type:    post.Type.String(),
			Content: post.Content,
		}
	}
	ciList := make([]*types.ReleasedCIMeta, len(rci))
	for idx, one := range rci {
		ciList[idx] = &types.ReleasedCIMeta{
			RciId:    one.ID,
			CommitID: one.CommitID,
			CommitSpec: &pbcommit.CommitSpec{
				ContentId: one.CommitSpec.ContentID,
				Content: &pbcontent.ContentSpec{
					Signature: one.CommitSpec.Signature,
					ByteSize:  one.CommitSpec.ByteSize,
				},
			},
			ConfigItemSpec: &pbci.ConfigItemSpec{
				Name:     one.ConfigItemSpec.Name,
				Path:     one.ConfigItemSpec.Path,
				FileType: string(one.ConfigItemSpec.FileType),
				FileMode: string(one.ConfigItemSpec.FileMode),
				Permission: &pbci.FilePermission{
					User:      one.ConfigItemSpec.Permission.User,
					UserGroup: one.ConfigItemSpec.Permission.UserGroup,
					Privilege: one.ConfigItemSpec.Permission.Privilege,
				},
			},
			ConfigItemAttachment: &pbci.ConfigItemAttachment{
				BizId: one.Attachment.AppID,
				AppId: one.Attachment.AppID,
			},
			RepositorySpec: &types.RepositorySpec{Path: uriDec.Path(one.CommitSpec.Signature)},
		}
	}
	meta.ConfigItems = ciList

	return meta, nil
}

// ListAppLatestReleaseKvMeta list a app's latest release metadata
func (rs *ReleasedService) ListAppLatestReleaseKvMeta(kt *kit.Kit, opts *types.AppInstanceMeta) (
	*types.AppLatestReleaseKvMeta, error) {

	releaseID, err := rs.GetMatchedRelease(kt, opts)
	if err != nil {
		return nil, err
	}

	rkv, err := rs.cache.ReleasedKv.Get(kt, opts.BizID, releaseID)
	if err != nil {
		return nil, err
	}

	meta := &types.AppLatestReleaseKvMeta{
		ReleaseId: releaseID,
	}

	kvList := make([]*types.ReleasedKvMeta, len(rkv))
	for idx, one := range rkv {

		kvList[idx] = &types.ReleasedKvMeta{
			Key:    one.Key,
			KvType: one.KvType,
			KvAttachment: &pbkv.KvAttachment{
				BizId: one.Attachment.BizID,
				AppId: one.Attachment.AppID,
			},
		}
	}
	meta.Kvs = kvList

	return meta, nil
}
