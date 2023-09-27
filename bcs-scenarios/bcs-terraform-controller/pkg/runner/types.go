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
 *
 */

package runner

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
)

const (
	// InitAction init
	InitAction = "init"
	// PlanAction plan
	PlanAction = "plan"
	// ApplyAction apply
	ApplyAction = "apply"
	// DestroyAction destroy
	DestroyAction = "destroy"
)

// InitRequest terraform init request
type InitRequest struct {
	TfInstance string `protobuf:"bytes,1,opt,name=tfInstance,proto3" json:"tfInstance,omitempty"`
	Upgrade    bool   `protobuf:"varint,2,opt,name=upgrade,proto3" json:"upgrade,omitempty"`
	ForceCopy  bool   `protobuf:"varint,3,opt,name=forceCopy,proto3" json:"forceCopy,omitempty"`
}

// InitReply terraform init reply
type InitReply struct {
	Message             string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	StateLockIdentifier string `protobuf:"bytes,2,opt,name=stateLockIdentifier,proto3" json:"stateLockIdentifier,omitempty"`
}

// NewTerraformRequest new terraform request
type NewTerraformRequest struct {
	WorkingDir string `protobuf:"bytes,1,opt,name=workingDir,proto3" json:"workingDir,omitempty"`
	ExecPath   string `protobuf:"bytes,2,opt,name=execPath,proto3" json:"execPath,omitempty"`
	// Terraform  []byte `protobuf:"bytes,3,opt,name=terraform,proto3" json:"terraform,omitempty"`
	Terraform  *tfv1.Terraform `protobuf:"bytes,3,opt,name=terraform,proto3" json:"terraform,omitempty"`
	InstanceID string          `protobuf:"bytes,4,opt,name=instanceID,proto3" json:"instanceID,omitempty"`
}

// NewTerraformReply new terraform reply
type NewTerraformReply struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

// PlanRequest terraform plan request
type PlanRequest struct {
	TfInstance string   `protobuf:"bytes,1,opt,name=tfInstance,proto3" json:"tfInstance,omitempty"`
	Out        string   `protobuf:"bytes,2,opt,name=out,proto3" json:"out,omitempty"`
	Refresh    bool     `protobuf:"varint,3,opt,name=refresh,proto3" json:"refresh,omitempty"`
	Destroy    bool     `protobuf:"varint,4,opt,name=destroy,proto3" json:"destroy,omitempty"`
	Targets    []string `protobuf:"bytes,5,rep,name=targets,proto3" json:"targets,omitempty"`
}

// PlanReply terraform plan reply
type PlanReply struct {
	Drifted             bool   `protobuf:"varint,1,opt,name=drifted,proto3" json:"drifted,omitempty"`
	Message             string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	StateLockIdentifier string `protobuf:"bytes,3,opt,name=stateLockIdentifier,proto3" json:"stateLockIdentifier,omitempty"`
	PlanCreated         bool   `protobuf:"varint,4,opt,name=planCreated,proto3" json:"planCreated,omitempty"`
}

// ApplyRequest apply terraform request
type ApplyRequest struct {
	TfInstance         string   `protobuf:"bytes,1,opt,name=tfInstance,proto3" json:"tfInstance,omitempty"`
	DirOrPlan          string   `protobuf:"bytes,2,opt,name=dirOrPlan,proto3" json:"dirOrPlan,omitempty"`
	RefreshBeforeApply bool     `protobuf:"varint,3,opt,name=refreshBeforeApply,proto3" json:"refreshBeforeApply,omitempty"`
	Targets            []string `protobuf:"bytes,4,rep,name=targets,proto3" json:"targets,omitempty"`
	Parallelism        int32    `protobuf:"varint,5,opt,name=parallelism,proto3" json:"parallelism,omitempty"`
}

// ApplyReply apply terraform reply
type ApplyReply struct {
	Message             string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	StateLockIdentifier string `protobuf:"bytes,2,opt,name=stateLockIdentifier,proto3" json:"stateLockIdentifier,omitempty"`
}

// DestroyRequest destroy terraform request
type DestroyRequest struct {
	TfInstance string   `protobuf:"bytes,1,opt,name=tfInstance,proto3" json:"tfInstance,omitempty"`
	Targets    []string `protobuf:"bytes,2,rep,name=targets,proto3" json:"targets,omitempty"`
}

// DestroyReply destroy terraform reply
type DestroyReply struct {
	Message             string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	StateLockIdentifier string `protobuf:"bytes,2,opt,name=stateLockIdentifier,proto3" json:"stateLockIdentifier,omitempty"`
}

// localPrintfer 本地化的输出对象
// 由于底层tf命令执行的原因，这里处理上要包一层
type localPrintfer struct{}

// Printf 对应 tfexec.printfer 接口
func (l *localPrintfer) Printf(format string, v ...interface{}) {
	blog.Info(fmt.Sprintf(format, v...))
}
