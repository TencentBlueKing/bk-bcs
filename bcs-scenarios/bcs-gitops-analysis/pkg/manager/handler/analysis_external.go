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

package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/ccv3"
)

// AnalysisExternalHandler defines the analysis handler for external
type AnalysisExternalHandler struct {
	op                *options.AnalysisOptions
	bkccClient        ccv3.Interface
	cacheLock         sync.Mutex
	cache             []AnalysisProject
	bizDeptInfoCache  *sync.Map
	userDeptInfoCache *sync.Map
}

// NewAnalysisExternalHandler create AnalysisInterface instance
func NewAnalysisExternalHandler() AnalysisInterface {
	return &AnalysisExternalHandler{
		op:                options.GlobalOptions(),
		bizDeptInfoCache:  &sync.Map{},
		userDeptInfoCache: &sync.Map{},
	}
}

// Init the external handler with time tick collect analysis
func (h *AnalysisExternalHandler) Init() error {
	h.bkccClient = ccv3.NewHandler()
	go func() {
		analysisTicker := time.NewTicker(1 * time.Minute)
		defer analysisTicker.Stop()
		for {
			select {
			case <-analysisTicker.C:
				h.analysisProjects()
			}
		}
	}()
	return nil
}

func (h *AnalysisExternalHandler) analysisProjects() {
	projects, err := h.getExternalRawData()
	if err != nil {
		blog.Errorf("get external analysis failed: %s", err.Error())
		return
	}
	var parallel = 10
	var wg sync.WaitGroup
	wg.Add(parallel)
	result := make([]AnalysisProject, 0, len(projects))
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := idx; j < len(projects); j += parallel {
				proj := projects[j]
				if err = h.fillGroupLevel(&proj); err != nil {
					blog.Errorf("fill group level failed")
				}
				result = append(result, proj)
			}
		}(i)
	}
	wg.Wait()

	blog.Infof("new collect projects(%d) analysis data success", len(projects))
	h.cacheLock.Lock()
	h.cache = result
	h.cacheLock.Unlock()
}

// fillGroupLevel fill the group level
func (h *AnalysisExternalHandler) fillGroupLevel(proj *AnalysisProject) error {
	v, ok := h.bizDeptInfoCache.Load(proj.BizID)
	if ok {
		bizDeptInfo := v.(*ccv3.BusinessDeptInfo)
		fillProjectDeptInfo(proj, bizDeptInfo)
	} else {
		businessInfo, err := h.bkccClient.GetBizDeptInfo([]int64{proj.BizID})
		if err != nil {
			blog.Warnf("get project '%s' business ‘%d’ dept-info failed: %s", proj.ProjectName,
				proj.BizID, err.Error())
		} else {
			if bizDeptInfo, ok := businessInfo[proj.BizID]; ok {
				h.bizDeptInfoCache.Store(proj.BizID, bizDeptInfo)
				fillProjectDeptInfo(proj, bizDeptInfo)
			}
		}
	}
	for _, user := range proj.ActivityUsers {
		if common.IsAdminUser(user.UserName) {
			continue
		}
		v, ok = h.userDeptInfoCache.Load(user.UserName)
		if ok {
			fillUserDeptInfo(user, v.(*ccv3.UserDeptInfo))
			continue
		}
		if userDeptInfo, err := h.bkccClient.GetUserDeptInfo(user.UserName); err != nil {
			blog.Warnf("query user '%s' dept info failed: %s", user.UserName, err.Error())
		} else {
			h.userDeptInfoCache.Store(user.UserName, userDeptInfo)
			fillUserDeptInfo(user, userDeptInfo)
		}
	}
	return nil
}

var (
	externalRawDataPath = "/gitopsmanager/proxy/api/v1/analysis_new/raw_data"
)

// GetAnalysisProjects return analysis projects data
func (h *AnalysisExternalHandler) GetAnalysisProjects() []AnalysisProject {
	h.cacheLock.Lock()
	defer h.cacheLock.Unlock()

	result := make([]AnalysisProject, 0, len(h.cache))
	for i := range h.cache {
		item := h.cache[i]
		result = append(result, *(&item).DeepCopy())
	}
	return result
}

// getExternalRawData get the external raw data
func (h *AnalysisExternalHandler) getExternalRawData() ([]AnalysisProject, error) {
	req, err := http.NewRequest(http.MethodGet, h.op.ExternalAnalysisUrl+externalRawDataPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get external raw data failed")
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
	result := make([]AnalysisProject, 0)
	if err = json.Unmarshal(bs, &result); err != nil {
		return nil, errors.Wrapf(err, "unmarshal external data failed")
	}
	return result, nil
}

// GetResourceInfo fake function
func (h *AnalysisExternalHandler) GetResourceInfo() []AnalysisProjectResourceInfo {
	return nil
}
