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

package kubedriver

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type KubeVersion struct {
	Major      string
	Minor      string
	GitVersion string
}

func (v KubeVersion) String() string {
	return fmt.Sprintf("git:%s/major:%s/minor:%s", v.GitVersion, v.Major, v.Minor)
}

func (v KubeVersion) IsValid() bool {
	// Minikube may not return minor and major
	return v.GitVersion != ""
}

type KubeAPIPrefer struct {
	Groups []struct {
		Name     string `json:"name"`
		Versions []struct {
			GroupVersion string `json:"groupVersion"`
			Version      string `json:"version"`
		} `json:"versions"`
		PreferredVersion struct {
			GroupVersion string `json:"groupVersion"`
			Version      string `json:"version"`
		} `json:"preferredVersion"`
	} `json:"groups"`
}

func (p KubeAPIPrefer) Map() map[string]string {
	kubePrefer := map[string]string{}
	for _, group := range p.Groups {
		kubePrefer[group.Name] = group.PreferredVersion.Version
	}
	return kubePrefer
}
