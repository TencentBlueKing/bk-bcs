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

// Package analyze xx
package analyze

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
)

// AnalysisCollection 用来采集运营数据的相关信息
// - 采集 argocd metrics: 获取到 argo metric 后，将数据存储到数据库中进行打点记录
// - 更新活跃用户信息
type AnalysisCollection interface {
	Init() error
	Start(ctx context.Context) error
	UpdateActivityUser(project, user string)
}

type activityUserItem struct {
	Project string
	User    string
}

type analysisClient struct {
	db               dao.Interface
	metricConfig     *common.MetricConfig
	monitorClient    *monitoring.Clientset
	k8sClient        *kubernetes.Clientset
	metricQuery      *metric.ServiceMonitorQuery
	activityUserChan chan *activityUserItem
}

var (
	globalAnalysisClient *analysisClient
)

// NewAnalysisClient create the analysis client
func NewAnalysisClient() AnalysisCollection {
	globalAnalysisClient = &analysisClient{
		db:               dao.GlobalDB(),
		metricConfig:     options.GlobalOptions().MetricConfig,
		activityUserChan: make(chan *activityUserItem, 10000),
	}
	return globalAnalysisClient
}

// GetAnalysisClient return the global analysis client
func GetAnalysisClient() AnalysisCollection {
	return globalAnalysisClient
}

// Init the in-cluster client
func (c *analysisClient) Init() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get k8s incluster config failed")
	}
	c.monitorClient, err = monitoring.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create prometheus client failed")
	}
	c.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create k8s lient failed")
	}
	c.metricQuery = &metric.ServiceMonitorQuery{
		MonitorClient: c.monitorClient,
		K8sClient:     c.k8sClient,
	}
	return nil
}

// Start the for-select to handle activity user and metrics
func (c *analysisClient) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go c.handleAppMetric(ctx)
		case userItem := <-c.activityUserChan:
			c.handleActivityUser(userItem)
		case <-ctx.Done():
			blog.Warnf("analysis closed")
			return nil
		}
	}
}

// UpdateActivityUser update activity user, it will add operation num into db
func (c *analysisClient) UpdateActivityUser(project, user string) {
	c.activityUserChan <- &activityUserItem{
		Project: project,
		User:    user,
	}
}

func (c *analysisClient) handleActivityUser(item *activityUserItem) {
	activityUser, err := c.db.GetActivityUser(item.Project, item.User)
	if err != nil {
		blog.Errorf("[analysis] get activity user '%s/%s' failed: %s", item.Project, item.User, err.Error())
		return
	}
	if activityUser == nil {
		activityUser = &dao.ActivityUser{
			Project:          item.Project,
			UserName:         item.User,
			OperateNum:       1,
			LastActivityTime: time.Now(),
		}
		if err = c.db.SaveActivityUser(activityUser); err != nil {
			blog.Errorf("[analysis] save activity user failed: %s", err.Error())
			return
		}
		return
	}
	activityUser.OperateNum++
	if err = c.db.UpdateActivityUser(activityUser); err != nil {
		blog.Errorf("[analysis] update activity user failed: %s", err.Error())
		return
	}
}

// handleAppMetric used to calculate the sync number of every application with cluster
func (c *analysisClient) handleAppMetric(ctx context.Context) {
	syncInfos := c.parseAppMetrics(ctx)
	blog.Infof("[analysis] parse application metrics succeed: %d", len(syncInfos))
	for _, item := range syncInfos {
		syncInfo, err := c.db.GetSyncInfo(item.Project, item.Cluster,
			item.Application, item.Phase)
		if err != nil {
			blog.Warnf("[analysis] get sync info '%v' failed", *item)
			continue
		}
		if syncInfo == nil {
			item.SyncTotal = item.PreviousSync
			if err = c.db.SaveSyncInfo(item); err != nil {
				blog.Errorf("[analysis] save sync info '%v' failed", *item)
			}
			continue
		}
		if item.PreviousSync < syncInfo.PreviousSync {
			syncInfo.SyncTotal += item.PreviousSync
		} else {
			syncInfo.SyncTotal += item.PreviousSync - syncInfo.PreviousSync
		}
		syncInfo.PreviousSync = item.PreviousSync
		if err = c.db.UpdateSyncInfo(syncInfo); err != nil {
			blog.Errorf("[analysis] update syncinfo '%v' failed", *syncInfo)
			continue
		}
	}
	blog.Infof("[analysis] handle app metric successful")
}

var (
	appMetricRegex     = regexp.MustCompile(`argocd_app_sync_total{(.*)}\s+(\d+)`)
	appMetricItemRegex = regexp.MustCompile(`(.*)="(.*)"`)
)

// parseAppMetrics will get the argocd application-controller's metrics, and then parse
// them to get the sync info of every application and cluster
// nolint funlen
func (c *analysisClient) parseAppMetrics(ctx context.Context) []*dao.SyncInfo {
	ns := c.metricConfig.AppMetricNamespace
	name := c.metricConfig.AppMetricName
	metrics, err := c.metricQuery.Do(ctx, ns, name)
	if err != nil {
		blog.Errorf("[analysis] query service monitor '%s/%s' failed: %s", ns, name, err.Error())
		return nil
	}
	result := make([]*dao.SyncInfo, 0)
	for _, metricStr := range metrics {
		if !strings.HasPrefix(metricStr, "argocd_app_sync_total") {
			continue
		}
		matches := appMetricRegex.FindStringSubmatch(metricStr)
		if len(matches) != 3 {
			blog.Warnf("[analysis] metric '%s' format error", metricStr)
			continue
		}
		var num int64
		num, err = strconv.ParseInt(matches[2], 0, 64)
		if err != nil {
			blog.Warnf("[analysis] metric '%s' num format error", metricStr)
			continue
		}

		syncInfo := &dao.SyncInfo{
			PreviousSync: num,
		}
		items := strings.Split(matches[1], ",")
		for _, item := range items {
			itemMatches := appMetricItemRegex.FindStringSubmatch(item)
			if len(itemMatches) != 3 {
				blog.Warnf("[analysis] metric '%s' find items length failed", metricStr)
				continue
			}
			switch itemMatches[1] {
			case "dest_server":
				tmp := strings.Split(itemMatches[2], "/")
				syncInfo.Cluster = tmp[len(tmp)-1]
			case "name":
				syncInfo.Application = itemMatches[2]
			case "phase":
				syncInfo.Phase = itemMatches[2]
			case "project":
				syncInfo.Project = itemMatches[2]
			}
		}
		if syncInfo.Cluster == "" || syncInfo.Application == "" || syncInfo.Phase == "" || syncInfo.Project == "" {
			blog.Warnf("[analysis] metric '%s' sync info lost some message")
			continue
		}
		result = append(result, syncInfo)
	}
	return result
}
