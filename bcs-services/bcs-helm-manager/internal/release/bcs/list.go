/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcs

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
)

func (c *cluster) list(ctx context.Context, option release.ListOption) (int, []*release.Release, error) {
	clientSet := c.ensureSdkClient()

	results, err := clientSet.List(ctx, option)
	if err != nil {
		blog.Errorf("list helm release from cluster failed, %s, cluster: %s, namespace: %s",
			err.Error(), c.clusterID, option.Namespace)
		return 0, nil, err
	}

	releaseutil.Reverse(results, releaseutil.SortByDate)

	total := len(results)
	if option.Page > 0 && option.Size > 0 {
		results = filterIndex(int((option.Page-1)*option.Size), int(option.Size), results)
	}

	r := make([]*release.Release, 0, len(results))
	for _, item := range results {
		chartVersion := ""
		if item.Chart != nil && item.Chart.Metadata != nil {
			chartVersion = item.Chart.Metadata.Version
		}

		manifest := item.Manifest
		for _, v := range item.Hooks {
			manifest += "---\n" + v.Manifest
		}
		rl := &release.Release{
			Name:         item.Name,
			Namespace:    item.Namespace,
			Revision:     item.Version,
			ChartVersion: chartVersion,
			Hooks:        item.Hooks,
			Manifest:     item.Manifest,
		}
		if item.Info != nil {
			rl.Status = item.Info.Status.String()
			rl.Description = item.Info.Description
			rl.UpdateTime = item.Info.LastDeployed.Local().Format(common.TimeFormat)
		}
		if item.Chart != nil {
			rl.Chart = item.Chart.Name()
			rl.AppVersion = item.Chart.AppVersion()
		}
		r = append(r, rl)
	}

	return total, r, nil
}

// filterIndex handle the offset and limit from release.ListOption
// take from index offset to index offset+limit-1
func filterIndex(offset, limit int, release []*rspb.Release) []*rspb.Release {
	if offset >= len(release) {
		return nil
	}

	if limit < 0 {
		limit = 0
	}

	if offset+limit > len(release) {
		return release[offset:]
	}

	return release[offset : offset+limit]
}
