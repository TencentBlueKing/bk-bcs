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

package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/repo"
)

func (s *Service) createRepo(w http.ResponseWriter, r *http.Request) {
	rid := r.Header.Get(constant.RidKey)
	resp := new(BaseResp)

	req := new(repo.CreateRepoReq)
	if err := unmarshal(r.Body, req); err != nil {
		logs.Errorf("unmarshal body to request failed, err: %v, rid: %s", err, rid)
		resp.Err(w, err)
		return
	}

	repoPath := filepath.Clean(fmt.Sprintf("%s/%s/%s", s.Workspace, req.ProjectID, req.Name))
	if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
		logs.Errorf("mkdir repo directory failed, err: %v, rid: %s", err, rid)
		resp.Err(w, fmt.Errorf("mkdir repo directory failed, err: %v", err))
		return
	}

	resp.WriteResp(w, nil)
	return
}
