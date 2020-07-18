/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"os"
	"os/signal"
	"syscall"

	k8score "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/internal/inspector"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/internal/options"
)

// Server server for cloud net agent
type Server struct {
	option *options.NetAgentOption

	inspector *inspector.NodeNetworkInspector

	cloudNetClient pbcloudnet.CloudNetserviceClient

	k8sClient k8score.CoreV1Interface
}

// New create server
func New(option *options.NetAgentOption) *Server {
	return &Server{
		option: option,
	}
}

func (s *Server) initInspector() error {
	s.inspector = inspector.New(s.option)
	if err := s.inspector.Init(); err != nil {
		return err
	}
	return nil
}

// Init init server
func (s *Server) Init() {
	if err := s.initInspector(); err != nil {
		blog.Fatalf("init Inspector failed, err %s", err.Error())
	}
}

// Run run server
func (s *Server) Run() {

	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
		syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		select {
		case <-interupt:
			blog.Infof("Get signal from system. Exit\n")
			return
		}
	}
}
