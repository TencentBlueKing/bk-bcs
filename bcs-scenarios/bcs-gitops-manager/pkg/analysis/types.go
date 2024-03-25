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

// Package analysis xxx
package analysis

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
)

// AnalysisInterface defines the analysis data interface
// nolint
type AnalysisInterface interface {
	Init() error
	Start(ctx context.Context) error

	ListSyncs(proj string) ([]dao.SyncInfo, error)
	ListActivityUsers(proj string) ([]dao.ActivityUser, error)
	UpdateActivityUser(project, user string)
	ListApplicationCollects(project string) ([]*dao.ResourcePreference, error)
	ApplicationCollect(project, name string) error
	ApplicationCancelCollect(project, name string) error
}

// AnalysisClient defines the client that to handle analysis data
// nolint
type AnalysisClient struct {
	db           dao.Interface
	metricConfig *common.MetricConfig

	monitorClient *monitoring.Clientset
	k8sClient     *kubernetes.Clientset

	metricQuery      *metric.ServiceMonitorQuery
	activityUserChan chan *activityUserItem
}

// NewAnalysisClient create analysis client
func NewAnalysisClient(db dao.Interface, metricConfig *common.MetricConfig) AnalysisInterface {
	analysisClient = &AnalysisClient{
		db:               db,
		metricConfig:     metricConfig,
		activityUserChan: make(chan *activityUserItem, 10000),
	}
	return analysisClient
}

var (
	analysisClient AnalysisInterface
)

// GetAnalysisClient return the global analysis client
func GetAnalysisClient() AnalysisInterface {
	return analysisClient
}

// Init the in-cluster client
func (c *AnalysisClient) Init() error {
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
func (c *AnalysisClient) Start(ctx context.Context) error {
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
		}
	}
}
