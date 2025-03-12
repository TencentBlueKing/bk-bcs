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

	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm/values"
)

// IsInstalled check if the release is installed
func (h *helmClient) IsInstalled(opt *HelmOptions) (bool, error) {
	resp, err := h.helmSvc.GetReleaseDetailV1(h.getMetadataCtx(context.Background()), &helmmanager.GetReleaseDetailV1Req{
		ProjectCode: opt.ProjectID, // 兼容ProjectCode和ProjectId
		ClusterID:   opt.ClusterID,
		Namespace:   opt.Namespace,
		Name:        opt.ReleaseName,
	})
	if err != nil {
		blog.Errorf("[HelmManager] GetReleaseDetail %s failed, err: %s", opt.ReleaseName, err.Error())
		return false, err
	}
	if resp == nil {
		err = fmt.Errorf("[HelmManager] GetReleaseDetail %s failed, resp is empty", opt.ReleaseName)
		blog.Errorf(err.Error())
		return false, err
	}
	// not found release
	if resp.Code != 0 {
		blog.Infof("[HelmManager] GetReleaseDetail %s failed, code: %d, message: %s", opt.ReleaseName, resp.Code, resp.Message)
		return false, nil
	}

	blog.Infof("[HelmManager] [%s:%s] GetReleaseDetail success[%s:%s] status: %s",
		resp.Data.Chart, resp.Data.ChartVersion, resp.Data.Namespace, resp.Data.Name, resp.Data.Status)

	return true, nil
}

// InstallRelease install release
func (h *helmClient) InstallRelease(opt *HelmOptions, helmValues ...string) error {
	if h.debug {
		blog.Infof("[HelmManager] Debug InstallRelease: %+v, values: %v", opt, helmValues)
		return nil
	}

	repo := PubicRepo

	// if chart version is empty, use latest version
	if opt.ChartVersion == "" {
		// if not specify version, use latest version
		v, err := h.getChartLatestVersion(opt.ProjectID, repo, opt.ChartName)
		if err != nil {
			blog.Errorf("[HelmManager] getChartLatestVersion failed: %v", err)
			return err
		}
		opt.ChartVersion = v
	}

	// if helmValues is empty, use default values of the chart
	value, err := values.MergeValues(helmValues...)
	if err != nil {
		return err
	}

	req := &helmmanager.InstallReleaseV1Req{
		ProjectCode: opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   opt.Namespace,
		Name:        opt.ReleaseName,
		Chart:       opt.ChartName,
		Repository:  repo,
		Version:     opt.ChartVersion,
		Values:      value,
		Args:        []string{DefaultArgsFlagInsecure, DefaultArgsFlagWait},
	}

	resp := &helmmanager.InstallReleaseV1Resp{}
	err = retry.Do(func() error {
		resp, iErr := h.helmSvc.InstallReleaseV1(h.getMetadataCtx(context.Background()), req)
		if iErr != nil {
			blog.Errorf("[HelmManager] InstallRelease failed, err: %s", err.Error())
			return iErr
		}
		if resp == nil {
			blog.Errorf("[HelmManager] InstallRelease failed, resp is empty")
			return fmt.Errorf("InstallRelease failed, resp is empty")
		}

		if resp.Code != 0 || !resp.Result {
			blog.Errorf("[HelmManager] InstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
			return fmt.Errorf("InstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
		}

		return nil
	}, retry.Attempts(DefaultRetryCount), retry.Delay(DefaultTimeout), retry.DelayType(retry.FixedDelay))
	if err != nil {
		return fmt.Errorf("call api HelmManager InstallRelease failed: %v, resp: %v", err, resp)
	}

	return nil
}

func (h *helmClient) getChartLatestVersion(project string, repo, chart string) (string, error) {
	resp, err := h.helmSvc.GetChartDetailV1(h.getMetadataCtx(context.Background()), &helmmanager.GetChartDetailV1Req{
		ProjectCode: project,
		RepoName:    repo,
		Name:        chart,
	})
	if err != nil {
		blog.Errorf("[HelmManager] getChartLatestVersion failed: %v", err)
		return "", err
	}

	if resp.Code != 0 || !resp.Result {
		blog.Errorf("[HelmManager] getChartLatestVersion[%s] failed: %v", resp.RequestID, resp.Message)
		return "", err
	}

	return resp.Data.LatestVersion, nil
}
