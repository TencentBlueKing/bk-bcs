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

package cpuset_device

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	mesosdriver "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/types"

	docker "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	// KubeletSocketName socket file name for kubelet
	KubeletSocketName = "kubelet.sock"
	// CpusetSocketName socket file name for cpuset
	CpusetSocketName = "cpuset.sock"
	// CpusetResourceName resource name for cpuset
	CpusetResourceName = "bkbcs.tencent.com/cpuset"

	//EnvBkbcsAllocateCpuset docker env, examples: bkbcs_allocate_cpuset=node:0;cpuset:0,1,2,3
	EnvBkbcsAllocateCpuset = "bkbcs_allocate_cpuset"
)

// CpusetDevicePlugin device plugin for cpuset
type CpusetDevicePlugin struct {
	sync.RWMutex
	// config
	conf *config.Config

	// external resource cpuset name
	resourceName string
	// the grpc cpuset device socket
	// serve: ListAndWatch、Allocate function
	cpusetSocket string

	// docker client
	client *docker.Client
	// default unix:///var/run/docker.sock
	dockerSocket string

	server *grpc.Server
	// all cpuset device, reported device plugin manager
	devices []*pluginapi.Device
	// cpuset nodes, key = node_id, example 0 or 1
	nodes map[string]*types.CpusetNode
}

// NewCpusetDevicePlugin new cpuset device plugin
func NewCpusetDevicePlugin(conf *config.Config) *CpusetDevicePlugin {
	c := &CpusetDevicePlugin{
		resourceName: CpusetResourceName,
		cpusetSocket: fmt.Sprintf("%s/%s", conf.PluginSocketDir, CpusetSocketName),
		conf:         conf,
		dockerSocket: conf.DockerSocket,
		devices:      make([]*pluginapi.Device, 0),
		nodes:        make(map[string]*types.CpusetNode),
		server:       grpc.NewServer([]grpc.ServerOption{}...),
	}

	return c
}

// Start start cpu device plugin
func (c *CpusetDevicePlugin) Start() error {
	// init cpuset device
	err := c.initCpusetDevice()
	if err != nil {
		return err
	}

	// connect docker socket, and create docker client
	c.client, err = docker.NewClient(c.dockerSocket)
	if err != nil {
		blog.Errorf(err.Error())
		return err
	}

	// list running containers, update allocated cpusets in nodes
	err = c.updateCpusetNodes()
	if err != nil {
		return err
	}

	// watch docker create&stop event, handler container cpuset resources
	go c.listenerDockerEvent()
	// loop list containers to update cpuset node info
	go c.loopUpdateCpusetNodes()

	// start grpc server
	err = c.serve()
	if err != nil {
		return err
	}
	blog.Infof("grpc serve on %s success", c.cpusetSocket)

	// if device plugin in k8s cluster, then register device plugin info to kubelet
	if c.conf.Engine == "k8s" {
		err = c.register()
		if err != nil {
			blog.Errorf("register kubelet failed: %s", err.Error())
			return err
		}
		blog.Infof("Registered device plugin for '%s' with Kubelet", c.resourceName)
		// else device plugin in mesos cluster, then report extended resources info to mesos scheduler
	} else {
		err = c.reportExtendedResources()
		if err != nil {
			return err
		}
		blog.Infof("device plugin %s report mesos scheduler extended resources success", c.resourceName)
	}

	return nil
}

func (c *CpusetDevicePlugin) serve() error {
	os.Remove(c.cpusetSocket)
	sock, err := net.Listen("unix", c.cpusetSocket)
	if err != nil {
		blog.Errorf("Listen %s failed: %s", c.cpusetSocket, err.Error())
		return err
	}
	// register device plugin grpc server
	pluginapi.RegisterDevicePluginServer(c.server, c)
	go func() {
		blog.Infof("start grpc serve on %s", c.cpusetSocket)
		err := c.server.Serve(sock)
		if err != nil {
			blog.Errorf("grpc serve failed: %s", err.Error())
			os.Exit(1)
		}
	}()

	// Wait for server to start by launching a blocking connexion
	conn, err := c.dial(c.cpusetSocket, 5*time.Second)
	if err != nil {
		blog.Errorf("dial %s failed: %s", c.cpusetSocket, err.Error())
		return err
	}
	conn.Close()
	return nil
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (c *CpusetDevicePlugin) register() error {
	conn, err := c.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(c.cpusetSocket),
		ResourceName: c.resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

// dial establishes the gRPC communication with device plugin.
func (c *CpusetDevicePlugin) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
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

// numa info, command: numactl -H
// available: 2 nodes (0-1)
// node 0 cpus: 0 1 2 3 4 5 6 7 8 9 10 11 24 25 26 27 28 29 30 31 32 33 34 35
// node 0 size: 65414 MB
// node 0 free: 61563 MB
// node 1 cpus: 12 13 14 15 16 17 18 19 20 21 22 23 36 37 38 39 40 41 42 43 44 45 46 47
// node 1 size: 65536 MB
// node 1 free: 61398 MB
func (c *CpusetDevicePlugin) initCpusetDevice() error {
	nodes, err := NUMANodes()
	if err != nil {
		blog.Errorf(err.Error())
		return err
	}

	for _, node := range nodes {
		cpusets, err := NUMACPUsOfNode(node)
		if err != nil {
			blog.Errorf(err.Error())
			return err
		}
		blog.Infof("numa node %s cpus(%v)", node, cpusets)

		o := &types.CpusetNode{
			Id:              node,
			Cpuset:          make([]string, 0),
			AllocatedCpuset: make([]string, 0),
		}
		for _, cpuset := range cpusets {
			device := &pluginapi.Device{
				ID:     fmt.Sprintf("node:%s;cpuset:%s", node, cpuset),
				Health: "Healthy",
			}

			c.devices = append(c.devices, device)
			o.Cpuset = append(o.Cpuset, cpuset)
		}
		c.nodes[node] = o

	}
	blog.Infof("CpusetDevicePlugin init cpuset device success")
	return nil
}

func (c *CpusetDevicePlugin) reportExtendedResources() error {
	conf := &mesosdriver.MesosDriverClientConfig{
		ZkAddr:     c.conf.BcsZk,
		ClientCert: c.conf.ClientCert,
	}

	client, err := mesosdriver.NewMesosDriverClient(conf)
	if err != nil {
		blog.Errorf("NewMesosPlatform failed: %s", err.Error())
		return err
	}
	ex := &commtypes.ExtendedResource{
		InnerIP:  c.conf.NodeIP,
		Name:     c.resourceName,
		Capacity: float64(len(c.devices)),
		Socket:   c.cpusetSocket,
	}
	err = client.UpdateAgentExtendedResources(c.conf.ClusterID, ex)
	if err != nil {
		blog.Errorf("Update Agent ExtendedResources failed: %s", err.Error())
		return err
	}
	return nil
}

// GetDevicePluginOptions get device plugin options
func (c *CpusetDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// ListAndWatch lists devices and update that list according to the health status
func (c *CpusetDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	return s.Send(&pluginapi.ListAndWatchResponse{Devices: c.devices})
}

// Allocate which return list of devices.
func (c *CpusetDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	c.Lock()
	defer c.Unlock()

	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		blog.Infof("request allocate devices(%v)", req.DevicesIDs)
		var mnode *types.CpusetNode
		for _, node := range c.nodes {
			if node.Capacity() >= len(req.DevicesIDs) {
				mnode = node
				break
			}
		}
		// don't contain enough cpuset
		if mnode == nil {
			return nil, fmt.Errorf("no enough cpuset to allocated container")
		}
		cpuset, err := mnode.AllocateCpuset(len(req.DevicesIDs))
		if err != nil {
			blog.Errorf(err.Error())
			return nil, err
		}
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				EnvBkbcsAllocateCpuset: fmt.Sprintf("node:%s;cpuset:%s", mnode.Id, strings.Join(cpuset, ",")),
			},
		}

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

// PreStartContainer callback before starting container
func (c *CpusetDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (
	*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}
