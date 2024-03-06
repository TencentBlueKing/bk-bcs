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

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	tfproto "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/proto"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/worker/podrunner"
)

const (
	defaultQueueSize = 1000
)

// TerraformQueue 定义 Terraform 处理队列，提供 Push 消息到对应 bucket 队列桶能力
type TerraformQueue interface {
	Push(ctx context.Context, terraform *tfv1.Terraform)
}

// NewTerraformServer create the instance of TerraformServer. It will create the queue channels
// which set in options. Function `Push` will send crs to every channel by mod by ASCII.
func NewTerraformServer() *TerraformServer {
	h := &TerraformServer{
		op:    option.GlobalOption(),
		queue: option.GlobalOption().WorkerQueue,
	}
	h.terraformChs = make([]chan *tfv1.Terraform, 0, h.queue)
	for i := 0; i < h.queue; i++ {
		h.terraformChs = append(h.terraformChs, make(chan *tfv1.Terraform, defaultQueueSize))
	}
	return h
}

// TerraformServer defines the grpc server used to handle cr message
type TerraformServer struct {
	tfproto.UnimplementedQueueServer

	sync.Mutex
	op           *option.ControllerOption
	queue        int
	runner       podrunner.Runner
	terraformChs []chan *tfv1.Terraform
	grpcServer   *grpc.Server
}

// Init will check the terraform-worker util it ready, terraform-worker is a StatefulSet
func (h *TerraformServer) Init(ctx context.Context) error {
	h.runner = podrunner.NewRunner()
	if err := h.runner.Init(ctx); err != nil {
		return errors.Wrapf(err, "init worker runner failed")
	}
	return nil
}

// Start the grpc server
func (h *TerraformServer) Start(ctx context.Context) error {
	ep := fmt.Sprintf("%s:%d", h.op.Address, h.op.GRPCPort)
	listener, err := net.Listen("tcp", ep)
	if err != nil {
		return errors.Wrapf(err, "listen grpc port '%s' failed", ep)
	}
	h.grpcServer = grpc.NewServer()
	tfproto.RegisterQueueServer(h.grpcServer, h)
	logctx.Infof(ctx, "Starting gRPC server on: %s", ep)
	if err = h.grpcServer.Serve(listener); err != nil {
		return errors.Wrapf(err, "serve grpc '%s' failed", ep)
	}
	return nil
}

// Stop the grpc server
func (h *TerraformServer) Stop() {
	h.grpcServer.Stop()
}

// Push the terraform message to channel by math mod
func (h *TerraformServer) Push(ctx context.Context, terraform *tfv1.Terraform) {
	h.Lock()
	defer h.Unlock()

	sum := 0
	for _, char := range terraform.Name {
		sum += int(char)
	}
	mod := sum % h.queue
	logctx.Infof(ctx, "terraform is sent to 'queue-%d', current length: %d", mod, len(h.terraformChs[mod]))
	h.terraformChs[mod] <- terraform
}

// Poll defines the grpc function, it will return the terraform message for corresponding worker
func (h *TerraformServer) Poll(ctx context.Context, req *tfproto.PollRequest) (*tfproto.PollResponse, error) {
	h.Lock()
	defer h.Unlock()

	if len(h.terraformChs[req.Index]) == 0 {
		return &tfproto.PollResponse{
			Data: nil,
		}, nil
	}
	tf, ok := <-h.terraformChs[req.Index]
	if !ok {
		return nil, errors.Errorf("terraform ch is closed")
	}
	bs, err := json.Marshal(tf)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal terraform failed")
	}
	return &tfproto.PollResponse{
		Data: bs,
	}, nil
}
