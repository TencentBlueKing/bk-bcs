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

package deviceplugin

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// EniDevicePlugin is the plugin which implements k8s device plugin interface
type EniDevicePlugin struct {
	// the name of resource managed by this plugin
	resourceName string

	// k8s device plugin manager sock
	kubeletSockPath string

	// the path of unix sock file for this plugin
	socketPath string
	// the grpc server for this plugin
	server *grpc.Server

	limit int

	deviceLock sync.Mutex
	devices    []*pluginapi.Device

	deviceUpdateCh chan struct{}
	stopCh         chan struct{}
}

// NewEniDevicePlugin create EniDevicePlugin object
func NewEniDevicePlugin(kubeletSocketPath, pluginSocketPath, resourceName string) *EniDevicePlugin {
	return &EniDevicePlugin{
		resourceName:    resourceName,
		kubeletSockPath: kubeletSocketPath,
		socketPath:      pluginSocketPath,
	}
}

func (p *EniDevicePlugin) init() {
	p.server = grpc.NewServer([]grpc.ServerOption{}...)
	p.stopCh = make(chan struct{})
	p.deviceUpdateCh = make(chan struct{}, 2)
}

func (p *EniDevicePlugin) cleanup() {
	close(p.stopCh)
	close(p.deviceUpdateCh)
	p.server = nil
	p.stopCh = nil
	p.deviceUpdateCh = nil
}

func (p *EniDevicePlugin) serve() error {
	os.Remove(p.socketPath)
	sock, err := net.Listen("unix", p.socketPath)
	if err != nil {
		blog.Errorf("listen socket file path %s failed, err %s", p.socketPath, err.Error())
		return err
	}

	pluginapi.RegisterDevicePluginServer(p.server, p)
	go func() {
		blog.Infof("starting device plugin grpc server")
		if err := p.server.Serve(sock); err != nil {
			blog.Errorf("device plugin grpc server serve failed, err %s", err.Error())
		}
	}()

	// connect to with 5 second timeout, to wait for server ready
	conn, err := p.grpcDial(p.socketPath, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

func (p *EniDevicePlugin) register() error {
	conn, err := p.grpcDial(p.kubeletSockPath, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(p.socketPath),
		ResourceName: p.resourceName,
		Options: &pluginapi.DevicePluginOptions{
			PreStartRequired: true,
		},
	}

	_, err = client.Register(context.Background(), req)
	if err != nil {
		blog.Errorf("do plugin register failed, err %s", err.Error())
		return err
	}
	blog.Infof("register plugin to kubelet successfully")
	return nil
}

func (p *EniDevicePlugin) grpcDial(sock string, timeout time.Duration) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(sock,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		blog.Errorf("connect %s failed, err %s", sock, err.Error())
		return nil, err
	}
	return conn, nil
}

// Start start plugin server
func (p *EniDevicePlugin) Start() error {
	p.init()
	if err := p.serve(); err != nil {
		blog.Errorf("device plugin server failed to serve, err %s", err.Error())
		p.cleanup()
		return err
	}

	if err := p.register(); err != nil {
		blog.Errorf("register device plugin failed, err %s", err.Error())
		p.Stop()
		return err
	}
	blog.Infof("successfully register device plugin")
	return nil
}

// Stop stop plugin server
func (p *EniDevicePlugin) Stop() error {
	if p.server == nil {
		blog.Infof("grpc server is nil")
		return nil
	}
	blog.Infof("stop server")
	p.server.Stop()
	if err := os.Remove(p.socketPath); err != nil && !os.IsNotExist(err) {
		blog.Infof("delete unix socket file %s failed, err %s", p.socketPath, err.Error())
		return err
	}
	p.cleanup()
	return nil
}

// GetDevicePluginOptions return options to k8s device plugin manager
func (p *EniDevicePlugin) GetDevicePluginOptions(
	ctx context.Context, r *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// PreStartContainer is called before container start
func (p *EniDevicePlugin) PreStartContainer(
	ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *EniDevicePlugin) sendToListWatchServer(s pluginapi.DevicePlugin_ListAndWatchServer) {
	p.deviceLock.Lock()
	defer p.deviceLock.Unlock()
	for _, d := range p.devices {
		blog.Infof("send devices %s to kubelet", d.String())
	}
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: p.devices}); err != nil {
		blog.Warnf("send list watch response to kubelet failed, err %s", err.Error())
	}
	blog.Infof("update devices to kubelet successfully")
}

// ListAndWatch lists devices to k8s device plugin manager
func (p *EniDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	p.sendToListWatchServer(s)
	for {
		select {
		case <-p.stopCh:
			blog.Infof("listwatch stop")
			return nil
		case <-p.deviceUpdateCh:
			p.sendToListWatchServer(s)
		}
	}
}

// Allocate be called when container is creating
func (p *EniDevicePlugin) Allocate(
	ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	response := pluginapi.AllocateResponse{
		ContainerResponses: []*pluginapi.ContainerAllocateResponse{},
	}

	blog.Infof("Request Containers: %v", r.GetContainerRequests())
	for range r.GetContainerRequests() {
		response.ContainerResponses = append(response.ContainerResponses,
			&pluginapi.ContainerAllocateResponse{},
		)
	}

	return &response, nil
}

// SetDeviceLimit set limit devices
func (p *EniDevicePlugin) SetDeviceLimit(limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negtive")
	}
	if limit == p.limit {
		blog.Infof("limit stays %d", limit)
		return nil
	}
	devices := make([]*pluginapi.Device, 0)
	for i := 0; i < limit; i++ {
		devices = append(devices, &pluginapi.Device{
			ID:     strconv.Itoa(i),
			Health: pluginapi.Healthy,
		})
	}
	p.deviceLock.Lock()
	p.devices = devices
	p.deviceLock.Unlock()
	p.deviceUpdateCh <- struct{}{}
	blog.Infof("devices changes, trigger request to kubelet")
	return nil
}
