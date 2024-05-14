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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
)

// MetricCollect defines the collector of metric
type MetricCollect struct {
	op *options.AnalysisOptions
	db dao.Interface

	monitorClient *monitoring.Clientset
	k8sClient     *kubernetes.Clientset
	metricQuery   *metric.ServiceMonitorQuery
}

// NewMetricCollect create the metric collector
func NewMetricCollect() *MetricCollect {
	return &MetricCollect{
		op: options.GlobalOptions(),
		db: dao.GlobalDB(),
	}
}

// Init the in-cluster client
func (c *MetricCollect) Init() error {
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
	blog.Infof("metric collector init success")
	return nil
}

// Start the metric collector
func (c *MetricCollect) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	blog.Infof("metric collection started")
	defer ticker.Stop()
	defer blog.Infof("analysis metric collection finished")
	for {
		select {
		case <-ticker.C:
			go c.collectApplicationMetric(ctx)
		case <-ctx.Done():
			blog.Infof("metric collector stopped")
			return
		}
	}
}

func (c *MetricCollect) collectApplicationMetric(ctx context.Context) {
	syncInfos := c.parseAppMetrics(ctx)
	blog.Infof("analysis parse application metrics succeed: %d", len(syncInfos))
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
func (c *MetricCollect) parseAppMetrics(ctx context.Context) []*dao.SyncInfo {
	ns := c.op.AppMetricNamespace
	name := c.op.AppMetricName
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
