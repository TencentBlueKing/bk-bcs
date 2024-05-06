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

// Package external xx
package external

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/analyze"
)

type analysisHandler struct {
	op *options.AnalysisOptions
}

// NewExternalAnalysisHandler create external analysis handler
func NewExternalAnalysisHandler() analyze.AnalysisInterface {
	blog.Infof("analysis external handler opened")
	return &analysisHandler{
		op: options.GlobalOptions(),
	}
}

const (
	overviewUrl = "/gitopsmanager/proxy/api/v1/analysis/overview"
)

type analysisOverviewResp struct {
	Code    int                            `json:"code"`
	Message string                         `json:"message"`
	Data    []*analyze.AnalysisOverviewAll `json:"data"`
}

// AnalysisOverview return the external analysis overview
func (h *analysisHandler) AnalysisOverview() (*analyze.AnalysisOverviewAll, error) {
	body, err := h.queryExternalOverview(overviewUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "query external analysis overview failed")
	}
	resp := new(analysisOverviewResp)
	if err = json.Unmarshal(body, resp); err != nil {
		return nil, errors.Wrapf(err, "unmarshal external analysis overview body failed")
	}
	if len(resp.Data) == 0 {
		return nil, errors.Errorf("external analysis not have response")
	}
	return resp.Data[0], nil
}

// Init not implement
func (h *analysisHandler) Init() error {
	return nil
}

// QueryArgoProjects not implement
func (h *analysisHandler) QueryArgoProjects(projs []string) ([]v1alpha1.AppProject, error) {
	return nil, nil
}

// AnalysisProjectsAll not implement
func (h *analysisHandler) AnalysisProjectsAll() []*analyze.AnalysisProject {
	return nil
}

// Applications not implement
func (h *analysisHandler) Applications() ([]*analyze.ApplicationInfo, error) {
	return nil, nil
}

// GetBusinessName not implement
func (h *analysisHandler) GetBusinessName(bizID int) string {
	return ""
}

// AnalysisProject not implement
func (h *analysisHandler) AnalysisProject(ctx context.Context, argoProjs []v1alpha1.AppProject) (
	[]analyze.AnalysisProject, error) {
	return nil, nil
}

// ResourceInfosAll not implement
func (h *analysisHandler) ResourceInfosAll() []analyze.ProjectResourceInfo {
	return nil
}

// TopProjects not implement
func (h *analysisHandler) TopProjects() []*analyze.AnalysisProjectOverview {
	return nil
}

func (h *analysisHandler) queryExternalOverview(urlPath string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, h.op.ExternalAnalysisUrl+urlPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "collect external data create request failed")
	}
	req.Header.Set("Authorization", "Bearer "+h.op.ExternalAnalysisToken)
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
		return nil, errors.Errorf("collect external data resp code not 200 but %d: %s",
			resp.StatusCode, string(bs))
	}
	return bs, nil
}
