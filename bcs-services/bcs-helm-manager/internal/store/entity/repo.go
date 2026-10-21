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
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// Repository 定义了仓库基础信息, 存储在helm-manager的数据库中, 是chart相关操作的基础
type Repository struct {
	// basic info
	TenantID          string `json:"tenantID" bson:"tenantID"`
	TenantProjectCode string `json:"tenantProjectCode" bson:"tenantProjectCode"`
	ProjectID         string `json:"projectID" bson:"projectID"`
	Name              string `json:"name" bson:"name"`
	DisplayName       string `json:"displayName" bson:"displayName"`
	Personal          bool   `json:"personal" bson:"personal"`
	Public            bool   `json:"public" bson:"public"`
	Type              string `json:"type" bson:"type"`
	RepoURL           string `json:"repoURL" bson:"repoURL"`

	// remote repo settings
	Remote         bool   `json:"remote" bson:"remote"`
	RemoteURL      string `json:"remoteURL" bson:"remoteURL"`
	RemoteUsername string `json:"remoteUsername" bson:"remoteUsername"`
	RemotePassword string `json:"remotePassword" bson:"remotePassword"`

	// auth
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`

	CreateBy   string `json:"createBy" bson:"createBy"`
	UpdateBy   string `json:"updateBy" bson:"updateBy"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
	UpdateTime int64  `json:"updateTime" bson:"updateTime"`
}

// GetRepoProjectID get repo project id
func (r *Repository) GetRepoProjectID() string {
	if r.Public {
		return repo.GetBKRepoProjectID("system", options.GlobalOptions.Repo.PublicRepoProject)
	}
	projectCode := r.ProjectID
	if r.TenantProjectCode != "" {
		projectCode = r.TenantProjectCode
	}
	return repo.GetBKRepoProjectID(r.TenantID, projectCode)
}

// GetRepoName get repo name
func (r *Repository) GetRepoName() string {
	if r.Public {
		return options.GlobalOptions.Repo.PublicRepoName
	}
	return r.Name
}

// Transfer2Proto transfer the data into protobuf struct
func (r *Repository) Transfer2Proto(ctx context.Context) *helmmanager.Repository {
	displayName := r.DisplayName
	if len(displayName) == 0 {
		displayName = common.ProjectRepoDefaultDisplayName
	}

	// 语言转换
	displayName = i18n.GetMsg(ctx, displayName)

	return &helmmanager.Repository{
		ProjectCode: common.GetStringP(r.ProjectID),
		Name:        common.GetStringP(r.Name),
		DisplayName: common.GetStringP(displayName),
		Public:      common.GetBoolP(r.Public),
		Type:        common.GetStringP(r.Type),
		RepoURL:     common.GetStringP(r.RepoURL),
		Remote:      common.GetBoolP(r.Remote),
		RemoteURL:   common.GetStringP(r.RemoteURL),
		Username:    common.GetStringP(r.Username),
		Password:    common.GetStringP(r.Password),
		CreateBy:    common.GetStringP(r.CreateBy),
		UpdateBy:    common.GetStringP(r.UpdateBy),
		CreateTime:  common.GetStringP(time.Unix(r.CreateTime, 0).UTC().Format(time.RFC3339)),
		UpdateTime:  common.GetStringP(time.Unix(r.UpdateTime, 0).UTC().Format(time.RFC3339)),
	}
}

// LoadFromProto load data from protobuf struct
func (r *Repository) LoadFromProto(repository *helmmanager.Repository) M {
	if repository == nil {
		return nil
	}
	m := make(M)

	if repository.ProjectCode != nil {
		r.ProjectID = repository.GetProjectCode()
		m[FieldKeyProjectID] = r.ProjectID
	}
	if repository.Name != nil {
		r.Name = repository.GetName()
		m[FieldKeyName] = r.Name
	}
	if repository.DisplayName != nil {
		r.DisplayName = repository.GetDisplayName()
		m[FieldKeyDisplayName] = r.Name
	}
	if repository.Public != nil {
		r.Public = repository.GetPublic()
		m[FieldKeyPublic] = r.Name
	}
	if repository.Type != nil {
		r.Type = repository.GetType()
		m[FieldKeyType] = r.Type
	}
	if repository.Remote != nil {
		r.Remote = repository.GetRemote()
		m[FieldKeyRemote] = r.Remote
	}
	if repository.RemoteURL != nil {
		r.RemoteURL = repository.GetRemoteURL()
		m[FieldKeyRemoteURL] = r.RemoteURL
	}
	if repository.RemoteUsername != nil {
		r.RemoteUsername = repository.GetRemoteUsername()
		m[FieldKeyRemoteUsername] = r.RemoteUsername
	}
	if repository.RemotePassword != nil {
		r.RemotePassword = repository.GetRemotePassword()
		m[FieldKeyRemotePassword] = r.RemotePassword
	}
	if repository.RepoURL != nil {
		r.RepoURL = repository.GetRepoURL()
		m[FieldKeyRepoURL] = r.RepoURL
	}
	if repository.Username != nil {
		r.Username = repository.GetUsername()
		m[FieldKeyUsername] = r.Username
	}
	if repository.Password != nil {
		r.Password = repository.GetPassword()
		m[FieldKeyPassword] = r.Password
	}
	if repository.CreateBy != nil {
		r.CreateBy = repository.GetCreateBy()
		m[FieldKeyCreateBy] = r.CreateBy
	}
	if repository.UpdateBy != nil {
		r.UpdateBy = repository.GetUpdateBy()
		m[FieldKeyUpdateBy] = r.UpdateBy
	}

	return m
}
