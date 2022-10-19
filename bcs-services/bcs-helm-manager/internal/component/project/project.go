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

// Package project xxx
package project

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	microRgt "github.com/micro/go-micro/v2/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/discovery"
)

const (
	// ProjectManagerServiceName project manager service name
	ProjectManagerServiceName = "project.bkbcs.tencent.com"
)

// ProjectClient xxx
type ProjectClient struct {
	Discovery       *discovery.ModuleDiscovery
	ClientTLSConfig *tls.Config
}

// Client xxx
var Client *ProjectClient

// NewClient create project service client
func NewClient(tlsConfig *tls.Config, microRgt microRgt.Registry) error {
	dis := discovery.NewModuleDiscovery(ProjectManagerServiceName, microRgt)
	err := dis.Start()
	if err != nil {
		return err
	}
	Client = &ProjectClient{Discovery: dis, ClientTLSConfig: tlsConfig}
	return nil
}

func (p *ProjectClient) getProjectClient() (bcsproject.BCSProjectClient, error) {
	node, err := p.Discovery.GetRandServiceInst()
	if err != nil {
		return nil, err
	}
	blog.V(4).Infof("get random project-manager instance [%s] from etcd registry successful", node.Address)

	cfg := bcsapi.Config{}
	// discovery hosts
	cfg.Hosts = []string{node.Address}
	cfg.TLSConfig = p.ClientTLSConfig
	cfg.InnerClientName = "bcs-helm-manager"
	return NewProjectManager(&cfg), nil
}

// NewProjectManager create ProjectManager SDK implementation
func NewProjectManager(config *bcsapi.Config) bcsproject.BCSProjectClient {
	rand.Seed(time.Now().UnixNano())
	if len(config.Hosts) == 0 {
		// ! pay more attention for nil return
		return nil
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}
	for k, v := range config.Header {
		header[k] = v
	}
	md := metadata.New(header)
	auth := &bcsapi.Authentication{InnerClientName: config.InnerClientName}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
		auth.Insecure = true
	}
	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts)
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			blog.Errorf("Create project manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		blog.Errorf("create no project manager client after all instance tries")
		return nil
	}
	// init project manager client
	return bcsproject.NewBCSProjectClient(conn)
}

var projectCache *sync.Map = &sync.Map{}

// GetProjectIDByCode get project id from project code
func GetProjectIDByCode(username string, projectCode string) (string, error) {
	// load project data from cache
	v, ok := projectCache.Load(projectCode)
	if ok {
		if project, ok := v.(*bcsproject.Project); ok {
			return project.ProjectID, nil
		}
	}
	client, err := Client.getProjectClient()
	p, err := client.GetProject(context.Background(), &bcsproject.GetProjectRequest{ProjectIDOrCode: projectCode})
	if err != nil {
		return "", fmt.Errorf("GetProject error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return "", fmt.Errorf("GetProject error, code: %d, data: %v", p.Code, p.GetData())
	}
	// save project data to cache
	projectCache.Store(projectCode, p.Data)
	return p.Data.ProjectID, nil
}
