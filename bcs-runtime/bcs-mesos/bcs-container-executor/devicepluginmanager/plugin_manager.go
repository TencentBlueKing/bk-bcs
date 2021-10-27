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

package devicepluginmanager

import (
	"fmt"
	"net"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	comtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// DevicePluginManager manager for device plugins
type DevicePluginManager struct {
}

//new DevicePluginManager function
func NewDevicePluginManager() *DevicePluginManager {
	return &DevicePluginManager{}
}

//request device plugin to listandwatch device list ids
//return deviceIds, examples: ["cpuset0","cpuset1","cpuset2"...]
func (m *DevicePluginManager) ListAndWatch(ex *comtypes.ExtendedResource) ([]*pluginapi.Device, error) {
	//connect grpc socket
	conn, err := m.dial(ex.Socket, 5*time.Second)
	if err != nil {
		blog.Errorf("connect extended resource %s socket %s failed: %s", ex.Name, ex.Socket, err.Error())
		return nil, err
	}
	defer conn.Close()

	client := pluginapi.NewDevicePluginClient(conn)
	listAndWatchClient, err := client.ListAndWatch(context.Background(), &pluginapi.Empty{})
	if err != nil {
		blog.Errorf("extended resource %s ListAndWatch failed: %s", ex.Name, err.Error())
		return nil, err
	}
	response, err := listAndWatchClient.Recv()
	if err != nil {
		blog.Errorf("extended resource %s ListAndWatch receive message failed: %s", ex.Name, err.Error())
		return nil, err
	}
	blog.Infof("extended resource %s ListAndWatch success, devices(%s)", ex.Name, response.Devices)
	return response.Devices, nil
}

//request deviceplugin to allocate extended resources
//before create container call the function
func (m *DevicePluginManager) Allocate(ex *comtypes.ExtendedResource, deviceIds []string) (map[string]string, error) {
	//connect grpc socket
	conn, err := m.dial(ex.Socket, 5*time.Second)
	if err != nil {
		blog.Errorf("connect extended resource %s socket %s failed: %s", ex.Name, ex.Socket, err.Error())
		return nil, err
	}
	defer conn.Close()

	client := pluginapi.NewDevicePluginClient(conn)
	in := &pluginapi.AllocateRequest{
		ContainerRequests: make([]*pluginapi.ContainerAllocateRequest, 0),
	}
	req := &pluginapi.ContainerAllocateRequest{DevicesIDs: deviceIds}
	in.ContainerRequests = append(in.ContainerRequests, req)
	response, err := client.Allocate(context.Background(), in)
	if err != nil {
		blog.Errorf("extended resource %s Allocate devices(%v) failed: %s", ex.Name, deviceIds, err.Error())
		return nil, err
	}
	if len(response.ContainerResponses) == 0 {
		err = fmt.Errorf("ContainerResponses is empty")
		blog.Errorf("extended resource %s Allocate devices(%v) failed: %s", ex.Name, deviceIds, err.Error())
		return nil, err
	}

	//the envs are appended when the container is created
	//Some Settings of device plugin are done according to these docker envs
	blog.Infof("extended resource %s Allocate success, envs(%v)", ex.Name, response.ContainerResponses[0].Envs)
	return response.ContainerResponses[0].Envs, nil
}

//request deviceplugin to allocate extended resources
//before create container call the function
func (m *DevicePluginManager) PreStartContainer(ex *comtypes.ExtendedResource, deviceIds []string) error {
	//connect grpc socket
	conn, err := m.dial(ex.Socket, 5*time.Second)
	if err != nil {
		blog.Errorf("connect extended resource %s socket %s failed: %s", ex.Name, ex.Socket, err.Error())
		return err
	}
	defer conn.Close()

	client := pluginapi.NewDevicePluginClient(conn)
	in := &pluginapi.PreStartContainerRequest{DevicesIDs: deviceIds}
	_, err = client.PreStartContainer(context.Background(), in)
	if err != nil {
		blog.Errorf("extended resource %s Allocate devices(%v) failed: %s", ex.Name, deviceIds, err.Error())
		return err
	}

	blog.Infof("extended resource %s PreStartContainer success", ex.Name)
	return nil
}

// dial establishes the gRPC communication with device plugin.
func (m *DevicePluginManager) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
