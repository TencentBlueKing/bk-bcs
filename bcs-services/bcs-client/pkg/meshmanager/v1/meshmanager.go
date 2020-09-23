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

package v1

import (
	"context"
	"regexp"

	clientmeshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/meshmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type meshManager struct {
	clientOption      types.ClientOptions
	meshManagerClient meshmanager.MeshManagerClient
	conn              *grpc.ClientConn
}

//NewMeshManager create client for bcs-mesh-manager
func NewMeshManager(options types.ClientOptions) clientmeshmanager.MeshManager {
	m := &meshManager{
		clientOption: options,
	}
	return m
}

func (m *meshManager) dialGrpc() error {
	var err error
	//https://127.0.0.1:80 -> 127.0.0.1:80
	re := regexp.MustCompile("https?://")
	s := re.Split(m.clientOption.BcsApiAddress, 2)
	addr := s[len(s)-1]
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	md := metadata.New(header)
	m.conn, err = grpc.Dial(
		addr,
		grpc.WithDefaultCallOptions(grpc.Header(&md)),
		grpc.WithPerRPCCredentials(utils.NewTokenAuth(m.clientOption.BcsToken)),
		grpc.WithTransportCredentials(credentials.NewTLS(m.clientOption.ClientSSL)),
	)
	if err != nil {
		return err
	}
	m.meshManagerClient = meshmanager.NewMeshManagerClient(m.conn)
	return nil
}

func (m *meshManager) closeGrpc() {
	m.conn.Close()
	m.conn = nil
	m.meshManagerClient = nil
}

// CreateMeshCluster create meshcluster crd and install istio service
func (m *meshManager) CreateMeshCluster(req *meshmanager.CreateMeshClusterReq) (*meshmanager.CreateMeshClusterResp, error) {
	err := m.dialGrpc()
	if err != nil {
		return nil, err
	}
	defer m.closeGrpc()

	return m.meshManagerClient.CreateMeshCluster(context.TODO(), req)
}

// DeleteMeshCluster delete meshcluster crd and uninstall istio service
func (m *meshManager) DeleteMeshCluster(req *meshmanager.DeleteMeshClusterReq) (*meshmanager.DeleteMeshClusterResp, error) {
	err := m.dialGrpc()
	if err != nil {
		return nil, err
	}
	defer m.closeGrpc()

	return m.meshManagerClient.DeleteMeshCluster(context.TODO(), req)
}

//ListMeshCluster list meshcluster crds, contains istio components service status
func (m *meshManager) ListMeshCluster(req *meshmanager.ListMeshClusterReq) (*meshmanager.ListMeshClusterResp, error) {
	err := m.dialGrpc()
	if err != nil {
		return nil, err
	}
	defer m.closeGrpc()

	return m.meshManagerClient.ListMeshCluster(context.TODO(), req)
}
