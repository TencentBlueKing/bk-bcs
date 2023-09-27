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

package storage

// IPPoolDetailResponse response from storage
type IPPoolDetailResponse struct {
	ID         string    `json:"_id"`
	ClusterID  string    `json:"clusterId"`
	CreateTime string    `json:"createTime"`
	Datas      []*IPPool `json:"data"`
}

// IPPool information for cluster underlay ip resource
type IPPool struct {
	ClusterID string   `json:"cluster"`
	Net       string   `json:"net"`
	Mask      int      `json:"mask"`
	Gateway   string   `json:"gateway"`
	Created   string   `json:"created"`
	Hosts     []string `json:"hosts"`
	Reserved  []string `json:"reserved"`
	Available []string `json:"available"`
	Active    []string `json:"active"`
}
