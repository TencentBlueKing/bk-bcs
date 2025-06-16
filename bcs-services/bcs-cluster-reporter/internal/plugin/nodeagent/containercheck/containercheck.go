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
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"

	"github.com/containerd/containerd/namespaces"
	containerd "github.com/containerd/containerd/v2/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

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
	metricmanager.Register(containerStatusMetric)
	metricmanager.Register(containerPorcessStatus)
	metricmanager.Register(runtimeStatus)
}

// Plugin xxx
type Plugin struct {
	opt    *Options
	ready  bool
	Detail Detail
	pluginmanager.NodePlugin
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
	if runMode == pluginmanager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					go p.Check(pluginmanager.CheckOption{})
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
	} else if runMode == pluginmanager.RunModeOnce {
		p.Check(pluginmanager.CheckOption{})
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

// Check check container status and state
func (p *Plugin) Check(option pluginmanager.CheckOption) {
	// 初始化变量
	result := make([]pluginmanager.CheckItem, 0, 0)
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())

	p.ready = false

	// 需要processcheck提供的kubelet参数
	if !strings.Contains(pluginmanager.Pm.GetPluginstr(), "processcheck") {
		config := pluginmanager.Pm.GetConfig()

		kubeletParams := make(map[string]string)
		processInfo, err := process.GetProcessInfo("kubelet", 0)
		if err != nil {
			klog.Errorf(err.Error())
			checkItem := pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: config.NodeConfig.NodeName,
				Detail:     fmt.Sprintf("check %s failed: %s", runtimeTarget, getProcessFailStatus),
				Normal:     false,
				Status:     getProcessFailStatus,
			}
			checkItem.Detail = fmt.Sprintf("containercheck error: %s", getProcessFailStatus)
			result = append(result, checkItem)
			return
		}

		for _, param := range processInfo.Params {
			if strings.HasPrefix(param, "--") && strings.Contains(param, "=") {
				param = strings.TrimPrefix(param, "--")
				key := strings.SplitN(param, "=", 2)[0]
				value := strings.SplitN(param, "=", 2)[1]
				kubeletParams[key] = value
			} else {
				param = strings.TrimPrefix(param, "--")
				kubeletParams[param] = ""
			}
			config.NodeConfig.KubeletParams = kubeletParams
			pluginmanager.Pm.SetConfig(config)
		}
	} else {
		pluginmanager.Pm.Ready("processcheck", "")
	}

	config := pluginmanager.Pm.GetConfig()
	node := config.NodeConfig
	nodeName := node.NodeName

	var runtimeErr error

	containerStatusGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
	containerPidStatusGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
	runtimeStatusGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)

	defer func() {
		p.CheckLock.Unlock()

		if runtimeErr != nil {
			checkItem := pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Detail:     fmt.Sprintf("check %s failed: %s", runtimeTarget, runtimeErr.Error()),
				Normal:     false,
				Status:     runtimeErrorStatus,
			}
			klog.Errorf("runtime error: %s", runtimeErr.Error())
			checkItem.Detail = fmt.Sprintf("runtime error: %s", runtimeErr.Error())
			result = append(result, checkItem)

			runtimeStatusGaugeVecSetList = append(runtimeStatusGaugeVecSetList, &metricmanager.GaugeVecSet{
				Labels: []string{nodeName, runtimeErrorStatus}, Value: float64(1),
			})
		}

		metricmanager.RefreshMetric(containerStatusMetric, containerStatusGaugeVecSetList)
		metricmanager.RefreshMetric(containerPorcessStatus, containerPidStatusGaugeVecSetList)
		metricmanager.RefreshMetric(runtimeStatus, runtimeStatusGaugeVecSetList)

		p.Result = pluginmanager.CheckResult{
			Items: result,
		}
		p.ready = true
		klog.Infof("end %s", p.Name())
	}()

	var sockList = sockPathes
	var socketPath string

	if p.opt.SockPath != "" {
		sockList = []string{p.opt.SockPath}
		klog.Infof("sockPath param is %s, remove default sockpathes", p.opt.SockPath)
	}

	// 获取可用的socket
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
		checkItemList, gvsList, err := dockerCheck(socketPath, node)
		if err != nil {
			runtimeErr = err
			return
		}
		result = append(result, checkItemList...)
		containerStatusGaugeVecSetList = append(containerStatusGaugeVecSetList, gvsList...)
	} else if strings.Contains(socketPath, "containerd.sock") {
		checkItemList, gvsList, err := containerdCheck(socketPath, node)
		if err != nil {
			runtimeErr = err
			return
		}
		result = append(result, checkItemList...)
		containerStatusGaugeVecSetList = append(containerStatusGaugeVecSetList, gvsList...)
	} else {
		runtimeErr = fmt.Errorf("unknown socket %s", socketPath)
		return
	}

	runtimeStatusGaugeVecSetList = append(runtimeStatusGaugeVecSetList, &metricmanager.GaugeVecSet{
		Labels: []string{nodeName, Normalstatus}, Value: float64(1),
	})
	result = append(result, pluginmanager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: nodeName,
		Level:      pluginmanager.WARNLevel,
		Normal:     true,
		Status:     Normalstatus,
	})
}

// dockerCheck 检查docker容器状态
func dockerCheck(socketPath string, node pluginmanager.NodeConfig) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0)
	nodeName := node.NodeName
	// 检查docker容器状态
	cli, err := GetDockerCli(socketPath)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		_ = cli.Close()
	}()

	containerList, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 确认检查相关的参数
	rootDir := "/var/lib/kubelet"
	if value, ok := node.KubeletParams["root-dir"]; ok {
		rootDir = value
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

			gvsList = append(gvsList, &metricmanager.GaugeVecSet{
				Labels: []string{container.ID, strings.Join(container.Names, "_"), nodeName, status}, Value: float64(1),
			})
			checkItemList = append(checkItemList, pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Normal:     false,
				Detail:     fmt.Sprintf("container %s state is %s", strings.Join(container.Names, "_"), status),
				Status:     inspectCoantainerError,
			})
			continue
		}

		// TODO 检测容器参数配置 需要开启特权模式
		//CheckContainerSystemParam(containerInfo.State.Pid)

		// 验证dns pod中的resolv内容正确
		// k8s_POD 排除pause镜像
		if strings.Contains(containerInfo.Name, "kube-system") && (strings.Contains(containerInfo.Name, "coredns") ||
			strings.Contains(containerInfo.Name, "kube-dns")) && !strings.Contains(containerInfo.Name, "k8s_POD") {
			checkItem, status, err := CheckDNSContainer(containerInfo.Name, containerInfo.ResolvConfPath, nodeName, node.HostPath)
			if err != nil {
				klog.Errorf("check container %s failed: %s", container.Names, err.Error())
				gvsList = append(gvsList, &metricmanager.GaugeVecSet{
					Labels: []string{container.ID, containerInfo.Name, nodeName, status}, Value: float64(1),
				})
				checkItemList = append(checkItemList, *checkItem)
			}
		}

		// 判断root-dir是否正确
		if _, ok := containerInfo.Config.Labels["io.kubernetes.container.name"]; ok && !strings.Contains(containerInfo.Name, "k8s_POD") {
			rootDirFlag := false
			for _, bindPath := range containerInfo.HostConfig.Binds {
				if strings.Contains(bindPath, rootDir) {
					rootDirFlag = true
					break
				}
			}

			if !rootDirFlag {
				gvsList = append(gvsList, &metricmanager.GaugeVecSet{
					Labels: []string{container.ID, containerInfo.Name, nodeName, wrongRootDirStatus}, Value: float64(1),
				})
				checkItemList = append(checkItemList, pluginmanager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Detail:     fmt.Sprintf("container %s root-dir is not %s", containerInfo.Name, rootDir),
					Status:     wrongRootDirStatus,
				})
			}
		}

	}

	return checkItemList, gvsList, nil
}

// containerdCheck 检查containerd容器状态
func containerdCheck(socketPath string, node pluginmanager.NodeConfig) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0)
	nodeName := node.NodeName

	// 连接到 containerd
	cli, err := containerd.New(socketPath)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = cli.Close()
	}()

	ctx := namespaces.WithNamespace(util.GetCtx(10*time.Second), "k8s.io")

	containerList, err := cli.Containers(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 确认检查相关的参数
	rootDir := "/var/lib/kubelet"
	klog.Info(node.KubeletParams)
	if value, ok := node.KubeletParams["root-dir"]; ok {
		rootDir = value
	}

	// check container status
	for _, container := range containerList {
		klog.Infof("start check for containerd container %s", container.ID())
		status, podName, err := ContainerdContainerCheck(container, ctx)
		if err != nil {
			klog.Errorf("check container %s failed: %s", podName, err.Error())
		}

		if status != Normalstatus {
			gvsList = append(gvsList, &metricmanager.GaugeVecSet{
				Labels: []string{container.ID(), podName, nodeName, status}, Value: float64(1),
			})
			checkItemList = append(checkItemList, pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Normal:     false,
				Status:     inconsistentStatus,
				Detail:     fmt.Sprintf("container of %s state is %s", podName, status),
			})
			continue
		}

		// 验证dns pod中的resolv内容正确
		// k8s_POD 排除pause镜像
		spec, err := container.Spec(ctx)
		if err != nil {
			klog.Errorf("check container %s failed: %s", podName, err.Error())
			continue
		}

		if strings.Contains(podName, "kube-system") && (strings.Contains(podName, "coredns") ||
			strings.Contains(podName, "kube-dns")) && !strings.Contains(podName, "k8s_POD") {
			resolvConfPath := ""
			for _, mount := range spec.Mounts {
				if mount.Destination == "/etc/resolv.conf" {
					resolvConfPath = mount.Source
				}
			}
			// 验证dns pod中的resolv内容正确
			checkItem, status, err := CheckDNSContainer(podName, resolvConfPath, nodeName, node.HostPath)
			if err != nil {
				klog.Errorf("check container %s failed: %s", podName, err.Error())
				gvsList = append(gvsList, &metricmanager.GaugeVecSet{
					Labels: []string{container.ID(), podName, nodeName, status}, Value: float64(1),
				})
				checkItemList = append(checkItemList, *checkItem)
			}

		}

		// 判断root-dir是否正确
		if _, ok := spec.Annotations["io.kubernetes.cri.container-name"]; ok {
			rootDirFlag := false
			for _, bindPath := range spec.Mounts {
				if strings.Contains(bindPath.Source, rootDir) {
					rootDirFlag = true
					break
				}
			}

			if !rootDirFlag {
				gvsList = append(gvsList, &metricmanager.GaugeVecSet{
					Labels: []string{container.ID(), podName, nodeName, wrongRootDirStatus}, Value: float64(1),
				})
				checkItemList = append(checkItemList, pluginmanager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Detail:     fmt.Sprintf("container %s root-dir is not %s", podName, rootDir),
					Status:     wrongRootDirStatus,
				})
			}
		}

	}

	return checkItemList, gvsList, nil
}

// DockerContainerCheck 检查容器状态一致性以及进程状态
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

// ContainerdContainerCheck 检查容器状态一致性以及进程状态
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
func CheckDNSContainer(name string, resolvConfPath string, nodeName string, hostPath string) (*pluginmanager.CheckItem, string, error) {
	checkItem := &pluginmanager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: nodeName,
		Normal:     true,
	}

	// 判断该容器是否为kube-system下的 dns 容器
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

	// 判断容器内的resolv文件中的nameserver配置和母机上的是否一致
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

// CheckContainerSystemParam 容器内核参数检查
func CheckContainerSystemParam(pid int) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	result := util.ExecCommand([]string{"nsenter", "-n", "-t", fmt.Sprintf("%d", pid), "sysctl", "-a"})

	reader := strings.NewReader(result)
	scanner := bufio.NewScanner(reader)

	systemParams := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		klog.Infof(line)
		if !strings.Contains(line, "=") {
			continue
		}
		line = strings.Replace(line, " ", "", -1)

		lineStrs := strings.Split(line, "=")
		if len(lineStrs) == 2 {
			systemParams[lineStrs[0]] = lineStrs[1]
		}

	}

	klog.Info(systemParams)

	return nil, nil, nil
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}

// GetResult xxx
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return p.Result
}

// Execute xxx
func (p *Plugin) Execute() {
	p.Check(pluginmanager.CheckOption{})
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}
