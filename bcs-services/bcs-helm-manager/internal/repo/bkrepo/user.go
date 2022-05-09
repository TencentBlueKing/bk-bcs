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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
)

const (
	userCreateRepoUri = "/auth/api/user/create/repo"
)

func (rh *repositoryHandler) createUser(ctx context.Context) (string, string, error) {
	blog.Infof("create user to bk-repo for repository %s in project %s", rh.repository, rh.projectID)

	var data []byte
	u := rh.generateUserUsername()
	p := rh.generateUserPassword()
	if err := codec.EncJson(&user{
		Name:      u,
		Password:  p,
		UserID:    u,
		ProjectID: rh.projectID,
		RepoName:  rh.repository,
	}, &data); err != nil {
		blog.Errorf("create user to bk-repo for repository %s in project %s failed, %s",
			rh.repository, rh.projectID, err.Error())
		return "", "", err
	}

	resp, err := rh.post(ctx, userCreateRepoUri, nil, data)
	if err != nil {
		blog.Errorf("create user to bk-repo for repository %s in project %s failed, %s",
			rh.repository, rh.projectID, err.Error())
		return "", "", err
	}

	var r createUserResp
	if err := codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("create repository to bk-repo for repository %s in project %s "+
			"decode resp failed, %s, with resp %s",
			rh.repository, rh.projectID, err.Error(), resp.Reply)
		return "", "", err
	}
	if r.Code != respCodeOK {
		blog.Errorf("create user to bk-repo for repository %s in project %s "+
			"get resp with error code %d, message %s, traceID %s",
			rh.repository, rh.projectID, r.Code, r.Message, r.TraceID)

		return "", "", fmt.Errorf("request error with code %d, %s", r.Code, r.Message)
	}

	blog.Infof("create user to bk-repo successfully for repository %s in project %s, username: %s",
		rh.repository, rh.projectID, u)
	return u, p, nil
}

func (rh *repositoryHandler) generateUserUsername() string {
	return rh.generateUserCode("user-", 6)
}

func (rh *repositoryHandler) generateUserPassword() string {
	return rh.generateUserCode("", 12)
}

func (rh *repositoryHandler) generateUserCode(prefix string, length int) string {
	return prefix + common.RandomString(length)
}

type user struct {
	Name      string `json:"name"`
	Password  string `json:"pwd"`
	UserID    string `json:"userId"`
	ProjectID string `json:"projectId"`
	RepoName  string `json:"repoName"`
}

type createUserResp struct {
	basicResp
	Data bool `json:"data"`
}
