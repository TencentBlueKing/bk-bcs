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
	"bytes"
	"encoding/gob"

	"github.com/mohae/deepcopy"
)

// AnalysisInterface defines the interface for analysis
type AnalysisInterface interface {
	Init() error
	GetAnalysisProjects() []AnalysisProject
	GetResourceInfo() []AnalysisProjectResourceInfo
}

// AnalysisProject defines the object for analysis
type AnalysisProject struct {
	BizID       int64  `json:"bizID"`
	BizName     string `json:"bizName"`
	ProjectID   string `json:"projectID"`
	ProjectCode string `json:"projectCode"`
	ProjectName string `json:"projectName"`

	GroupLevel0 string `json:"groupLevel0"`
	GroupLevel1 string `json:"groupLevel1"`
	GroupLevel2 string `json:"groupLevel2"`
	GroupLevel3 string `json:"groupLevel3"`
	GroupLevel4 string `json:"groupLevel4"`
	GroupLevel5 string `json:"groupLevel5"`

	Clusters        []*AnalysisCluster           `json:"clusters"`
	ApplicationSets []*AnalysisApplicationSet    `json:"applicationSets"`
	Applications    []*AnalysisApplication       `json:"applications"`
	Secrets         []*AnalysisSecret            `json:"secrets"`
	Repos           []*AnalysisRepo              `json:"repos"`
	ActivityUsers   []*AnalysisActivityUser      `json:"activityUsers"`
	Syncs           []*AnalysisSync              `json:"syncs"`
	ResourceInfo    *AnalysisProjectResourceInfo `json:"resourceInfo"`
}

// DeepCopy object
func (p *AnalysisProject) DeepCopy() *AnalysisProject {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(p); err != nil {
		return &AnalysisProject{}
	}
	dst := new(AnalysisProject)
	if err := dec.Decode(&dst); err != nil {
		return &AnalysisProject{}
	}
	return dst
}

// AnalysisProjectResourceInfo defines the resource-info of project
type AnalysisProjectResourceInfo struct {
	Name         string `json:"name"`
	ResourceAll  int64  `json:"resourceAll"`
	GameWorkload int64  `json:"gameWorkload"`
	Workload     int64  `json:"workload"`
	Pod          int64  `json:"pod"`
}

// DeepCopy object
func (ri *AnalysisProjectResourceInfo) DeepCopy() *AnalysisProjectResourceInfo {
	return deepcopy.Copy(ri).(*AnalysisProjectResourceInfo)
}

// AnalysisCluster defines the cluster info
type AnalysisCluster struct {
	ClusterName   string `json:"clusterName"`
	ClusterServer string `json:"clusterServer"`
	ClusterID     string `json:"clusterID"`
}

// AnalysisApplicationSet defines the applicationSet info
type AnalysisApplicationSet struct {
	Name string `json:"name"`
}

// AnalysisApplication defines the application info
type AnalysisApplication struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Cluster string `json:"cluster"`
}

// AnalysisSecret defines the secret info
type AnalysisSecret struct {
	Name string `json:"name"`
}

// AnalysisRepo defines the repo info
type AnalysisRepo struct {
	RepoName string `json:"repoName"`
	RepoUrl  string `json:"repoUrl"`
}

// AnalysisActivityUser defines the user info
type AnalysisActivityUser struct {
	UserName         string `json:"userName"`
	ChineseName      string `json:"chineseName"`
	Project          string `json:"project"`
	OperateNum       int64  `json:"operateNum"`
	LastActivityTime string `json:"lastActivityTimeTime"`

	GroupLevel0 string `json:"groupLevel0"`
	GroupLevel1 string `json:"groupLevel1"`
	GroupLevel2 string `json:"groupLevel2"`
	GroupLevel3 string `json:"groupLevel3"`
	GroupLevel4 string `json:"groupLevel4"`
	GroupLevel5 string `json:"groupLevel5"`
}

// AnalysisSync defines the sync info
type AnalysisSync struct {
	Application string `json:"application"`
	Cluster     string `json:"cluster"`
	SyncTotal   int64  `json:"syncTotal"`
	UpdateTime  string `json:"updateTime,omitempty"`
}
