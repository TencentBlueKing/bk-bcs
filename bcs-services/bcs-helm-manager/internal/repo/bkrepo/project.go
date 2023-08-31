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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
)

const (
	projectCheckURI  = "/repository/api/project/exist/"
	projectCreateURI = "/repository/api/project/create"
)

func (ph *projectHandler) ensureProject(ctx context.Context, prj *repo.Project) error {
	if prj == nil {
		return fmt.Errorf("project can not be empty")
	}

	exist, err := ph.checkProject(ctx, prj.Name)
	if err != nil {
		blog.Errorf("ensure project %s, check project failed, %s", prj.Name, err.Error())
		return err
	}
	if exist {
		blog.Infof("ensure project %s, exist confirmed", prj.Name)
		return nil
	}

	if err = ph.createProject(ctx, prj); err != nil && err != errAlreadyExist {
		blog.Errorf("ensure project failed, %s, name: %s", err.Error(), prj.Name)
		return err
	}

	blog.Infof("ensure project successfully, %s", prj.Name)
	return nil
}

func (ph *projectHandler) checkProject(ctx context.Context, name string) (bool, error) {
	blog.Infof("check project from bk-repo: %s", name)

	resp, err := ph.get(ctx, projectCheckURI+name, nil, nil)
	if err != nil {
		blog.Errorf("check project from bk-repo failed, %s, name: %s", err.Error(), name)
		return false, err
	}

	var r checkProjectResp
	if err := json.Unmarshal(resp.Reply, &r); err != nil {
		blog.Errorf("check project from bk-repo decode resp failed, %s, with resp %s", err.Error(), resp.Reply)
		return false, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("check project from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return false, fmt.Errorf("request error with code %d, %s", r.Code, r.Message)
	}

	return r.Data, nil
}

func (ph *projectHandler) createProject(ctx context.Context, prj *repo.Project) error {
	blog.Infof("create project to bk-repo with data %v", prj)

	var data []byte
	var err error
	if data, err = json.Marshal((&project{}).load(prj)); err != nil {
		blog.Errorf("create project to bk-repo encode json failed, %s, with data %v", err.Error(), prj)
		return err
	}

	resp, err := ph.post(ctx, projectCreateURI, nil, data)
	if err != nil {
		blog.Errorf("create project to bk-repo post failed, %s, with data %v", err.Error(), prj)
		return err
	}

	var r createProjectResp
	if err := json.Unmarshal(resp.Reply, &r); err != nil {
		blog.Errorf("create project to bk-repo decode resp failed, %s, with resp %s", err.Error(), resp.Reply)
		return err
	}
	if r.Code != respCodeOK {
		blog.Errorf("create project to bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)

		// check user is existed
		if strings.Contains(r.Message, "existed") {
			return errAlreadyExist
		}
		return fmt.Errorf("request error with code %d, %s", r.Code, r.Message)
	}

	blog.Infof("create project to bk-repo successfully with data %v, traceID %s", prj, r.TraceID)
	return nil
}

type project struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

func (p *project) load(prj *repo.Project) *project {
	if prj == nil {
		return p
	}

	p.Name = prj.Name
	p.DisplayName = prj.DisplayName
	p.Description = prj.Description
	return p
}

type createProjectResp struct {
	basicResp
}

type checkProjectResp struct {
	basicResp
	Data bool `json:"data"`
}
