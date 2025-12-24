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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	helmtime "helm.sh/helm/v3/pkg/time"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
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
			rl.UpdateTime = item.Info.LastDeployed.UTC().Format(time.RFC3339)
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

// listV2  只取release的特殊数据
func (c *cluster) listV2(ctx context.Context, option release.ListOption) (int, []*release.Release, error) {

	results, err := c.getReleases(ctx, option)
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
			rl.UpdateTime = item.Info.LastDeployed.UTC().Format(time.RFC3339)
		}
		if item.Chart != nil {
			rl.Chart = item.Chart.Name()
			rl.AppVersion = item.Chart.AppVersion()
		}
		r = append(r, rl)
	}

	return total, r, nil
}

// 获取release列表
func (c *cluster) getReleases(ctx context.Context, option release.ListOption) ([]*rspb.Release, error) {
	secretList, err := c.getSecretList(ctx, option)
	if err != nil {
		return nil, err
	}
	var releases []*rspb.Release
	// 主要针对Labels的内容进行解析
	for _, data := range secretList.Items {
		releases = append(releases, &rspb.Release{
			Name: data.Labels["name"],
			Info: &rspb.Info{
				LastDeployed: helmtime.Time{Time: parseStringUnixTime(data.Labels["modifiedAt"])},
				Status:       rspb.Status(data.Labels["status"])},
			Namespace: data.ObjectMeta.Namespace,
			Version:   parseVersion(data.Labels["version"]),
		})
	}
	return releases, nil
}

// 获取版本号，string->int
func parseVersion(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// Unix字符串转换
func parseStringUnixTime(s string) time.Time {
	unixTimestamp, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Now()
	}

	// 将 UNIX 时间戳转换为 time.Time 类型
	return time.Unix(unixTimestamp, 0)
}

// 获取自定义字段的secrete list
func (c *cluster) getSecretList(ctx context.Context, option release.ListOption) (*v1.SecretList, error) {
	apiServer := options.GlobalOptions.Release.APIServer
	bearerToken := options.GlobalOptions.Release.Token
	url := fmt.Sprintf("%s/clusters/%s/api/v1/namespaces/%s/secrets?%s", apiServer, c.clusterID, option.Namespace,
		"labelSelector=owner=helm,status!=superseded")
	header := http.Header{}
	header.Set("Authorization", "Bearer "+bearerToken)
	body, err := component.HttpRequest(ctx, url, http.MethodGet, header, nil)
	if err != nil {
		return nil, err
	}
	result := v1.SecretList{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
