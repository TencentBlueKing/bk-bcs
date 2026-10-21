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

package bcs

import (
	"context"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
)

func (c *cluster) history(ctx context.Context, op release.HelmHistoryOption) ([]*release.Release, error) {
	rl, err := c.ensureSdkClient().History(ctx, op.Namespace, op.Name, op.Max)
	if err != nil {
		blog.Errorf("get helm release history from cluster failed, %s, cluster: %s, namespace: %s, name: %s",
			err.Error(), c.clusterID, op.Namespace, op.Name)
		return nil, err
	}

	result := make([]*release.Release, 0, len(rl))
	for _, v := range rl {
		chartVersion := ""
		if v.Chart.Metadata != nil {
			chartVersion = v.Chart.Metadata.Version
		}
		values := chartutil.Values(v.Config)
		valuesYaml, _ := values.YAML()
		result = append(result, &release.Release{
			Name:         v.Name,
			Namespace:    v.Namespace,
			Revision:     v.Version,
			Status:       v.Info.Status.String(),
			Chart:        v.Chart.Name(),
			ChartVersion: chartVersion,
			AppVersion:   v.Chart.AppVersion(),
			UpdateTime:   v.Info.LastDeployed.UTC().Format(time.RFC3339),
			Description:  v.Info.Description,
			Values:       valuesYaml,
		})
	}
	sort.Sort(release.ReleasesSlice(result))
	return result, nil
}
