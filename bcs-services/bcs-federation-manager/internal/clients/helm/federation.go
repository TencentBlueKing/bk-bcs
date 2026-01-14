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

// Package helm xxx
package helm

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm/values"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/kubeconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/types"
)

// GetFederationCharts get federation charts
func (h *helmClient) GetFederationCharts() *types.FederationCharts {
	return h.opts.Charts
}

// IsInstalledForFederation check if federation modules is installed
func (h *helmClient) IsInstalledForFederation(ctx context.Context, opt *ReleaseBaseOptions) (bool, error) {

	var (
		err         error
		isInstalled bool
		helmOpts    = &HelmOptions{
			ProjectID: opt.ProjectID,
			ClusterID: opt.ClusterID,
		}
	)

	// charts for checking installed
	charts := []*types.Chart{
		h.opts.Charts.ClusternetHub,
		h.opts.Charts.Scheduler,
		h.opts.Charts.Controller,
		h.opts.Charts.Apiserver,
	}

	for _, chart := range charts {
		helmOpts.Namespace = chart.ReleaseNamespace
		helmOpts.ReleaseName = chart.ReleaseName
		isInstalled, err = h.IsInstalled(h.getMetadataCtx(ctx), helmOpts)
		if err != nil {
			return false, fmt.Errorf("check federation modules is installed failed, err: %v", err)
		}
		if isInstalled {
			return true, nil
		}
	}

	return false, nil
}

// InstallClusternetHub install bcs-clusternet-hub
func (h *helmClient) InstallClusternetHub(ctx context.Context, opt *ReleaseBaseOptions) error {
	if h.opts.Charts.ClusternetHub == nil {
		return fmt.Errorf("no clusternet hub chart found")
	}
	chart := h.opts.Charts.ClusternetHub

	valuesAll := []string{}
	if chart.DefaultValues != "" {
		valuesAll = append(valuesAll, chart.DefaultValues)
	}

	return h.installFederationChart(h.getMetadataCtx(ctx), opt, chart, valuesAll...)
}

// UninstallClusternetHub( uninstall bcs-clusternet-hub
func (h *helmClient) UninstallClusternetHub(ctx context.Context, opt *ReleaseBaseOptions) error {
	helmOpt := &HelmOptions{
		ProjectID:   opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   h.opts.Charts.ClusternetHub.ReleaseNamespace,
		ReleaseName: h.opts.Charts.ClusternetHub.ReleaseName,
	}

	isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpt)
	if err != nil {
		return err
	}
	if !isInstalled {
		return nil
	}
	return h.UninstallRelease(ctx, helmOpt)
}

// InstallClusternetScheduler install bcs-clusternet-scheduler
func (h *helmClient) InstallClusternetScheduler(ctx context.Context, opt *ReleaseBaseOptions) error {
	if h.opts.Charts.Scheduler == nil {
		return fmt.Errorf("no clusternet scheduler chart found")
	}
	chart := h.opts.Charts.Scheduler

	valuesAll := []string{}
	if chart.DefaultValues != "" {
		valuesAll = append(valuesAll, chart.DefaultValues)
	}

	return h.installFederationChart(h.getMetadataCtx(ctx), opt, chart, valuesAll...)
}

// UninstallClusternetScheduler uninstall bcs-clusternet-scheduler
func (h *helmClient) UninstallClusternetScheduler(ctx context.Context, opt *ReleaseBaseOptions) error {
	helmOpt := &HelmOptions{
		ProjectID:   opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   h.opts.Charts.Scheduler.ReleaseNamespace,
		ReleaseName: h.opts.Charts.Scheduler.ReleaseName,
	}

	isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpt)
	if err != nil {
		return err
	}
	if !isInstalled {
		return nil
	}
	return h.UninstallRelease(h.getMetadataCtx(ctx), helmOpt)
}

// InstallClusternetController install bcs-clusternet-controller
func (h *helmClient) InstallClusternetController(ctx context.Context, opt *ReleaseBaseOptions) error {
	if h.opts.Charts.Controller == nil {
		return fmt.Errorf("no clusternet controller chart found")
	}
	chart := h.opts.Charts.Controller

	valuesAll := []string{}
	if chart.DefaultValues != "" {
		valuesAll = append(valuesAll, chart.DefaultValues)
	}

	return h.installFederationChart(h.getMetadataCtx(ctx), opt, chart, valuesAll...)
}

// UninstallClusternetController uninstall bcs-clusternet-controller
func (h *helmClient) UninstallClusternetController(ctx context.Context, opt *ReleaseBaseOptions) error {
	helmOpt := &HelmOptions{
		ProjectID:   opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   h.opts.Charts.Controller.ReleaseNamespace,
		ReleaseName: h.opts.Charts.Controller.ReleaseName,
	}

	isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpt)
	if err != nil {
		return err
	}
	if !isInstalled {
		return nil
	}
	return h.UninstallRelease(h.getMetadataCtx(ctx), helmOpt)
}

// InstallUnifiedApiserver install bcs-unified-apiserver
func (h *helmClient) InstallUnifiedApiserver(ctx context.Context, opt *BcsUnifiedApiserverOptions) error {
	if h.opts.Charts.Apiserver == nil {
		return fmt.Errorf("no unified apiserver chart found")
	}
	chart := h.opts.Charts.Apiserver

	// default values
	valuesAll := []string{}
	if chart.DefaultValues != "" {
		valuesAll = append(valuesAll, chart.DefaultValues)
	}

	// todo add cert and key into values
	// render moduleValues
	moduleValues := values.NewBcsUnifiedAPIServerValues()

	// set federation host cluster id
	if err := moduleValues.SetFederationClusterId(opt.ReleaseBaseOptions.ClusterID); err != nil {
		return err
	}

	// set user token, which is used to query host cluster and sub cluster
	if err := moduleValues.SetUserToken(opt.UserToken); err != nil {
		return err
	}

	// set federation host cluster id
	if err := moduleValues.SetLoadbalancerId(opt.LoadBalancerId); err != nil {
		return err
	}

	valuesAll = append(valuesAll, moduleValues.Yaml())

	return h.installFederationChart(h.getMetadataCtx(ctx), &opt.ReleaseBaseOptions, chart, valuesAll...)
}

// UninstallUnifiedApiserver uninstall bcs-unified-apiserver
func (h *helmClient) UninstallUnifiedApiserver(ctx context.Context, opt *BcsUnifiedApiserverOptions) error {
	helmOpt := &HelmOptions{
		ProjectID:   opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   h.opts.Charts.Apiserver.ReleaseNamespace,
		ReleaseName: h.opts.Charts.Apiserver.ReleaseName,
	}

	isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpt)
	if err != nil {
		return err
	}
	if !isInstalled {
		return nil
	}
	return h.UninstallRelease(ctx, helmOpt)
}

// InstallClusternetAgent install bcs-clusternet-agent
func (h *helmClient) InstallClusternetAgent(ctx context.Context, opt *BcsClusternetAgentOptions) error {
	if h.opts.Charts.ClusternetAgent == nil {
		return fmt.Errorf("no clusternet agent chart found")
	}
	chart := &types.Chart{
		ChartVersion:     h.opts.Charts.ClusternetAgent.ChartVersion,
		ChartName:        h.opts.Charts.ClusternetAgent.ChartName,
		ReleaseNamespace: h.opts.Charts.ClusternetAgent.ReleaseNamespace,
		ReleaseName:      formatFederationReleaseName(h.opts.Charts.ClusternetAgent.ReleaseName, opt.SubClusterId),
		DefaultValues:    h.opts.Charts.ClusternetAgent.DefaultValues,
	}

	// default values
	valuesAll := []string{}
	if chart.DefaultValues != "" {
		valuesAll = append(valuesAll, chart.DefaultValues)
	}

	// render moduleValues
	serverAddress := fmt.Sprintf("%s/clusters/%s", opt.BcsGateWayAddress, opt.SubClusterId)
	cfg := kubeconfig.NewConfigForProvider(serverAddress, opt.UserToken, opt.SubClusterId)

	moduleValues := values.NewBcsClusternetAgentValues(opt.SubClusterId)
	moduleValues.SetKubeConfig(cfg.Yaml())
	moduleValues.SetRegistrationToken(opt.RegistrationToken)

	valuesAll = append(valuesAll, moduleValues.Yaml())

	return h.installFederationChart(h.getMetadataCtx(ctx), &opt.ReleaseBaseOptions, chart, valuesAll...)
}

// UnInstallClusternetAgent uninstall bcs-clusternet-agent
func (h *helmClient) UninstallClusternetAgent(ctx context.Context, opt *BcsClusternetAgentOptions) error {
	helmOpt := &HelmOptions{
		ProjectID:   opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   h.opts.Charts.ClusternetAgent.ReleaseNamespace,
		ReleaseName: formatFederationReleaseName(h.opts.Charts.ClusternetAgent.ReleaseName, opt.SubClusterId),
	}

	isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpt)
	if err != nil {
		return err
	}
	if !isInstalled {
		return nil
	}
	return h.UninstallRelease(h.getMetadataCtx(ctx), helmOpt)
}

// InstallEstimatorAgent install bcs-estimator-agent
func (h *helmClient) InstallEstimatorAgent(ctx context.Context, opt *BcsEstimatorAgentOptions) error {
	if h.opts.Charts.EstimatorAgent == nil {
		return fmt.Errorf("no estimator agent chart found")
	}
	chart := &types.Chart{
		ChartVersion:     h.opts.Charts.EstimatorAgent.ChartVersion,
		ChartName:        h.opts.Charts.EstimatorAgent.ChartName,
		ReleaseNamespace: h.opts.Charts.EstimatorAgent.ReleaseNamespace,
		ReleaseName:      formatFederationReleaseName(h.opts.Charts.EstimatorAgent.ReleaseName, opt.SubClusterId),
		DefaultValues:    h.opts.Charts.EstimatorAgent.DefaultValues,
	}

	// default values
	valuesAll := []string{}
	if chart.DefaultValues != "" {
		valuesAll = append(valuesAll, chart.DefaultValues)
	}

	// render moduleValues
	serverAddress := fmt.Sprintf("%s/clusters/%s", opt.BcsGateWayAddress, opt.SubClusterId)
	cfg := kubeconfig.NewConfigForProvider(serverAddress, opt.UserToken, opt.SubClusterId)

	moduleValues := values.NewBcsEstimatorAgentValues(opt.SubClusterId)
	moduleValues.SetKubeConfig(cfg.Yaml())

	valuesAll = append(valuesAll, moduleValues.Yaml())

	// skip when existed
	opt.ReleaseBaseOptions.SkipWhenExisted = true

	return h.installFederationChart(h.getMetadataCtx(ctx), &opt.ReleaseBaseOptions, chart, valuesAll...)
}

// UnInstallEstimatorAgent uninstall bcs-estimator-agent
func (h *helmClient) UninstallEstimatorAgent(ctx context.Context, opt *BcsEstimatorAgentOptions) error {
	helmOpt := &HelmOptions{
		ProjectID:   opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   h.opts.Charts.EstimatorAgent.ReleaseNamespace,
		ReleaseName: formatFederationReleaseName(h.opts.Charts.EstimatorAgent.ReleaseName, opt.SubClusterId),
	}
	isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpt)
	if err != nil {
		return err
	}
	if !isInstalled {
		return nil
	}
	return h.UninstallRelease(h.getMetadataCtx(ctx), helmOpt)
}

func (h *helmClient) installFederationChart(ctx context.Context, opt *ReleaseBaseOptions, chart *types.Chart, values ...string) error {
	helmOpts := &HelmOptions{
		ProjectID:    opt.ProjectID,
		ClusterID:    opt.ClusterID,
		Namespace:    chart.ReleaseNamespace,
		ReleaseName:  chart.ReleaseName,
		ChartName:    chart.ChartName,
		ChartVersion: chart.ChartVersion,
		IsPublic:     true,
	}

	if opt.SkipWhenExisted {
		isInstalled, err := h.IsInstalled(h.getMetadataCtx(ctx), helmOpts)
		if err != nil {
			return err
		}
		if isInstalled {
			blog.Infof("release %s already installed in cluster: %s, skip install", helmOpts.ReleaseName, helmOpts.ClusterID)
			return nil
		}
	}

	return h.InstallRelease(h.getMetadataCtx(ctx), helmOpts, values...)
}

func formatFederationReleaseName(prefix, clusterId string) string {
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(clusterId))
}
