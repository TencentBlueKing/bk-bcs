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
 *
 */

package web

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

// FeatureFlags map of feature flags
type FeatureFlags map[string]bool

// FeatureFlagsHandler 特新开关
func (s *WebServer) FeatureFlagsHandler(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("projectCode")
	clusterID := r.URL.Query().Get("clusterID")
	conf := config.G.FeatureFlags
	featureFlags := FeatureFlags{}
	for k, v := range conf {
		if v.Enabled {
			// if enabled, default is true, list as black list
			featureFlags[k] = true
			for _, pattern := range v.List {
				if matchResource(pattern, projectCode, clusterID) {
					featureFlags[k] = false
				}

			}
		} else {
			// if disabled, default is false, list as white list
			featureFlags[k] = false
			for _, pattern := range v.List {
				if matchResource(pattern, projectCode, clusterID) {
					featureFlags[k] = true
				}
			}
		}
	}
	okResponse := &OKResponse{Message: "OK", Data: featureFlags, RequestID: r.Header.Get("x-request-id")}
	render.JSON(w, r, okResponse)
}

func matchResource(pattern string, projectCode string, clusterID string) bool {
	patterns := strings.Split(pattern, ":")
	if len(patterns) == 0 {
		return false
	}
	if len(patterns) == 1 {
		return matchProject(patterns[0], projectCode)
	}
	projectPattern, clusterPattern := patterns[0], patterns[1]
	return matchProject(projectPattern, projectCode) && matchCluster(clusterPattern, clusterID)
}

func matchProject(pattern string, projectCode string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == projectCode {
		return true
	}
	return false
}

func matchCluster(pattern string, clusterID string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == clusterID {
		return true
	}
	return false
}
