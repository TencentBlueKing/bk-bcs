/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package release NOTES
package release

import (
	"fmt"
	"time"

	clientset "bscp.io/cmd/feed-server/bll/client-set"
	"bscp.io/cmd/feed-server/bll/eventc"
	"bscp.io/cmd/feed-server/bll/lcache"
	"bscp.io/cmd/feed-server/bll/types"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/kit"
	pbcommit "bscp.io/pkg/protocol/core/commit"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	"bscp.io/pkg/thirdparty/repo"

	"golang.org/x/time/rate"
)

// New initialize the release service instance.
func New(cs *clientset.ClientSet, cache *lcache.Cache, w eventc.Watcher) (*ReleasedService, error) {
	uriDecorator, err := repo.NewUriDecorator(cc.FeedServer().Repository)
	if err != nil {
		return nil, fmt.Errorf("init repository uri decorator failed, err: %v", err)
	}

	limiter := cc.FeedServer().MRLimiter
	return &ReleasedService{
		cs:                   cs,
		cache:                cache,
		uriDecorator:         uriDecorator,
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
	uriDecorator         repo.UriDecoratorInter
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

	uriDec := rs.uriDecorator.Init(opts.BizID)
	meta := &types.AppLatestReleaseMeta{
		ReleaseId: releaseID,
		Repository: &types.Repository{
			Root: uriDec.Root(),
		},
	}
	ciList := make([]*types.ReleasedCIMeta, len(rci))
	for idx, one := range rci {
		ciList[idx] = &types.ReleasedCIMeta{
			RciId: one.ID,
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
			RepositorySpec: &types.RepositorySpec{Path: uriDec.Path(one.CommitSpec.Signature)},
		}
	}
	meta.ConfigItems = ciList

	return meta, nil
}
