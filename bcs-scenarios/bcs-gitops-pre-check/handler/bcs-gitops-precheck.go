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

// Package handler xxx
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/common"
	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

// PreCheckHandler handler struct
type PreCheckHandler struct {
	opts *Opts
}

// New handler
func New(opts *Opts) *PreCheckHandler {
	return &PreCheckHandler{opts: opts}
}

// Opts handler opts
type Opts struct {
	PublicFunc common.PublicFunc
}

var (
	// InternalErr internal err code
	InternalErr = uint32(500)
	// Success success code
	Success        = uint32(0)
	requestSuccess = "request success"
)

// GetMrInfo node cordon
func (p *PreCheckHandler) GetMrInfo(ctx context.Context,
	req *precheck.GetMrInfoReq, rsp *precheck.GetMrInfoRsp) error {
	blog.Infof("receive GetMRInfo request, repo:%s, mrIID:%s", req.Repository, req.MrIID)
	data, err := p.opts.PublicFunc.GetMrInfo(ctx, req.GetRepository(), req.GetMrIID())
	if err != nil {
		blog.Errorf("GetMrInfo request failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	msg := requestSuccess
	rsp.Message = msg
	rsp.Data = data
	return nil
}

// RecordTaskByPlugin record
func (p *PreCheckHandler) RecordTaskByPlugin(ctx context.Context, req *precheck.PreCheckTask,
	rsp *precheck.PreCheckTaskRsp) error {
	blog.Infof("receive RecordTaskByPlugin request, project:%s, repo:%s, user:%s", req.Project,
		req.RepositoryAddr, req.TriggerByUser)
	blog.Info("RecordTaskByPlugin req detail:", req)
	task, err := p.opts.PublicFunc.RecordPreCheckTask(ctx, req)
	if err != nil {
		blog.Errorf("record preCheckTask failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	msg := requestSuccess
	rsp.Message = msg
	rsp.Data = task
	return nil
}

// GetTaskByID get task
func (p *PreCheckHandler) GetTaskByID(ctx context.Context, req *precheck.GetTaskByIDReq,
	rsp *precheck.PreCheckTaskRsp) error {
	blog.Infof("receive GetTaskByID request, id:%s, project:%s, diffDetail", req.Id, req.Project, req.DiffDetail)
	if req.Id == "" || req.Project == "" {
		blog.Errorf("get preCheckTask failed, id %s illegal or project %s illegal", req.Id, req.Project)
		rsp.Code = &InternalErr
		msg := fmt.Sprintf("id %s illegal or project %s illegal", req.Id, req.Project)
		rsp.Message = msg
		return nil
	}
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		blog.Errorf("get preCheckTask failed, id %s illegal", req.Id)
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	task, err := p.opts.PublicFunc.GetPreCheckTask(ctx, id, req.Project)
	if err != nil {
		blog.Errorf("get task failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	if !req.DiffDetail {
		handleDiffDetail(task)
	}
	rsp.Code = &Success
	msg := requestSuccess
	rsp.Message = msg
	rsp.Data = task
	return nil
}

// UpdateTask update
func (p *PreCheckHandler) UpdateTask(ctx context.Context, req *precheck.PreCheckTask,
	rsp *precheck.PreCheckTaskRsp) error {
	blog.Infof("receive UpdateTask request, id:%s", req.Id)
	task, err := p.opts.PublicFunc.UpdatePreCheckTask(ctx, req)
	if err != nil {
		blog.Errorf("update preCheckTask failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	msg := requestSuccess
	rsp.Message = msg
	rsp.Data = task
	return nil
}

// ListTask list
func (p *PreCheckHandler) ListTask(ctx context.Context, req *precheck.ListTaskByIDReq,
	rsp *precheck.ListPreCheckTaskRsp) error {
	blog.Infof("receive ListTask request, project:%v, repos:%s", req.Projects, req.Repos)
	task, err := p.opts.PublicFunc.QueryPreCheckTaskList(ctx, req)
	if err != nil {
		blog.Errorf("list task failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	msg := requestSuccess
	rsp.Message = msg
	rsp.Data = task
	return nil
}

func handleDiffDetail(task *precheck.PreCheckTask) {
	if task.CheckDetail == nil || task.CheckDetail["diff"] == nil || task.CheckDetail["diff"].CheckDetail == nil {
		return
	}
	for app := range task.CheckDetail["diff"].CheckDetail {
		diffApp := task.CheckDetail["diff"].CheckDetail[app]
		if diffApp.Detail == nil || len(diffApp.Detail) == 0 {
			continue
		}
		newDetail := make([]*precheck.ResourceCheckDetail, 0)
		for index := range diffApp.Detail {
			if diffApp.Detail[index].Detail == "" {
				continue
			}
			newResource := &precheck.ResourceCheckDetail{}
			bytes, err := json.Marshal(diffApp.Detail[index])
			if err != nil {
				blog.Errorf("marshal detail to bytes error:%s", err.Error())
				continue
			}
			err = json.Unmarshal(bytes, newResource)
			if err != nil {
				blog.Errorf("unmarshal detail error:%s", err.Error())
				continue
			}
			newDetail = append(newDetail, newResource)
		}
		task.CheckDetail["diff"].CheckDetail[app].Detail = newDetail
	}
}
