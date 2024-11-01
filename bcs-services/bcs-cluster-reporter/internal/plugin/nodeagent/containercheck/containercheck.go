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

// Package containercheck xxx
package containercheck

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/containerd/containerd/namespaces"
	containerd "github.com/containerd/containerd/v2/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/types/process"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

var (
	containerStatusLabels        = []string{"id", "name", "node", "status"}
	containerPorcessStatusLabels = []string{"id", "name", "node", "status"}
	runtimeStatusLabels          = []string{"node", "status"}
	containerStatusMetric        = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "container_status",
		Help: "container_status",
	}, containerStatusLabels)
	containerPorcessStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "container_process_status",
		Help: "container_process_status",
	}, containerPorcessStatusLabels)
	runtimeStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "runtime_status",
		Help: "runtime_status",
	}, runtimeStatusLabels)

	sockPathes = []string{
		"/run/docker.sock",
		"/run/containerd/containerd.sock",
	}
)

func init() {
	metric_manager.Register(containerStatusMetric)
	metric_manager.Register(containerPorcessStatus)
	metric_manager.Register(runtimeStatus)
}

// Plugin xxx
type Plugin struct {
	opt    *Options
	ready  bool
	Detail Detail
	plugin_manager.NodePlugin
}

// Detail xxx
type Detail struct {
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}
	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return err
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	// run as daemon
	if runMode == plugin_manager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					go p.Check()
				} else {
					klog.Infof("the former %s didn't over, skip in this loop", p.Name())
				}
				select {
				case result := <-p.StopChan:
					klog.Infof("stop plugin %s by signal %d", p.Name(), result)
					return
				case <-time.After(time.Duration(interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == plugin_manager.RunModeOnce {
		p.Check()
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return pluginName
}

// Check xxx
func (p *Plugin) Check() {
	// 初始化变量
	result := make([]plugin_manager.CheckItem, 0, 0)
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())

	node := plugin_manager.Pm.GetConfig().NodeConfig
	nodeName := node.NodeName

	var runtimeErr error

	containerStatusGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	containerPidStatusGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	runtimeStatusGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)

	p.ready = false

	defer func() {
		p.CheckLock.Unlock()

		if runtimeErr != nil {
			checkItem := plugin_manager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Detail:     fmt.Sprintf("check %s failed: %s", runtimeTarget, runtimeErr.Error()),
				Normal:     false,
				Status:     runtimeErrorStatus,
			}
			klog.Errorf("runtime error: %s", runtimeErr.Error())
			checkItem.Detail = fmt.Sprintf("runtime error: %s", runtimeErr.Error())
			result = append(result, checkItem)

			runtimeStatusGaugeVecSetList = append(runtimeStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
				Labels: []string{nodeName, runtimeErrorStatus}, Value: float64(1),
			})
		}

		metric_manager.RefreshMetric(containerStatusMetric, containerStatusGaugeVecSetList)
		metric_manager.RefreshMetric(containerPorcessStatus, containerPidStatusGaugeVecSetList)
		metric_manager.RefreshMetric(runtimeStatus, runtimeStatusGaugeVecSetList)

		p.Result = plugin_manager.CheckResult{
			Items: result,
		}
		if !p.ready {
			p.ready = true
		}
		klog.Infof("end %s", p.Name())
	}()

	var sockList = sockPathes
	var socketPath string

	if p.opt.SockPath != "" {
		sockList = []string{p.opt.SockPath}
		klog.Infof("sockPath param is %s, remove default sockpathes", p.opt.SockPath)
	}

	for _, socketPath = range sockList {
		conn, err := net.Dial("unix", path.Join(node.HostPath, socketPath))
		if err != nil {
			socketPath = ""
			klog.Errorf(err.Error())
			continue
		} else {
			err = conn.Close()
			if err != nil {
				klog.Errorf("close socket failed: %s", err.Error())
			}
			break
		}
	}

	socketPath = path.Join(node.HostPath, socketPath)
	if strings.Contains(socketPath, "docker.sock") {
		cli, err := GetDockerCli(socketPath)
		if err != nil {
			runtimeErr = err
			return
		}

		defer func() {
			err = cli.Close()
			if err != nil {
				klog.Errorf("close docker cli failed: %s", err.Error())
			}
		}()

		containerList, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			runtimeErr = err
			return
		}

		// check container status
		for _, container := range containerList {
			klog.Infof("start check for docker container %s", container.Names)
			status, containerInfo, err := DockerContainerCheck(cli, container.ID, container.State)
			if err != nil {
				klog.Errorf("check container %s failed: %s", container.Names, err.Error())
			}

			if status != Normalstatus {
				klog.Errorf("container id: %s,inspect: %s, state: %s", container.ID, status, container.State)

				containerStatusGaugeVecSetList = append(containerStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
					Labels: []string{container.ID, strings.Join(container.Names, "_"), nodeName, status}, Value: float64(1),
				})
				result = append(result, plugin_manager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Detail:     fmt.Sprintf("container %s state is %s", strings.Join(container.Names, "_"), status),
					Status:     inspectCoantainerError,
				})
				continue
			}

			// 验证dns pod中的resolv内容正确
			checkItem, status, err := CheckDNSContainer(containerInfo.Name, containerInfo.ResolvConfPath, nodeName, node.HostPath)
			if err != nil {
				klog.Errorf("check container %s failed: %s", container.Names, err.Error())
				containerStatusGaugeVecSetList = append(containerStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
					Labels: []string{container.ID, strings.Join(container.Names, "_"), nodeName, status}, Value: float64(1),
				})
				result = append(result, *checkItem)
			}
		}
	} else if strings.Contains(socketPath, "containerd.sock") {
		// 连接到 containerd
		cli, err := containerd.New(socketPath)
		if err != nil {
			runtimeErr = err
			return
		}
		defer func() {
			err = cli.Close()
			if err != nil {
				klog.Errorf("close docker cli failed: %s", err.Error())
			}
		}()

		ctx := namespaces.WithNamespace(util.GetCtx(10*time.Second), "k8s.io")

		containerList, err := cli.Containers(ctx)
		if err != nil {
			runtimeErr = err
			return
		}

		// check container status
		for _, container := range containerList {
			klog.Infof("start check for containerd container %s", container.ID())
			status, podName, err := ContainerdContainerCheck(container, ctx)
			if err != nil {
				klog.Errorf("check container %s failed: %s", podName, err.Error())
			}

			if status != Normalstatus {
				containerStatusGaugeVecSetList = append(containerStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
					Labels: []string{container.ID(), podName, nodeName, status}, Value: float64(1),
				})
				result = append(result, plugin_manager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Status:     inconsistentStatus,
					Detail:     fmt.Sprintf("container of %s state is %s", podName, status),
				})
				continue
			}

			// 验证dns pod中的resolv内容正确
			spec, err := container.Spec(ctx)
			if err != nil {
				klog.Errorf("check container %s failed: %s", podName, err.Error())
				continue
			}
			resolvConfPath := ""
			for _, mount := range spec.Mounts {
				if mount.Destination == "/etc/resolv.conf" {
					resolvConfPath = mount.Source
				}
			}
			checkItem, status, err := CheckDNSContainer(podName, resolvConfPath, nodeName, node.HostPath)
			if err != nil {
				klog.Errorf("check container %s failed: %s", podName, err.Error())
				containerStatusGaugeVecSetList = append(containerStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
					Labels: []string{container.ID(), podName, nodeName, status}, Value: float64(1),
				})
				result = append(result, *checkItem)
			}
		}

	} else {
		runtimeErr = fmt.Errorf("unknown socket %s", socketPath)
		return
	}

	runtimeStatusGaugeVecSetList = append(runtimeStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
		Labels: []string{nodeName, Normalstatus}, Value: float64(1),
	})
	result = append(result, plugin_manager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: nodeName,
		Level:      plugin_manager.WARNLevel,
		Normal:     true,
		Status:     Normalstatus,
	})
}

// DockerContainerCheck xxx
func DockerContainerCheck(cli *client.Client, containerID string, state string) (string, types.ContainerJSON, error) {
	containerInfo, err := GetContainerInfo(cli, containerID)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return containerNotFoundStatus, containerInfo, err
		} else {
			return inspectCoantainerError, containerInfo, err
		}
	}

	if containerInfo.State.Status != state {
		return inconsistentStatus, containerInfo, nil
	}

	if containerInfo.State.Pid == 0 {
		return processNotExistStatus, containerInfo, nil
	}

	pidStatus, err := GetContainerPIDStatus(containerInfo.State.Pid)
	if err != nil {
		return getProcessFailStatus, containerInfo, err
	}

	if pidStatus == "D" || pidStatus == "Z" {
		return pidStatus, containerInfo, err
	}

	return Normalstatus, containerInfo, nil
}

// ContainerdContainerCheck xxx
func ContainerdContainerCheck(container containerd.Container, ctx context.Context) (string, string, error) {
	info, err := container.Info(ctx, containerd.WithoutRefreshedMetadata)
	if err != nil {
		return inspectCoantainerError, "", err
	}

	podName := ""
	// docker runtime的情况下，虽然containerd sock可以访问，但没有K8S的信息
	if name, ok := info.Labels["io.kubernetes.pod.name"]; ok {
		podName = name
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return Normalstatus, podName, nil
	}

	pid := task.Pid()
	pidStatus, err := GetContainerPIDStatus(int(pid))
	if err != nil {
		return inspectCoantainerError, "", err
	}

	if pidStatus == "D" || pidStatus == "Z" {
		return pidStatus, podName, err
	}
	return Normalstatus, podName, nil
}

// CheckDNSContainer 验证dns pod中的resolv内容正确
func CheckDNSContainer(name string, resolvConfPath string, nodeName string, hostPath string) (*plugin_manager.CheckItem, string, error) {
	checkItem := &plugin_manager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: nodeName,
		Normal:     true,
	}
	if strings.Contains(name, "kube-system") && (strings.Contains(name, "coredns") || strings.Contains(name, "kube-dns")) && !strings.Contains(name, "k8s_POD") {
		klog.Infof("check dns pod %s %s", name, resolvConfPath)

		containerPath := path.Join(hostPath, resolvConfPath)
		dnsResolv, err := os.ReadFile(containerPath)
		if err != nil {
			checkItem.Normal = false
			checkItem.Detail = fmt.Sprintf("dns container %s read %s failed: %s", name, containerPath, err.Error())
			checkItem.Status = readFileFailStatus
			return checkItem, readFileFailStatus, err
		}

		hostResolv, err := os.ReadFile(path.Join(hostPath, "/etc/resolv.conf"))
		if err != nil {
			checkItem.Detail = fmt.Sprintf("read %s failed: %s", hostPath, err.Error())
			checkItem.Status = readFileFailStatus
			if err != nil {
				return checkItem, readFileFailStatus, err
			}
		}

		dnsLines := make([]string, 0, 0)
		for _, dnsLine := range strings.Split(string(dnsResolv), "\n") {
			if !strings.HasPrefix(dnsLine, "nameserver") {
				continue
			}
			dnsLines = append(dnsLines, dnsLine)
		}

		hostLines := make([]string, 0, 0)
		for _, hostLine := range strings.Split(string(hostResolv), "\n") {
			if !strings.HasPrefix(hostLine, "nameserver") {
				continue
			}
			hostLines = append(hostLines, hostLine)
		}

		sort.Strings(dnsLines)
		sort.Strings(hostLines)

		equal := true
		if len(dnsLines) != len(hostLines) {
			equal = false
		} else {
			for i, item := range dnsLines {
				if hostLines[i] != item {
					equal = false
					break
				}
			}
		}

		if !equal {
			err = fmt.Errorf("content of dns %s is %s, different from %s ", containerPath, dnsLines, hostPath)
			checkItem.Normal = false
			checkItem.Detail = err.Error()
			checkItem.Status = Normalstatus
			return checkItem, dnsInconsistencyStatus, err
		}
	}

	return nil, Normalstatus, nil
}

// GetDockerCli xxx
func GetDockerCli(sockPath string) (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithHost(fmt.Sprintf("unix://%s", sockPath)))
	return cli, err

}

// GetContainerInfo xxx
func GetContainerInfo(cli *client.Client, containerID string) (types.ContainerJSON, error) {
	ctx := util.GetCtx(10 * time.Second)
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	return containerInfo, err
}

// GetContainerPIDStatus xxx
func GetContainerPIDStatus(pid int) (string, error) {
	processInfo, err := process.GetProcess(int32(pid))
	if err != nil {
		return "", err
	} else {
		return processInfo.Status()
	}
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}

// GetResult xxx
func (p *Plugin) GetResult(string) plugin_manager.CheckResult {
	return p.Result
}

// Execute xxx
func (p *Plugin) Execute() {
	p.Check()
}

// GetDetail
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}
