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

// Package process xxx
package process

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/moby/sys/mountinfo"
	"github.com/shirou/gopsutil/process"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// HOST_PROC

var (
	mnt          NS = "mnt"
	hostPath        = util.GetHostPath()
	systemdPaths    = []string{
		"/etc/systemd/system/",
		"/usr/lib/systemd/system/",
	}
)

// GetProcessNS xxx
func GetProcessNS(pid int32, ns NS) (syscall.Stat_t, error) {
	mntNSFile := fmt.Sprintf("/proc/%d/ns/%s", pid, ns)

	var stat syscall.Stat_t
	err := syscall.Stat(mntNSFile, &stat)
	return stat, err
}

// GetProcessServiceConfigfiles xxx
func GetProcessServiceConfigfiles(starter string) (map[string]string, error) {
	serviceFiles := make(map[string]string)
	for _, systemdPath := range systemdPaths {
		var err error
		serviceFiles[starter], err = GetConfigfile(path.Join(hostPath, systemdPath, starter))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		fileInfo, err := os.Stat(path.Join(hostPath, systemdPath, fmt.Sprintf("%s.d", starter)))
		if err == nil {
			if fileInfo.IsDir() {
				files, err := os.ReadDir(path.Join(hostPath, systemdPath, fmt.Sprintf("%s.d", starter)))
				if err != nil {
					return nil, err
				}

				for _, serviceFile := range files {
					if !serviceFile.IsDir() {
						serviceFiles[serviceFile.Name()], err = GetConfigfile(path.Join(hostPath, systemdPath, fmt.Sprintf("%s.d", starter), serviceFile.Name()))
						if err != nil {
							return nil, err
						}
					}
				}
			}
		}
		break
	}

	return serviceFiles, nil
}

// GetStarter other, systemd, container, crontab, cmdline
func GetStarter(pid, ppid int32) (string, error) {
	starter := "other"
	switch ppid {
	case 1:
		cgroupFile := fmt.Sprintf("%s/proc/%d/cgroup", hostPath, pid)
		file, err := os.Open(cgroupFile)
		if err != nil {
			klog.Errorf("Get process cgroup info failed: %s", err.Error())
			return starter, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "/system.slice/") {
				starter = filepath.Base(strings.Split(line, ":")[2])
				break
			}
		}

	default:
		parent, err := process.NewProcess(ppid)
		if err != nil {
			klog.Errorf("Get process parent info failed: %s", err.Error())
			return starter, err
		}

		parentName, err := parent.Name()
		if err != nil {
			klog.Errorf("Get process parentName failed: %s", err.Error())
			return starter, err
		}

		if parentName == "cron" {
			starter = "crontab"
		} else if strings.Contains(parentName, "runc") || strings.Contains(parentName, "containerd-shim") {
			starter = "container"
		} else if strings.Contains(parentName, "-bash") || strings.Contains(parentName, "-sh") {
			starter = "cmdline"
		}

	}

	return starter, nil
}

// GetConfigfile xxx
func GetConfigfile(path string) (string, error) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}

// GetMountInfoSourcePath xxx
func GetMountInfoSourcePath(source string) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/proc/1/mountinfo", util.GetHostPath()))
	if err != nil {
		return "", fmt.Errorf("open path %s failed: %s", fmt.Sprintf("%s/proc/1/mountinfo", util.GetHostPath()), err.Error())
	}

	mountInfoList, err := mountinfo.GetMountsFromReader(f, nil)
	if err != nil {
		return "", fmt.Errorf("Get process mountinfo failed: %s", err.Error())
	}

	for _, mountInfo := range mountInfoList {
		if mountInfo.Source == source {
			return mountInfo.Mountpoint, nil
		}
	}
	return "", fmt.Errorf("Get process mountinfo failed: notfound")

}

// GetConfigfileList xxx
func GetConfigfileList(params []string, pid int32) (map[string]string, error) {
	configFiles := make(map[string]string)

	// 获取该进程mount信息
	f, err := os.Open(fmt.Sprintf("%s/proc/%d/mountinfo", util.GetHostPath(), pid))
	if err != nil {
		klog.Errorf("Get process %d ConfigFiles failed: %s", pid, err.Error())
	}

	mountInfoList, err := mountinfo.GetMountsFromReader(f, nil)
	if err != nil {
		klog.Errorf("Get process %d ConfigFiles failed: %s", pid, err.Error())
	}
	for _, param := range params {
		if strings.Contains(param, "log") {
			continue
		}

		re := regexp.MustCompile(`(?:^|=)(/[a-zA-Z0-9\._-]+)+`)
		//re := regexp.MustCompile(`(/[^/]+)+`)
		paths := re.FindAllString(param, -1)

		for _, configFilepath := range paths {
			configFilepath = strings.TrimPrefix(configFilepath, "=")
			if strings.HasSuffix(configFilepath, "sock") {
				continue
			}

			for _, mountInfo := range mountInfoList {
				// 还需要判断 /install-cni.sh 此类情况  // 获取到work目录，并找到对应的merged目录
				if mountInfo.Mountpoint == "/" || mountInfo.Source == "tmpfs" {
					continue
				}
				if strings.Contains(configFilepath, mountInfo.Mountpoint) {
					sourcePath, err := GetMountInfoSourcePath(mountInfo.Source)
					if err != nil {
						klog.Errorf("Get process %d ConfigFiles failed: %s", pid, err.Error())
					}

					remainingPath := strings.Replace(configFilepath, mountInfo.Mountpoint, "", -1)
					if remainingPath != "" {
						configFilepath = path.Join(sourcePath, mountInfo.Root, remainingPath)
					} else {
						configFilepath = path.Join(sourcePath, mountInfo.Root)
					}

					break
				}
			}

			processConfigFilepath := path.Join(hostPath, configFilepath)
			fileInfo, err := os.Stat(processConfigFilepath)
			if err != nil {
				klog.Infof("%d Get GetConfigfile %s content failed: %s", pid, configFilepath, err.Error())
				configFiles[configFilepath] = err.Error()
				continue
			}

			mode := fileInfo.Mode()

			if mode.IsDir() {
				configFiles[configFilepath] = "dir"
			} else if mode.IsRegular() {
				if mode&0111 != 0 {
					configFiles[configFilepath] = "executable"
				} else if fileInfo.Size() > 1024*1024*5 {
					configFiles[configFilepath] = "data"
				} else {
					contentBytes, err := os.ReadFile(processConfigFilepath)
					if err != nil {
						return nil, err
					}

					configFiles[configFilepath] = string(contentBytes)
				}
			} else {
				configFiles[configFilepath] = "other"
			}

		}

	}

	return configFiles, nil

}

// GetProcessStatusByPID xxx
func GetProcessStatusByPID(pid int32) (ProcessStatus, error) {
	var p *process.Process
	var err error
	var processStatus = ProcessStatus{}

	p, err = process.NewProcess(pid)
	if err != nil {
		klog.Errorf("Get processList failed: %s", err.Error())
		return processStatus, err
	}

	processStatus.Name, err = p.Name()
	if err != nil {
		return processStatus, err
	}

	processStatus.CreateTime, err = p.CreateTime()
	if err != nil {
		return processStatus, err
	}

	cpustat, err := p.Times()
	if err != nil {
		return processStatus, err
	}
	processStatus.CpuTime = cpustat.User + cpustat.System

	processStatus.Pid, err = p.Ppid()
	if err != nil {
		return processStatus, err
	}

	processStatus.Status, err = p.Status()
	if err != nil {
		return processStatus, err
	}

	return processStatus, nil
}

// GetProcessStatus xxx
func GetProcessStatus() ([]ProcessStatus, error) {
	var processList []*process.Process
	var err error

	processList, err = process.Processes()
	if err != nil {
		klog.Errorf("Get processList failed: %s", err.Error())
		return nil, err
	}

	processStatusList := make([]ProcessStatus, 0, 0)

	for _, p := range processList {
		processStatus := ProcessStatus{}
		processStatus.Name, err = p.Name()
		if err != nil {
			continue
		}

		processStatus.CreateTime, err = p.CreateTime()
		if err != nil {
			continue
		}

		cpustat, err := p.Times()
		if err != nil {
			continue
		}
		processStatus.CpuTime = cpustat.User + cpustat.System

		processStatus.Pid = p.Pid

		processStatus.Status, err = p.Status()
		if err != nil {
			klog.Errorf("Get process ppid status failed: %s", err.Error())
			continue
		}

		processStatusList = append(processStatusList, processStatus)
	}

	return processStatusList, nil
}

// GetProcessInfo xxx
func GetProcessInfo(exe string, id int32) (*ProcessInfo, error) {
	var processList []*process.Process
	var err error

	processList, err = process.Processes()
	if err != nil {
		klog.Errorf("Get processList failed: %s", err.Error())
		return nil, err
	}

	if exe == "" && id == 0 {
		return nil, fmt.Errorf("exe is %s, id is %d, not valid", exe, id)
	}

	processInfo := &ProcessInfo{
		ConfigFiles: make(map[string]string),
	}

	for _, p := range processList {
		if id != 0 && p.Pid != id {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		if exe != "" && !strings.Contains(exe, name) {
			continue
		}

		processInfo.Params, err = p.CmdlineSlice()
		if err != nil {
			klog.Errorf("Get process cmdline info failed: %s", err.Error())
			continue
		}

		if len(processInfo.Params) == 0 {
			continue
		}
		processInfo.BinaryPath = processInfo.Params[0]

		filename := filepath.Base(processInfo.BinaryPath)
		if filename != exe && exe != "" {
			continue
		}

		ppid, err := p.Ppid()
		if err != nil {
			klog.Errorf("Get process ppid info failed: %s", err.Error())
			continue
		}

		processInfo.Starter, err = GetStarter(p.Pid, ppid)
		if err != nil {
			klog.Errorf("Get process starter info failed: %s", err.Error())
		}

		//processInfo.Params = AddProcessParam(exe, processInfo.Params)
		if len(processInfo.Params) > 1 {
			processInfo.ConfigFiles, err = GetConfigfileList(processInfo.Params[1:], p.Pid)
			if err != nil {
				klog.Errorf("Get process %d ConfigFiles failed: %s", p.Pid, err.Error())
			}
		}

		if strings.HasSuffix(processInfo.Starter, ".service") {
			processInfo.ServiceFiles, err = GetProcessServiceConfigfiles(processInfo.Starter)
			if err != nil {
				klog.Infof("Get process %d ServiceFiles failed: %s", p.Pid, err.Error())
			}
		}

		processInfo.Status, err = p.Status()
		if err != nil {
			klog.Infof("Get process %d status failed: %s", p.Pid, err.Error())
		}

		return processInfo, nil
	}

	return nil, fmt.Errorf("%s process not found", exe)
}

// AddProcessParam xxx
func AddProcessParam(exe string, params []string) []string {
	// 有些进程会有默认读取的配置文件路径
	if strings.Contains(exe, "docker") {
		configFileFlag := false
		for _, param := range params {
			if strings.Contains(param, "config-file") {
				configFileFlag = true
				break
			}
		}
		if !configFileFlag {
			params = append(params, "--config-file")
			params = append(params, "/etc/docker/daemon.json")
		}
	} else if strings.Contains(exe, "containerd") {
		configFileFlag := false
		for _, param := range params {
			if strings.Contains(param, "--config") {
				configFileFlag = true
				break
			}
		}

		if !configFileFlag {
			params = append(params, "--config")
			params = append(params, "/etc/containerd/config.toml")
		}
	} else if strings.Contains(exe, "coredns") {
		params = append(params, "/etc/resolv.conf")
	}

	return params
}

// GetProcess xxx
func GetProcess(id int32) (*process.Process, error) {
	return process.NewProcess(id)
}
