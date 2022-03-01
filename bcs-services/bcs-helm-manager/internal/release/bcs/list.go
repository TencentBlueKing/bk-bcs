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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"

	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
)

func (c *cluster) list(ctx context.Context, option release.ListOption) (int, []*release.Release, error) {
	clientSet := c.ensureSdkClient()

	results, err := clientSet.List(ctx, option.Namespace)
	if err != nil {
		blog.Errorf("list helm release from cluster failed, %s, cluster: %s, namespace: %s",
			err.Error(), c.clusterID, option.Namespace)
		return 0, nil, err
	}

	if option.Name != "" {
		results = filterNameReleases(option.Name, results)
	}
	releaseutil.SortByDate(results)

	total := len(results)
	results = filterIndex(int(option.Page*option.Size), int(option.Size), results)

	r := make([]*release.Release, 0, len(results))
	for _, item := range results {
		chartVersion := ""
		if item.Chart.Metadata != nil {
			chartVersion = item.Chart.Metadata.Version
		}

		r = append(r, &release.Release{
			Name:         item.Name,
			Namespace:    item.Namespace,
			Revision:     item.Version,
			Status:       item.Info.Status.String(),
			Chart:        item.Chart.Name(),
			ChartVersion: chartVersion,
			AppVersion:   item.Chart.AppVersion(),
			UpdateTime:   item.Info.LastDeployed.Local().String(),
		})
	}

	return total, r, nil
}

func filterNameReleases(name string, releases []*rspb.Release) []*rspb.Release {
	var list = make([]*rspb.Release, 0, len(releases))
	for _, rls := range releases {
		// if name is not empty, then should filter by it.
		if name == "" || name == rls.Name {
			list = append(list, rls)
		}
	}
	return list
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
