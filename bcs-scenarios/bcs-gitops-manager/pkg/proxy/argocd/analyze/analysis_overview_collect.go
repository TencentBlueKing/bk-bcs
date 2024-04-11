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
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

func (c *analysisOverviewClient) refreshResourceInfoData() {
	ris, err := c.db.ListResourceInfosByProject(nil)
	if err != nil {
		blog.Errorf("analysis collect resource info init cache failed: %s", err.Error())
		return
	}
	apps := c.storage.AllApplications()
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

	c.cacheLock.Lock()
	c.resourceInfoData = data
	c.cacheLock.Unlock()
}

func (c *analysisOverviewClient) collectAllResourceInfo() {
	apps := c.storage.AllApplications()
	var parallel = 5
	var wg sync.WaitGroup
	wg.Add(parallel)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	blog.Infof("analysis collect applications(%d) resource info started", len(apps))
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := idx; j < len(apps); j += parallel {
				app := apps[j]
				err := c.collectAppResourceInfo(ctx, app)
				if err != nil {
					blog.Errorf("analysis collect app '%s' resource info failed: %s", app.Name, err.Error())
					continue
				}
			}
		}(i)
	}
	wg.Wait()
	c.refreshResourceInfoData()
	blog.Infof("analysis collect applications(%d) resource info success", len(apps))
}

const (
	// ResourceInfoGameWorkload defines the gameworkload field
	ResourceInfoGameWorkload = "gameworkload"
	// ResourceInfoWorkload defines the workload field
	ResourceInfoWorkload = "workload"
	// ResourceInfoPod defines the pod field
	ResourceInfoPod = "pod"
	// ResourceInfoAll defines the all field
	ResourceInfoAll = "all"
)

func (c *analysisOverviewClient) collectAppResourceInfo(ctx context.Context,
	app *v1alpha1.Application) error {
	resourceTree, err := c.storage.GetApplicationResourceTree(ctx, app.Name)
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
	if err = c.db.SaveOrUpdateResourceInfo(&dao.ResourceInfo{
		Project:     app.Spec.Project,
		Application: app.Name,
		Resources:   utils.SliceByteToString(riJSON),
	}); err != nil {
		return errors.Wrapf(err, "save resource info failed")
	}
	return nil
}

// collectAllOverviewData 采集所有的运营数据，并发送到 bkmonitor
func (c *analysisOverviewClient) collectAllOverviewData() {
	if err := c.collectInternalAllData(); err != nil {
		blog.Errorf("analysis collect internal data failed: %s", err.Error())
	} else {
		blog.Infof("analysis collect internal data success")
	}
	// no need to collect external data if sg_url is empty
	if c.op.AnalysisConfig.GitOpsAnalysisUrlSG == "" {
		return
	}

	if err := c.collectExternalAllData(); err != nil {
		blog.Errorf("analysis collect external data failed: %s", err.Error())
	} else {
		blog.Infof("analysis collect external data success")
	}
}

func (c *analysisOverviewClient) collectInternalAllData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	projList, err := c.storage.ListProjectsWithoutAuth(ctx)
	if err != nil {
		return errors.Wrapf(err, "list projects without auth failed")
	}
	result, err := c.AnalysisProject(ctx, projList.Items)
	if err != nil {
		return errors.Wrapf(err, "collect internal all data failed")
	}

	c.cacheLock.Lock()
	c.internalData = result
	c.cacheLock.Unlock()

	overviewAll := c.OverviewAllInternal()
	if overviewAll == nil {
		return nil
	}
	c.pushOverviewToBKMonitor(overviewAll, "internal")
	return nil
}

type analysisExternalResp struct {
	Code    int32             `json:"code"`
	Message string            `json:"message"`
	Data    []AnalysisProject `json:"data"`
}

func (c *analysisOverviewClient) collectExternalAllData() error {
	resp, err := c.queryExternalOverview()
	if err != nil {
		return errors.Wrapf(err, "query external overview failed")
	}
	c.cacheLock.Lock()
	c.externalData = resp.Data
	c.cacheLock.Unlock()

	overviewAll := c.OverviewAllExternal()
	if overviewAll == nil {
		return nil
	}
	c.pushOverviewToBKMonitor(overviewAll, "external")
	return nil
}

func (c *analysisOverviewClient) queryExternalOverview() (*analysisExternalResp, error) {
	req, err := http.NewRequest(http.MethodGet, c.op.AnalysisConfig.GitOpsAnalysisUrlSG, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "collect external data create request failed")
	}
	req.Header.Set("Authorization", "Bearer "+c.op.AnalysisConfig.GitOpsAnalysisTokenSG)
	req.Header.Set(common.HeaderBCSClient, common.ServiceNameShort)
	req.Header.Set(common.HeaderBKUserName, common.HeaderAdminClientUser)
	httpClient := http.DefaultClient
	httpClient.Timeout = 60 * time.Second
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "collect external data do request failed")
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "collect external data read body failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("collect external data resp code not 200 but %d: %s", resp.StatusCode, string(bs))
	}
	result := &analysisExternalResp{}
	if err = json.Unmarshal(bs, result); err != nil {
		return nil, errors.Wrapf(err, "collect external data unmarshal failed")
	}
	return result, nil
}

func (c *analysisOverviewClient) pushOverviewToBKMonitor(overviewAll *AnalysisOverviewAll, target string) {
	if !c.bkmClient.IsPushTurnOn() {
		return
	}
	bkmMessage := &bkMonitorMessage{
		DataID:      c.op.AnalysisConfig.BKMonitorPushDataID,
		AccessToken: c.op.AnalysisConfig.BKMonitorPushToken,
		Data: []*bkMonitorMessageData{
			{
				Metrics: map[string]interface{}{
					"effective_bizs":        overviewAll.EffectiveBizNum,
					"effective_projects":    overviewAll.EffectiveProjectNum,
					"effective_clusters":    overviewAll.EffectiveClusterNum,
					"applications":          overviewAll.ApplicationNum,
					"user_operates":         overviewAll.UserOperateNum,
					"application_syncs":     overviewAll.SyncTotal,
					"activity_1day_user":    overviewAll.Activity1DayUserNum,
					"activity_1day_project": overviewAll.Activity1DayProjectNum,
				},
				Dimension: map[string]string{},
				Target:    target,
				Timestamp: time.Now().UnixMilli(),
			},
		},
	}
	for proj, total := range overviewAll.ProjectSyncTotal {
		bkmMessage.Data = append(bkmMessage.Data, &bkMonitorMessageData{
			Metrics: map[string]interface{}{
				"project_sync": total,
			},
			Dimension: map[string]string{
				"project": proj,
			},
			Target:    target,
			Timestamp: time.Now().UnixMilli(),
		})
	}
	c.bkmClient.Push(bkmMessage)
}
