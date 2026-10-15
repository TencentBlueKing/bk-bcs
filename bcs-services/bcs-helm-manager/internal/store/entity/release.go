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

package entity

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// Release 定义了chart的部署信息, 存储在helm-manager的数据库中, 用于对部署版本做记录
type Release struct {
	Name         string   `json:"name" bson:"name"`
	Namespace    string   `json:"namespace" bson:"namespace"`
	ProjectCode  string   `json:"projectCode" bson:"projectCode"`
	ClusterID    string   `json:"clusterID" bson:"clusterID"`
	Repo         string   `json:"repo" bson:"repo"`
	ChartName    string   `json:"chartName" bson:"chartName"`
	ChartVersion string   `json:"chartVersion" bson:"chartVersion"`
	Revision     int      `json:"revision" bson:"revision"`
	ValueFile    string   `json:"valueFile" bson:"valueFile"`
	Values       []string `json:"values" bson:"values"`
	Args         []string `json:"args" bson:"args"`
	CreateBy     string   `json:"createBy" bson:"createBy"`
	UpdateBy     string   `json:"updateBy" bson:"updateBy"`
	CreateTime   int64    `json:"createTime" bson:"createTime"`
	UpdateTime   int64    `json:"updateTime" bson:"updateTime"`
	Status       string   `json:"status" bson:"status"`
	Message      string   `json:"message" bson:"message"`
	Env          string   `json:"env" bson:"env"`
}

// Transfer2DetailProto transfer the data into detail protobuf struct
func (r *Release) Transfer2DetailProto() *helmmanager.ReleaseDetail {
	return &helmmanager.ReleaseDetail{
		Name:         common.GetStringP(r.Name),
		Namespace:    common.GetStringP(r.Namespace),
		Revision:     common.GetUint32P(uint32(r.Revision)),
		Chart:        common.GetStringP(r.ChartName),
		ChartVersion: common.GetStringP(r.ChartVersion),
		Values:       r.Values,
		Args:         r.Args,
		UpdateTime:   common.GetStringP(time.Unix(r.UpdateTime, 0).String()),
		CreateBy:     common.GetStringP(r.CreateBy),
		UpdateBy:     common.GetStringP(r.UpdateBy),
		Status:       common.GetStringP(r.Status),
		Message:      common.GetStringP(r.Message),
		Notes:        common.GetStringP(""),
		Description:  common.GetStringP(""),
		Repo:         common.GetStringP(r.Repo),
		ValueFile:    common.GetStringP(r.ValueFile),
	}
}

// Transfer2Proto transfer the data into release protobuf struct
func (r *Release) Transfer2Proto() *helmmanager.Release {
	return &helmmanager.Release{
		Name:           common.GetStringP(r.Name),
		Namespace:      common.GetStringP(r.Namespace),
		Revision:       common.GetUint32P(uint32(r.Revision)),
		Chart:          common.GetStringP(r.ChartName),
		ChartVersion:   common.GetStringP(r.ChartVersion),
		AppVersion:     common.GetStringP(r.ChartVersion),
		UpdateTime:     common.GetStringP(time.Unix(r.UpdateTime, 0).UTC().Format(time.RFC3339)),
		CreateBy:       common.GetStringP(r.CreateBy),
		UpdateBy:       common.GetStringP(r.UpdateBy),
		Status:         common.GetStringP(r.Status),
		Message:        common.GetStringP(r.Message),
		Repo:           common.GetStringP(r.Repo),
		IamNamespaceID: common.GetStringP(utils.CalcIAMNsID(r.ClusterID, r.Namespace)),
		ProjectCode:    common.GetStringP(r.ProjectCode),
		ClusterID:      common.GetStringP(r.ClusterID),
		Env:            common.GetStringP(r.Env),
	}
}
