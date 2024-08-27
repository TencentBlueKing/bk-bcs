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

import (
	"encoding/json"
	"time"

	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

// PreCheckTask task
type PreCheckTask struct {
	ID               int    `json:"ID" gorm:"column:id;primaryKey;type:int(11) AUTO_INCREMENT NOT NULL"`
	Project          string `json:"project" gorm:"column:project;primaryKey;type:varchar(255);index:idx_project"`
	RepositoryAddr   string `json:"repositoryAddr" gorm:"column:repository_addr;type:varchar(255);index:idx_repo"`
	MrIID            string `json:"mrIID" gorm:"column:mr_iid;type:varchar(255)"`
	CheckCallbackGit bool   `json:"checkCallbackGit" gorm:"column:check_callback_git;type:boolean"`
	CheckRevision    string `json:"checkRevision" gorm:"column:check_revision;type:varchar(255)"`
	// nolint
	ApplicationName      string    `json:"applicationName" gorm:"column:application_name;type:varchar(255);index:idx_application"`
	TriggerType          string    `json:"triggerType" gorm:"column:trigger_type;type:varchar(255)"`
	BranchValue          string    `json:"branchValue" gorm:"column:branchValue;type:varchar(255)"`
	CheckDetail          string    `json:"checkDetail" gorm:"column:check_detail;type:JSON"`
	CreateTime           time.Time `json:"createTime" gorm:"column:create_time;type:datetime"`
	UpdateTime           time.Time `json:"updateTime" gorm:"column:update_time;type:datetime"`
	TriggerByUser        string    `json:"triggerByUser" gorm:"column:trigger_by_user;type:varchar(255)"`
	CreateBy             string    `json:"createBy" gorm:"column:create_by;type:varchar(255)"`
	Finish               bool      `json:"finish" gorm:"column:finish;type:boolean"`
	Pass                 bool      `json:"pass" gorm:"column:pass;type:boolean"`
	FlowID               string    `json:"flowID" gorm:"column:flow_id;type:varchar(255)"`
	InvolvedApplications string    `json:"involvedApplications" gorm:"column:involved_applications;type:JSON"`
	NeedReplaceRepo      bool      `json:"needReplaceRepo" gorm:"column:needReplaceRepo;type:boolean"`
	ReplaceRepo          string    `json:"replaceRepo" gorm:"column:replaceRepo;type:varchar(255)"`
	ReplaceProject       string    `json:"replaceProject" gorm:"column:replaceProject;type:varchar(255)"`
	FlowLink             string    `json:"flowLink" gorm:"column:flowLink;type:varchar(255)"`
	MrInfo               string    `json:"mrInfo" gorm:"column:mrInfo;type:JSON"`
	Message              string    `json:"message" gorm:"column:message;type:varchar(255)"`
	ChooseApplication    bool      `json:"chooseApplication" gorm:"column:chooseApplication;type:boolean"`
	AppFilter            string    `json:"appFilter" gorm:"column:appFilter;type:varchar(255)"`
	LabelSelector        string    `json:"labelSelector" gorm:"column:labelSelector;type:varchar(255)"`
}

// InitTask init
func InitTask() *PreCheckTask {
	checkDetail := make(map[string]*precheck.ApplicationCheckDetail)
	checkDetailStr, _ := json.Marshal(checkDetail)
	involvedApplications := make([]string, 0)
	involvedApplicationsStr, _ := json.Marshal(involvedApplications)
	mrInfo := &precheck.MRInfoData{}
	mrInfoStr, _ := json.Marshal(mrInfo)
	return &PreCheckTask{
		CheckDetail:          string(checkDetailStr),
		CreateTime:           time.Now().UTC(),
		UpdateTime:           time.Now().UTC(),
		InvolvedApplications: string(involvedApplicationsStr),
		MrInfo:               string(mrInfoStr),
		Pass:                 true,
		Finish:               false,
	}
}

// ResourceCheckDetail resource detail
type ResourceCheckDetail struct {
	Finish       bool   `json:"finish" gorm:"column:finish;type:boolean"`
	Pass         bool   `json:"pass" gorm:"column:pass;type:boolean"`
	ResourceType string `json:"resourceType" gorm:"column:resource_type;type:varchar(255)"`
	ResourceName string `json:"resourceName" gorm:"column:resource_name;type:varchar(255)"`
	Detail       string `json:"detail" gorm:"column:detail;type:varchar(255)"`
}

// PreCheckTaskQuery query
type PreCheckTaskQuery struct {
	Projects     []string `json:"projects"`
	Repositories []string `json:"repositories"`
	StartTime    string   `json:"startTime"`
	EndTime      string   `json:"endTime"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
	WithDetail   bool     `json:"withDetail"`
}
