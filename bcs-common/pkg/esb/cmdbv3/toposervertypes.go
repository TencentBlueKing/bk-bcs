/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmdbv3

// TopoInst 实例拓扑结构
type TopoInst struct {
	InstID               int64  `json:"bk_inst_id"`
	InstName             string `json:"bk_inst_name"`
	ObjID                string `json:"bk_obj_id"`
	ObjName              string `json:"bk_obj_name"`
	Default              int    `json:"default"`
	HostCount            int64  `json:"host_count"`
	ServiceInstanceCount int64  `json:"service_instance_count,omitempty"`
	ServiceTemplateID    int64  `json:"service_template_id,omitempty"`
	SetTemplateID        int64  `json:"set_template_id,omitempty"`
	HostApplyEnabled     *bool  `json:"host_apply_enabled,omitempty"`
	HostApplyRuleCount   *int64 `json:"host_apply_rule_count,omitempty"`
}

// TopoInstRst 拓扑实例
type TopoInstRst struct {
	TopoInst `json:",inline"`
	Child    []*TopoInstRst `json:"child"`
}

// SearchBusinessTopoWithStatisticsResult result of SearchBusinessTopoWithStatistics
type SearchBusinessTopoWithStatisticsResult struct {
	BaseResp
	Data []*TopoInstRst `json:"data"`
}
