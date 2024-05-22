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

package collect

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/analyze"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/analyze/external"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/bkm"
)

const (
	envInternal = "internal"
	envExternal = "external"
)

// AnalysisCollect defines the collect for analysis
type AnalysisCollect struct {
	op                      *options.AnalysisOptions
	analysisHandler         analyze.AnalysisInterface
	externalAnalysisHandler analyze.AnalysisInterface
	bkmClient               *bkm.BKMonitorClient
}

// NewAnalysisCollector create the collector of analysis
func NewAnalysisCollector() *AnalysisCollect {
	collector := &AnalysisCollect{
		op:              options.GlobalOptions(),
		analysisHandler: analyze.GlobalAnalysisHandler(),
		bkmClient:       bkm.NewBKMonitorClient(),
	}
	if collector.op.ExternalAnalysisUrl != "" && collector.op.ExternalAnalysisToken != "" {
		collector.externalAnalysisHandler = external.NewExternalAnalysisHandler()
	}
	return collector
}

// Start the collector of analysis
func (c *AnalysisCollect) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	blog.Infof("analysis collector started")
	defer ticker.Stop()
	defer blog.Infof("analysis collection finished")
	for {
		select {
		case <-ticker.C:
			c.collect()
		case <-ctx.Done():
			return
		}
	}
}

func (c *AnalysisCollect) collect() {
	// nolint
	overview, _ := c.analysisHandler.AnalysisOverview()
	if overview != nil {
		c.pushOverviewToBKMonitor(overview, envInternal)
	}
	if c.externalAnalysisHandler == nil {
		return
	}
	externalOverview, err := c.externalAnalysisHandler.AnalysisOverview()
	if err != nil {
		blog.Errorf("analysis query external overview failed: %s", err.Error())
		return
	}
	c.pushOverviewToBKMonitor(externalOverview, envExternal)
}

func (c *AnalysisCollect) pushOverviewToBKMonitor(overviewAll *analyze.AnalysisOverviewAll, target string) {
	if !c.bkmClient.IsPushTurnOn() {
		return
	}
	if overviewAll == nil {
		return
	}
	bkmMessage := &bkm.BKMonitorMessage{
		DataID:      c.op.BKMonitorPushDataID,
		AccessToken: c.op.BKMonitorPushToken,
		Data: []*bkm.BKMonitorMessageData{
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
		bkmMessage.Data = append(bkmMessage.Data, &bkm.BKMonitorMessageData{
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
