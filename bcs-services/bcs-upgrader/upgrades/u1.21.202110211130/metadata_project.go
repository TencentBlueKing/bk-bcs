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

package u1_21_202110211130

import (
	"fmt"
	"strconv"
	"time"
)

type respAllProject struct {
	ccBaseResp `json:",inline"`
	Data       respAllProjectData `json:"data"`
}

type respAllProjectData struct {
	Count   int         `json:"count"`
	Results []ccProject `json:"results"`
}

type ccProject struct {
	ApprovalStatus int       `json:"approval_status"`
	ApprovalTime   time.Time `json:"approval_time"`
	Approver       string    `json:"approver"`
	BgID           int       `json:"bg_id"`
	BgName         string    `json:"bg_name"`
	CcAppId        int       `json:"cc_app_id"`
	CenterID       int       `json:"center_id"`
	CenterName     string    `json:"center_name"`
	CreatedAt      time.Time `json:"created_at"`
	Creator        string    `json:"creator"`
	DataId         int       `json:"data_id"`
	DeployType     string    `json:"deploy_type"`
	DeptID         int       `json:"dept_id"`
	DeptName       string    `json:"dept_name"`
	Description    string    `json:"description"`
	EnglishName    string    `json:"english_name"`
	ID             int       `json:"id"`
	IsOfflined     bool      `json:"is_offlined"`
	IsSecrecy      bool      `json:"is_secrecy"`
	Kind           int       `json:"kind"`
	LogoAddr       string    `json:"logo_addr"`
	Name           string    `json:"name"`
	ProjectID      string    `json:"project_id"`
	ProjectName    string    `json:"project_name"`
	ProjectType    int       `json:"project_type"`
	Remark         string    `json:"remark"`
	UpdatedAt      time.Time `json:"updated_at"`
	Updator        string    `json:"updator"`
	UseBk          bool      `json:"use_bk"`
}

// get
type bcsProject struct {
	ProjectID   string      `json:"projectID"`   // required
	Name        string      `json:"name"`        // required
	EnglishName string      `json:"englishName"` // required
	Creator     string      `json:"creator"`     // required
	ProjectType int         `json:"projectType"` // required
	UseBKRes    bool        `json:"useBKRes"`    // required
	Description string      `json:"description"` // required
	IsOffline   bool        `json:"isOffline"`
	Kind        string      `json:"kind"`
	BusinessID  string      `json:"businessID"` // required
	DeployType  int         `json:"deployType"` // required
	BgID        string      `json:"bgID"`
	BgName      string      `json:"bgName"`
	DeptID      string      `json:"deptID"`
	DeptName    string      `json:"deptName"`
	CenterID    string      `json:"centerID"`
	CenterName  string      `json:"centerName"`
	IsSecret    bool        `json:"isSecret"`
	Updater     string      `json:"updater"` // update/get
	Credentials interface{} `json:"credentials"`
}

func data2BCSProject(ccProject ccProject) (*bcsProject, error) {

	var kind string
	businessID := strconv.Itoa(ccProject.CcAppId)
	bgID := strconv.Itoa(ccProject.BgID)
	deptID := strconv.Itoa(ccProject.DeptID)
	centerID := strconv.Itoa(ccProject.CenterID)
	switch ccProject.Kind {
	case 1:
		kind = "k8s"
		break
	case 2:
		kind = "mesos"
		break
	default:
		return nil, fmt.Errorf("")
	}
	deployType, err := strconv.Atoi(ccProject.DeployType)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	project := bcsProject{
		ProjectID:   ccProject.ProjectID,
		Name:        ccProject.Name,
		EnglishName: ccProject.EnglishName,
		Creator:     ccProject.Creator,
		ProjectType: 1, // TODO 此项待确认
		UseBKRes:    ccProject.UseBk,
		Description: ccProject.Description,
		IsOffline:   ccProject.IsOfflined,
		Kind:        kind,
		BusinessID:  businessID,
		DeployType:  deployType, // TODO 此项待定 deployType
		BgID:        bgID,
		BgName:      ccProject.BgName,
		DeptID:      deptID,
		DeptName:    ccProject.DeptName,
		CenterID:    centerID,
		CenterName:  ccProject.CenterName,
		IsSecret:    ccProject.IsSecrecy,
	}

	return &project, err
}

func diffProject(ccData ccProject, bcsData bcsProject) (isUpdate bool, project *bcsProject, err error) {

	var kind string
	switch ccData.Kind {
	case 1:
		kind = "k8s"
		break
	case 2:
		kind = "mesos"
		break
	default:
		return isUpdate, nil, fmt.Errorf("kind(%d) failed", ccData.Kind)
	}

	businessID := strconv.Itoa(ccData.CcAppId)
	bgID := strconv.Itoa(ccData.BgID)
	deptID := strconv.Itoa(ccData.DeptID)
	centerID := strconv.Itoa(ccData.CenterID)
	deployType, err := strconv.Atoi(ccData.DeployType)
	if err != nil {
		return isUpdate, nil, fmt.Errorf("deployType(%s) failed", ccData.DeployType)
	}

	project = &bcsProject{
		ProjectID:   ccData.ProjectID,
		Name:        ccData.Name,
		Updater:     bcsData.Updater,
		ProjectType: ccData.ProjectType,
		UseBKRes:    ccData.UseBk,
		Description: ccData.Description,
		IsOffline:   ccData.IsOfflined,
		Kind:        kind,
		DeployType:  deployType,
		BgID:        bgID,
		BgName:      ccData.BgName,
		DeptID:      deptID,
		DeptName:    ccData.DeptName,
		CenterID:    centerID,
		CenterName:  ccData.CenterName,
		IsSecret:    bcsData.IsSecret,
		BusinessID:  businessID,
		Credentials: bcsData.Credentials,
	}
	if *project == bcsData {
		return false, nil, nil
	}

	return true, project, nil
}
