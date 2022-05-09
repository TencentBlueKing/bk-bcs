/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bkrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
)

const (
	repositoryCreateUri = "/repository/api/repo/create"
	repositoryGetUri    = "/repository/api/repo/info"
)

func (rh *repositoryHandler) getRepository(ctx context.Context) (*repo.Repository, error) {

	blog.Infof("get repository from bk-repo projectID: %s, type: %s, name: %s",
		rh.projectID, rh.repoType, rh.repository)

	resp, err := rh.get(ctx, rh.getInfoRepositoryUri(), nil, nil)
	if err != nil {
		blog.Errorf("get repository from bk-repo failed, %s, projectID: %s, type: %s, name: %s",
			err.Error(), rh.projectID, rh.repoType, rh.repository)
		return nil, err
	}

	var r getRepositoryResp
	if err := codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("get repository from bk-repo decode resp failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("get repository from bk-repo get resp with error code %d, message: %s, traceID: %s",
			r.Code, r.Message, r.TraceID)
		// TODO: use code to identify
		if strings.Contains(r.Message, "not found") {
			return nil, errNotExist
		}
		return nil, fmt.Errorf("request error with code %d, %s", r.Code, r.Message)
	}

	if r.Data.Type != rh.repoType.String() {
		blog.Errorf("get repository from bk-repo get different repository type, query %s but get %s",
			rh.repoType.String(), r.Data.Type)
		return nil, fmt.Errorf("different repository type, query with %s but get repo with %s",
			rh.repoType.String(), r.Data.Type)
	}

	return &repo.Repository{
		ProjectID:   r.Data.ProjectID,
		Name:        r.Data.Name,
		Type:        repo.GetRepositoryType(r.Data.Type),
		Description: r.Data.Description,
	}, nil
}

func (rh *repositoryHandler) getInfoRepositoryUri() string {
	return repositoryGetUri + "/" + rh.projectID + "/" + rh.repository + "/" + rh.repoType.String()
}

func (rh *repositoryHandler) createRepository(ctx context.Context, rp *repo.Repository) (string, error) {
	blog.Infof("create repository to bk-repo with data %v", rp)

	var data []byte
	if err := codec.EncJson((&repository{}).load(rp), &data); err != nil {
		blog.Errorf("create repository to bk-repo encode json failed, %s, with data %v", err.Error(), rp)
		return "", err
	}

	blog.Infof("create repository to bk-repo with data: %s", string(data))
	resp, err := rh.post(ctx, repositoryCreateUri, nil, data)
	if err != nil {
		blog.Errorf("create repository to bk-repo post failed, %s, with data %v", err.Error(), rp)
		return "", err
	}

	var r createRepositoryResp
	if err := codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("create repository to bk-repo decode resp failed, %s, with resp %s", err.Error(), resp.Reply)
		return "", err
	}
	if r.Code != respCodeOK {
		blog.Errorf("create repository to bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)

		// TODO: use code to identify
		if strings.Contains(r.Message, "existed") {
			return "", errAlreadyExist
		}
		return "", fmt.Errorf("request error with code %d, %s", r.Code, r.Message)
	}

	blog.Infof("create repository to bk-repo successfully with data %v, traceID %s", rp, r.TraceID)
	return rh.getRepoURL(), nil
}

func (rh *repositoryHandler) getRepoURL() string {
	return rh.getUri("/helm/" + rh.projectID + "/" + rh.repository)
}

type repository struct {
	ProjectID     string                    `json:"projectId"`
	Name          string                    `json:"name"`
	Type          string                    `json:"type"`
	Category      string                    `json:"category"`
	Public        bool                      `json:"public"`
	Description   string                    `json:"description"`
	Configuration *repositoryRemoteSettings `json:"configuration"`

	// 未使用到的参数
	//StorageCredentialsKey string      `json:"storageCredentialsKey"`
	//Quota                 int64       `json:"quota"`
}

func (r *repository) load(rp *repo.Repository) *repository {
	if r == nil {
		return r
	}

	r.ProjectID = rp.ProjectID
	r.Name = rp.Name
	r.Category = "LOCAL"
	r.Public = false
	r.Description = rp.Description

	if rp.Remote {
		r.Category = "REMOTE"
		r.Configuration = &repositoryRemoteSettings{
			Type: "remote",
			URL:  rp.RemoteURL,
			Credentials: remoteCredentialsConfiguration{
				Username: rp.RemoteUsername,
				Password: rp.RemotePassword,
			},
		}
	}

	switch rp.Type {
	case repo.RepositoryTypeHelm:
		r.Type = "HELM"
	case repo.RepositoryTypeOCI:
		r.Type = "OCI"
	default:
		r.Type = "HELM"
	}

	return r
}

type repositoryRemoteSettings struct {
	Type        string                         `json:"type"`
	URL         string                         `json:"url"`
	Credentials remoteCredentialsConfiguration `json:"credentials"`
}

type remoteCredentialsConfiguration struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createRepositoryResp struct {
	basicResp
}

type getRepositoryResp struct {
	basicResp
	Data repository `json:"data"`
}
