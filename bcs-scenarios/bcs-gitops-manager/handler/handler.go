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

package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/controller"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

// Options for Handler
type Options struct {
	Mode string
	// admin namespace for data control
	AdminNamespace string
	// Storage client for gitops data access
	Storage store.Store
	// cluster & project controller for data sync
	ClusterControl controller.ClusterControl
	ProjectControl controller.ProjectControl
}

// NewGitOpsHandler create handler
func NewGitOpsHandler(opt *Options) *BcsGitopsHandler {
	return &BcsGitopsHandler{
		option: opt,
	}
}

// BcsGitopsHandler for manager
type BcsGitopsHandler struct {
	option *Options
}

// Init BCSGitOpsHandler
func (e *BcsGitopsHandler) Init() error {
	// nothing todo
	return nil
}

// Ping implementation
func (e *BcsGitopsHandler) Ping(ctx context.Context,
	req *pb.GitOpsRequest, rsp *pb.GitOpsResponse) error {
	rsp.Code = 0
	rsp.Message = "OK"
	return nil
}
