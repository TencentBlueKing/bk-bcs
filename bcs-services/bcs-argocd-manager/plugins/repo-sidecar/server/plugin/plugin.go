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

package plugin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/instance"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/plugins/proto"

	"google.golang.org/grpc"
)

// Plugin describe the plugin service fetcher for current argocd instance.
type Plugin struct {
	// bcs-argocd-server address
	serverAddress string

	// argocd instance ID for this plugin service
	instanceID string

	conn *grpc.ClientConn
}

// New create a new Plugin for given argocd instance id,
// If serverAddress invalid or instanceID no exist, then return error
func New(serverAddress, instanceID string) (*Plugin, error) {
	conn, err := grpc.Dial(serverAddress)
	if err != nil {
		return nil, err
	}

	p := &Plugin{
		serverAddress: serverAddress,
		instanceID:    instanceID,
		conn:          conn,
	}
	if err := p.checkInstance(); err != nil {
		return nil, err
	}

	return p, nil
}

// FetchService get plugin information from server, then generate a Service for rendering
func (p *Plugin) FetchService(ctx context.Context, pluginName string) (*Service, error) {
	return nil, nil
}

func (p *Plugin) checkInstance() error {
	resp, err := instance.NewInstanceClient(p.conn).
		GetArgocdInstance(context.Background(), &instance.GetArgocdInstanceRequest{
			Name: &p.instanceID,
		})

	if err != nil {
		return err
	}

	// TODO: should check the instance status running
	if resp.GetInstance() == nil {
		return fmt.Errorf("instance %s not exist", p.instanceID)
	}

	return nil
}

// Service describe the plugin service for rendering
type Service struct {
	Protocol string
	Address  string
	Headers http.Header
}

// DoRender go request the plugin service and get the render result back
func (s *Service) DoRender(_ context.Context, env []string, data []byte) ([]byte, error) {
	paramData := string(data)

	rs := &proto.PluginRenderParam{
		Data: &paramData,
		Env:  env,
	}

	switch s.Protocol {
	case "HTTP", "http":
		return s.doHttp(rs)
	default:
		return nil, fmt.Errorf("unknown protocol for plugin %s", s.Protocol)
	}
}

func (s *Service) doHttp(rs *proto.PluginRenderParam) ([]byte, error) {
	var data []byte
	if err := codec.EncJson(rs, &data); err != nil {
		return nil, err
	}

	c := httpclient.NewHttpClient()
	r, err := c.Post(s.Address, s.Headers, data)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("target http code %d", r.StatusCode)
	}

	var result proto.PluginRenderResp
	if err = codec.DecJson(r.Reply, &result); err != nil {
		return nil, err
	}

	if result.GetCode() != 0 {
		return nil, fmt.Errorf("target result code %d, message: %s", result.Code, result.GetMessage())
	}

	return []byte(result.GetData()), nil
}