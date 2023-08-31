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

package config

// FeatureFlagOption feature flag option
type FeatureFlagOption struct {
	// 1. if set enabled to true, it means all projects enable this feature,
	// only projects in list(blacklist) disable this feature
	// 2. if set enabled to false, it means only projects in list(whitelist) enable this feature
	Enabled bool `yaml:"enabled"`
	// List can be white list or black list, depends on enabled
	List []string `yaml:"list"`
}
