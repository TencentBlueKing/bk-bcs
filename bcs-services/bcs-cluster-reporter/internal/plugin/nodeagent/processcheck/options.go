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

// Package processcheck xxx
package processcheck

// Options bcs log options
type Options struct {
	Interval  int                  `json:"interval" yaml:"interval"`
	Processes []ProcessCheckConfig `json:"processes" yaml:"processes"`
}

type ProcessCheckConfig struct {
	Name       string `json:"name" yaml:"name"`
	ConfigFile string `json:"configFile" yaml:"configFile"`
}

// Validate validate options
func (o *Options) Validate() error {
	if o.Processes == nil {
		o.Processes = []ProcessCheckConfig{
			{Name: "kubelet"},
			{Name: "containerd", ConfigFile: "/etc/containerd/config.toml"},
			{Name: "dockerd", ConfigFile: "/etc/docker/daemon.json"},
		}
	}

	o.Processes = removeDuplicates(o.Processes)
	return nil
}

func removeDuplicates(pList []ProcessCheckConfig) []ProcessCheckConfig {
	result := []ProcessCheckConfig{}

	for _, p1 := range pList {
		flag := false
		for _, p2 := range result {
			if p1.Name == p2.Name && p1.ConfigFile == p2.ConfigFile {
				flag = true
				break
			}
		}

		if !flag {
			result = append(result, p1)
		}
	}

	return result
}
