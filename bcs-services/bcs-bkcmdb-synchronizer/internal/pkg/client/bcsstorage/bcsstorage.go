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

// Package bcsstorage define client for bcsstorage
package bcsstorage

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
)

type bcsStorageClient struct {
	Config *bcsapi.Config
}

// GetCMDBClient returns a CMDB client instance.
func (bsc *bcsStorageClient) GetCMDBClient() (client.CMDBClient, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerConnWithURL get a project manager grpc connection with url
func (bsc *bcsStorageClient) GetProjectManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerClient get a project manager grpc client
func (bsc *bcsStorageClient) GetProjectManagerClient() (pmp.BCSProjectClient, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerConn get a project manager grpc connection
func (bsc *bcsStorageClient) GetProjectManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// NewPMGrpcClientWithHeader create a project manager grpc client with header
func (bsc *bcsStorageClient) NewPMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ProjectManagerClientWithHeader {
	// implement me
	panic("implement me")
}

// GetDataManagerConnWithURL get a data manager grpc connection with url
func (bsc *bcsStorageClient) GetDataManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

func (bsc *bcsStorageClient) GetDataManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerConnWithURL get a cluster manager grpc connection with url
func (bsc *bcsStorageClient) GetClusterManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerClient get a cluster manager grpc client
func (bsc *bcsStorageClient) GetClusterManagerClient() (cmp.ClusterManagerClient, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerConn get a cluster manager grpc connection
func (bsc *bcsStorageClient) GetClusterManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// NewCMGrpcClientWithHeader create a cluster manager grpc client with header
func (bsc *bcsStorageClient) NewCMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ClusterManagerClientWithHeader {
	// implement me
	panic("implement me")
}

// GetStorageClient get a bcs storage client
func (bsc *bcsStorageClient) GetStorageClient() (bcsapi.Storage, error) {
	cli := bcsapi.NewClient(bsc.Config)
	return cli.Storage(), nil
}

// NewStorageClient create a bcs storage client
func NewStorageClient(config *bcsapi.Config) client.Client {
	return &bcsStorageClient{
		Config: config,
	}
}
