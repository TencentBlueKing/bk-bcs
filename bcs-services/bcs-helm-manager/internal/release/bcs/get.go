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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
)

func (c *cluster) get(ctx context.Context, op release.GetOption) (*release.Release, error) {
	rl, err := c.ensureSdkClient().Get(ctx, op.Namespace, op.Name, op.Revision)
	if err != nil {
		blog.Errorf("get helm release from cluster failed, %s, cluster: %s, namespace: %s, name: %s",
			err.Error(), c.clusterID, op.Namespace, op.Name)
		return nil, err
	}

	chartVersion := ""
	if rl.Chart.Metadata != nil {
		chartVersion = rl.Chart.Metadata.Version
	}
	values := chartutil.Values(rl.Config)
	valuesYaml, _ := values.YAML()
	re := &release.Release{
		Name:         rl.Name,
		Namespace:    rl.Namespace,
		Revision:     rl.Version,
		Status:       rl.Info.Status.String(),
		Chart:        rl.Chart.Name(),
		ChartVersion: chartVersion,
		AppVersion:   rl.Chart.AppVersion(),
		UpdateTime:   rl.Info.LastDeployed.UTC().Format(time.RFC3339),
		Description:  rl.Info.Description,
		Values:       valuesYaml,
		Manifest:     rl.Manifest,
		Hooks:        rl.Hooks,
		Notes:        rl.Info.Notes,
	}
	if op.GetObject {
		re.Objects = make([]runtime.Object, 0)
		infos, err := ManifestToK8sResources(op.Namespace, re.Manifest, c.sdkClientGroup.Config(c.clusterID))
		if err != nil {
			return re, err
		}
		re.Infos = infos
		for _, v := range infos {
			re.Objects = append(re.Objects, v.Object)
		}
	}
	return re, nil
}
