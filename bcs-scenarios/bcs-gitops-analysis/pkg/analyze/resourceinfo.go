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

package analyze

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
)

func (h *analysisHandler) getCacheResourceInfo() []ProjectResourceInfo {
	h.cacheLock.Lock()
	defer h.cacheLock.Unlock()
	result := append(make([]ProjectResourceInfo, 0, len(h.resourceInfo)), h.resourceInfo...)
	return result
}

// CollectResourceInfo collect resource info with multi-goroutines.
func (h *analysisHandler) CollectResourceInfo() {
	apps := h.storage.AllApplications()
	var parallel = 5
	var wg sync.WaitGroup
	wg.Add(parallel)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	blog.Infof("collect applications(%d) resource info started", len(apps))
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := idx; j < len(apps); j += parallel {
				app := apps[j]
				err := h.collectAppResourceInfo(ctx, app)
				if err != nil {
					blog.Errorf("analysis collect app '%s' resource info failed: %s", app.Name, err.Error())
					continue
				}
			}
		}(i)
	}
	wg.Wait()
	h.cacheResourceInfoData()
	blog.Infof("collect applications(%d) resource info success", len(apps))
}

// collectAppResourceInfo will get the resource-tree with application. Parse the resource-tree to
// collect the information which we need
func (h *analysisHandler) collectAppResourceInfo(ctx context.Context, app *v1alpha1.Application) error {
	resourceTree, err := h.storage.GetApplicationResourceTree(ctx, app.Name)
	if err != nil {
		return errors.Wrapf(err, "get  resource tree failed")
	}
	result := make(map[string]int)
	all := 0
	for i := range resourceTree.Nodes {
		node := resourceTree.Nodes[i]
		all++
		kind := strings.ToLower(node.Kind)
		if utils.StringsContainsOr(kind, "gamedeployment", "gamestatefulset") {
			result[ResourceInfoGameWorkload]++
			continue
		}
		if utils.StringsContainsOr(kind, "statefulset", "deployment", "job", "cronjob", "daemonset") {
			result[ResourceInfoWorkload]++
			continue
		}
		if kind == "pod" {
			result[ResourceInfoPod]++
			continue
		}
	}
	result[ResourceInfoAll] = all
	riJSON, _ := json.Marshal(result)
	if err = h.db.SaveOrUpdateResourceInfo(&dao.ResourceInfo{
		Project:     app.Spec.Project,
		Application: app.Name,
		Resources:   utils.SliceByteToString(riJSON),
	}); err != nil {
		return errors.Wrapf(err, "save resource info failed")
	}
	return nil
}

func (h *analysisHandler) cacheResourceInfoData() {
	ris, err := h.db.ListResourceInfosByProject(nil)
	if err != nil {
		blog.Errorf("analysis list resource infos failed: %s", err.Error())
		return
	}
	apps := h.storage.AllApplications()
	appMap := make(map[string]struct{})
	for _, app := range apps {
		appMap[app.Name] = struct{}{}
	}
	result := make(map[string]*ProjectResourceInfo)
	for i := range ris {
		ri := ris[i]
		_, ok := appMap[ri.Application]
		if !ok {
			continue
		}
		m := make(map[string]int64)
		if err = json.Unmarshal(utils.StringToSliceByte(ri.Resources), &m); err != nil {
			continue
		}
		pri, ok := result[ri.Project]
		if !ok {
			result[ri.Project] = &ProjectResourceInfo{
				Name:         ri.Project,
				ResourceAll:  m[ResourceInfoAll],
				GameWorkload: m[ResourceInfoGameWorkload],
				Workload:     m[ResourceInfoWorkload],
				Pod:          m[ResourceInfoPod],
			}
		} else {
			pri.ResourceAll += m[ResourceInfoAll]
			pri.GameWorkload += m[ResourceInfoGameWorkload]
			pri.Workload += m[ResourceInfoWorkload]
			pri.Pod += m[ResourceInfoPod]
		}
	}
	data := make([]ProjectResourceInfo, 0, len(result))
	for _, pri := range result {
		data = append(data, *pri)
	}

	h.cacheLock.Lock()
	h.resourceInfo = data
	h.cacheLock.Unlock()
}
