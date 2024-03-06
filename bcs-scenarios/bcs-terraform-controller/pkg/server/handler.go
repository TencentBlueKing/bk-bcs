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

// Package server xxx
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"

	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/core"
)

// Apply 执行tf apply命令
func (h *handler) Apply(ctx *gin.Context) {
	req := new(ApplyRequest)
	if err := ctx.BindJSON(req); err != nil {
		blog.Errorf("marshal request body failed(ApplyRequest), err: %s", err.Error())
		errReply(ctx, "marshal request body failed(ApplyRequest)")
		return
	}
	if len(req.Name) == 0 {
		errReply(ctx, "name is nil")
		return
	}
	req.Namespace = "default"
	tf := new(tfv1.Terraform)
	key := types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}
	cc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.client.Get(cc, key, tf); err != nil {
		text := fmt.Sprintf("get tf resource failed, req: %s, err: %s", req, err.Error())
		blog.Error(text)
		errReply(ctx, text)
		return
	}

	traceId := uuid.NewString()                      // tf-cr实例、tf-cr代码路径
	task := core.NewTask(nil, traceId, tf, h.client) // nolint do not pass a nil Context
	if err := task.Init(); err != nil {
		text := fmt.Sprintf("init task failed, req: %s, err: %s", req, err.Error())
		blog.Error(text)
		errReply(ctx, text)
		return
	}
	planLogs, err := task.Plan()
	if err != nil {
		text := fmt.Sprintf("tf cr plan failed, req: %s, err: %s, plan logs: %s", req, err.Error(), planLogs)
		blog.Error(text)
		errReply(ctx, text)
		return
	}
	applyLogs, err := task.Apply()
	if err != nil {
		text := fmt.Sprintf("tf cr apply failed, req: %s, err: %s, apply logs: %s", req, err.Error(), applyLogs)
		blog.Error(text)
		errReply(ctx, text)
		return
	}
	// note: 事件、状态
	// task.SyncEvent()
	// task.SyncStatus()

	detail := new(TaskApplyResult)
	detail.Result = true
	detail.Name = req.Name
	detail.Namespace = req.Namespace
	detail.Message = applyLogs

	ctx.JSON(http.StatusOK, &ApplyResponse{
		Code:    0,
		Result:  true,
		Message: "success",
		Data:    []*TaskApplyResult{detail},
	})
}

// CreatePlan 重新plan一次
func (h *handler) CreatePlan(ctx *gin.Context) {
	req := new(CreatePlanRequest)
	if err := ctx.BindJSON(req); err != nil {
		blog.Errorf("marshal request body failed(CreatePlanRequest), err: %s", err.Error())
		errReply(ctx, "marshal request body failed(CreatePlanRequest)")
		return
	}
	if len(req.ID) == 0 || len(req.Name) == 0 || len(req.Namespace) == 0 || len(req.TargetRevision) == 0 {
		errReply(ctx, "id or name or namespace or targetRevision is nil")
		return
	}

	cc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	key := types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}
	tf := new(tfv1.Terraform)
	if err := h.client.Get(cc, key, tf); err != nil {
		text := fmt.Sprintf("get tf resource failed, req: %v, err: %s", req, err.Error())
		blog.Error(text)
		errReply(ctx, text)
		return
	}

	// task := core.NewTask(req.ID, tf, h.client)
	// task.SetHook(req.Hook)
	// go task.Handler()
	// h.pool.Store(req.ID, task)

	reply(ctx, nil)
}

// ListTerraform 查询所有的tf资源
func (h *handler) ListTerraform(ctx *gin.Context) {
	cc, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	list := new(tfv1.TerraformList)
	if err := h.client.List(cc, list); err != nil {
		blog.Errorf("list terraform failed, err: %s", err)
		errReply(ctx, "list terraform failed")
		return
	}

	reply(ctx, list.Items)
}

// GetTerraform 按条件查询tf资源
func (h *handler) GetTerraform(ctx *gin.Context) {
	req := new(GetTerraformRequest)
	if err := ctx.BindJSON(req); err != nil {
		blog.Errorf("marshal request body failed(GetTerraformRequest), err: %s", err.Error())
		errReply(ctx, "marshal request body failed(GetTerraformRequest)")
		return
	}
	if len(req.Url) == 0 {
		errReply(ctx, "repo url is nil")
		return
	}

	cc, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	list := new(tfv1.TerraformList)
	if err := h.client.List(cc, list); err != nil {
		blog.Errorf("list terraform failed, err: %s", err)
		errReply(ctx, "list terraform failed")
		return
	}

	data := make([]tfv1.Terraform, 0)
	for i, tf := range list.Items { // 筛选
		if req.Url != tf.Spec.Repository.Repo {
			continue
		}
		data = append(data, list.Items[i])
	}

	reply(ctx, data)
}

// reply 成功响应
func reply(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, &BaseResponse{
		Data:    data,
		Code:    0,
		Result:  true,
		Message: "success",
	})
}

// errReply 错误响应
func errReply(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, &BaseResponse{
		Code:    -1,
		Result:  false,
		Message: message,
	})
}
