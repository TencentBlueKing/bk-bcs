/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/thirdparty/repo"
)

// authRepo authorize the repo auth callback request, returns no permission error if is unauthorized.
func (s *Service) authRepo(kt *kit.Kit, opt *AuthRepoReq) error {
	kt.User = opt.UserId

	bizID, err := opt.ParseBizID()
	if err != nil {
		return err
	}

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Repo, Action: meta.Find}, BizID: bizID}
	authorized, err := s.bll.Auth().Authorize(kt, authRes)
	if err != nil {
		return err
	}

	if !authorized {
		return errf.New(errf.PermissionDenied, "no permission")
	}

	return nil
}

// AuthRepoReq auth repo request.
type AuthRepoReq struct {
	UserId    string          `json:"userId,omitempty"`
	Type      string          `json:"type,omitempty"`
	Action    string          `json:"action,omitempty"`
	ProjectId string          `json:"projectId,omitempty"`
	RepoName  string          `json:"repoName,omitempty"`
	Nodes     []*AuthRepoNode `json:"nodes,omitempty"`
}

// Bind
func (op *AuthRepoReq) Bind(r *http.Request) error {
	return nil
}

// AuthRepoNode auth repo node info.
type AuthRepoNode struct {
	FullPath string            `json:"fullPath,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

var repoProject string

// ParseBizID parse repo authorize request into biz id.
func (op AuthRepoReq) ParseBizID() (uint32, error) {
	if repoProject == "" {
		repoProject = cc.FeedServer().Repository.BkRepo.Project
	}

	if op.Type != "NODE" || op.Action != "READ" || op.ProjectId != repoProject {
		return 0, errf.New(errf.InvalidParameter, "request is invalid")
	}

	if len(op.Nodes) == 0 {
		return 0, errf.New(errf.InvalidParameter, "nodes are not set")
	}

	var bizIDStr string
	var bizID uint32
	for _, node := range op.Nodes {
		if len(node.Metadata) == 0 {
			return 0, errf.New(errf.InvalidParameter, "node has no metadata")
		}

		// validate that all nodes have the same biz id, and get the biz id in uint32 form
		metadataBizID, exists := node.Metadata["biz_id"]
		if !exists {
			return 0, errf.New(errf.InvalidParameter, "node has no biz id")
		}

		if bizIDStr == "" {
			bizIDStr = metadataBizID
			bizIDVal, err := strconv.ParseUint(bizIDStr, 10, 64)
			if err != nil {
				return 0, errf.New(errf.InvalidParameter, fmt.Sprintf("node biz id %s is invalid", bizIDStr))
			}
			bizID = uint32(bizIDVal)
		}

		if metadataBizID != bizIDStr {
			return 0, errf.New(errf.InvalidParameter, "nodes have multiple biz id")
		}

		// validate that all nodes have the same app id, and get the app id in uint32 form
		_, exists = node.Metadata["app_id"]
		if !exists {
			return 0, errf.New(errf.InvalidParameter, "node has no app id")
		}

		// validate that node path matches the bscp node path patten
		pathArr := strings.Split(node.FullPath, "/")

		nodePath, err := repo.GenNodeFullPath(pathArr[len(pathArr)-1])
		if err != nil {
			return 0, errf.New(errf.InvalidParameter, fmt.Sprintf("generate node path failed, err: %v", err))
		}

		if node.FullPath != nodePath {
			return 0, errf.New(errf.InvalidParameter, "node full path is invalid")
		}
	}

	// validate that repo name matches the biz id in nodes
	repoName, err := repo.GenRepoName(bizID)
	if err != nil {
		return 0, errf.New(errf.InvalidParameter, fmt.Sprintf("generate repository name failed, err: %v", err))
	}

	if op.RepoName != repoName {
		return 0, errf.New(errf.InvalidParameter, "repo name is invalid")
	}

	return bizID, nil
}
