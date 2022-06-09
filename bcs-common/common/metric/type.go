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

package metric

import (
	"errors"
	"fmt"
)

// MetaData struct
type MetaData struct {
	Module     string            `json:"module"`
	IP         string            `json:"ip"`
	MetricPort uint              `json:"metricPort"`
	ClusterID  string            `json:"clusterID"`
	Labels     map[string]string `json:"label"`
}

// Valid xxx
func (m MetaData) Valid() error {
	var errs []error
	if len(m.Module) == 0 {
		errs = append(errs, errors.New("module is null"))
	}

	if len(m.IP) == 0 {
		errs = append(errs, errors.New("IPAddr is null"))
	}

	if m.MetricPort == 0 {
		errs = append(errs, errors.New("metric port is 0"))
	}

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// HealthInfo health info
type HealthInfo struct {
	RunMode    RunModeType `json:"runMode"`
	Module     string      `json:"module"`
	ClusterID  string      `json:"clusterID"`
	IP         string      `json:"ip"`
	HealthMeta `json:",inline"`
	AtTime     int64 `json:"atTime"`
}
