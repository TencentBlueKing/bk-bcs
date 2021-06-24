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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	mesosdriver "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/types"

	"github.com/fsnotify/fsnotify"
	docker "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	// KubeletSocketName socket file name for kubelet
	KubeletSocketName = "kubelet.sock"
	// CpusetSocketName socket file name for cpuset
	CpusetSocketName = "bcs-cpuset-device.sock"
	// CpusetResourceName resource name for cpuset
	CpusetResourceName = "bkbcs.tencent.com/cpuset"

	// EnvBkbcsAllocateCpuset docker env, examples: bkbcs_allocate_cpuset=node:0;cpuset:0,1,2,3
	EnvBkbcsAllocateCpuset = "bkbcs_allocate_cpuset"

	// EnvBcsCpusetCheckIntervalMinute env name for check interval
	EnvBcsCpusetCheckIntervalMinute = "BKBCS_CPUSET_CHECK_INTERVAL_MINUTE"
	// EnvBcsCpusetCleanDelayMinute env name for cpuset expire duration
	EnvBcsCpusetCleanDelayMinute = "BKBCS_CPUSET_CLEAN_DELAY_MINUTE"
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
	// the grpc socket of kubelet
	kubeletSocket string

	// docker client
	client *docker.Client
	// default unix:///var/run/docker.sock
	dockerSocket string

	server *grpc.Server
	// all cpuset device, reported device plugin manager
	devices []*pluginapi.Device
	// cpuset nodes, key = node_id, example 0 or 1
	nodes map[string]*types.CpusetNode

	// lock for write cgroup files
	cgroupFileLock sync.Mutex

	// interval for check running container
	checkInterval time.Duration

	// stop channel
	stopCh chan struct{}
}

// NewCpusetDevicePlugin new cpuset device plugin
func NewCpusetDevicePlugin(conf *config.Config) *CpusetDevicePlugin {
	c := &CpusetDevicePlugin{
		resourceName:  CpusetResourceName,
		cpusetSocket:  fmt.Sprintf("%s/%s", conf.PluginSocketDir, CpusetSocketName),
		kubeletSocket: fmt.Sprintf("%s/%s", conf.PluginSocketDir, KubeletSocketName),
		conf:          conf,
		dockerSocket:  conf.DockerSocket,
		devices:       make([]*pluginapi.Device, 0),
		nodes:         make(map[string]*types.CpusetNode),
		stopCh:        make(chan struct{}),
	}

	return c
}

// Start start cpu device plugin
func (c *CpusetDevicePlugin) Start() error {
	if err := c.loadEnv(); err != nil {
		return err
	}
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

	// if device plugin in k8s cluster, then register device plugin info to kubelet
	if c.conf.Engine == "k8s" {
		go c.registerLoop()
	} else {
		err = c.startServer()
		if err != nil {
			return err
		}
		// else device plugin in mesos cluster, then report extended resources info to mesos scheduler
		err = c.reportExtendedResources()
		if err != nil {
			return err
		}
		blog.Infof("device plugin %s report mesos scheduler extended resources success", c.resourceName)
	}

	return nil
}

func (c *CpusetDevicePlugin) startServer() error {
	c.server = grpc.NewServer([]grpc.ServerOption{}...)
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
	blog.Infof("grpc serve on %s success", c.cpusetSocket)
	return nil
}

func (c *CpusetDevicePlugin) stopServer() {
	if c.server != nil {
		c.server.Stop()
		c.server = nil
		blog.Infof("stop grpc serve")
	}
}

// registerLoop watch kubelet sock path and do register
func (c *CpusetDevicePlugin) registerLoop() {
	c.stopServer()
	if err := c.startServer(); err != nil {
		blog.Warnf("start grpc server failed, err %s", err.Error())
		time.Sleep(5 * time.Second)
		go c.registerLoop()
		return
	}
	blog.Infof("begin to register to kubelet")
	if err := c.register(); err != nil {
		blog.Warnf("register to kubelet failed, err %s, will try again,", err.Error())
		time.Sleep(5 * time.Second)
		go c.registerLoop()
		return
	}
	blog.Infof("begin watch kubelet socket path %s", c.kubeletSocket)
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		blog.Warnf("create file watcher failed, err %s", err.Error())
		time.Sleep(5 * time.Second)
		go c.registerLoop()
		return
	}
	defer fileWatcher.Close()
	err = fileWatcher.Add(c.kubeletSocket)
	if err != nil {
		blog.Warnf("watch file %s failed, err %s", c.kubeletSocket, err.Error())
		time.Sleep(5 * time.Second)
		go c.registerLoop()
		return
	}
	for {
		select {
		case we := <-fileWatcher.Events:
			// kubelet socket path event and event type is create
			if we.Name == c.kubeletSocket && (we.Op&fsnotify.Remove) == fsnotify.Remove {
				blog.Infof("file watcher event: kubelet socket file %s removed, try to restart device-plugin",
					c.kubeletSocket)
				time.Sleep(2 * time.Second)
				go c.registerLoop()
				return
			}
		case err := <-fileWatcher.Errors:
			blog.Warnf("file watcher errors %s", err.Error())
			go c.registerLoop()
			return
		case <-c.stopCh:
			blog.Infof("kubelet socket watcher exited")
			return
		}
	}
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (c *CpusetDevicePlugin) register() error {
	conn, err := c.dial(c.kubeletSocket, 5*time.Second)
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

func (c *CpusetDevicePlugin) loadEnv() error {
	checkIntervalStr := os.Getenv(EnvBcsCpusetCheckIntervalMinute)
	if len(checkIntervalStr) == 0 {
		c.checkInterval = 10 * time.Minute
	} else {
		checkInterval, err := strconv.Atoi(checkIntervalStr)
		if err != nil {
			return fmt.Errorf("env %s parse %s to int failed, err %s",
				EnvBcsCpusetCheckIntervalMinute, checkIntervalStr, err.Error())
		}
		c.checkInterval = time.Duration(checkInterval) * time.Minute
	}
	return nil
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

		nodeNumber, err := strconv.Atoi(node)
		if err != nil {
			return err
		}

		o := &types.CpusetNode{
			Id:                  node,
			Cpuset:              make([]string, 0),
			AllocatedCpuset:     make([]string, 0),
			AllocatedCpusetTime: make(map[string]time.Time),
		}
		for _, cpuset := range cpusets {
			// filter reserved cpuset
			if _, isReserved := c.conf.ReservedCPUSet[cpuset]; isReserved {
				blog.Infof("reserved cpu %s of node %s", cpuset, node)
				continue
			}
			device := &pluginapi.Device{
				ID:     fmt.Sprintf("%s", cpuset),
				Health: "Healthy",
				Topology: &pluginapi.TopologyInfo{
					Nodes: []*pluginapi.NUMANode{
						{
							ID: int64(nodeNumber),
						},
					},
				},
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
func (c *CpusetDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (
	*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// ListAndWatch lists devices and update that list according to the health status
func (c *CpusetDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	ticker := time.NewTicker(360 * time.Second)
	defer ticker.Stop()
	s.Send(&pluginapi.ListAndWatchResponse{Devices: c.devices})
	for {
		select {
		case <-c.stopCh:
			blog.Infof("list watch stop")
			return nil
		case <-ticker.C:
			s.Send(&pluginapi.ListAndWatchResponse{Devices: c.devices})
		}
	}
}

// Allocate which return list of devices.
func (c *CpusetDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (
	*pluginapi.AllocateResponse, error) {
	c.Lock()
	defer c.Unlock()

	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		blog.Infof("request allocate devices(%v)", req.DevicesIDs)
		if len(req.DevicesIDs) == 0 {
			blog.Warnf("request allocate devices is empty")
			return nil, fmt.Errorf("request allocate devices is empty")
		}
		nodeIDs := make(map[string]struct{})
		for _, id := range req.DevicesIDs {
			for _, node := range c.nodes {
				for _, cpuset := range node.Cpuset {
					if cpuset == id {
						nodeIDs[node.Id] = struct{}{}
						break
					}
				}
			}
		}
		if len(nodeIDs) > 1 {
			blog.Warnf("request devices %v belongs different numa node", req.DevicesIDs)
			return nil, fmt.Errorf("request devices %v belongs different numa node", req.DevicesIDs)
		}
		var nodeID string
		for key := range nodeIDs {
			nodeID = key
			break
		}

		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				EnvBkbcsAllocateCpuset: fmt.Sprintf("node:%s;cpuset:%s", nodeID, strings.Join(req.DevicesIDs, ",")),
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
