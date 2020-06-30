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

package manager

import (
	"bytes"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-daemon/process-daemon/config"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-daemon/process-daemon/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
	"io/ioutil"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	ProcessHeartBeatPeriodSeconds = 600
)

type manager struct {
	sync.RWMutex

	processInfos map[string]*types.ProcessInfo
	store        store.Store

	processLocks  map[string]*sync.Mutex
	processRWlock sync.RWMutex

	cli *httpclient.HttpClient

	conf *config.Config
}

func NewManager(conf *config.Config) Manager {
	m := &manager{
		processInfos: make(map[string]*types.ProcessInfo, 0),
		processLocks: make(map[string]*sync.Mutex, 0),
		store:        store.NewStore("./data"),
		conf:         conf,
	}
	m.initCli()
	return m
}

func (m *manager) Init() error {
	processInfos, err := m.store.GetAllProcessInfos()
	if err != nil {
		blog.Errorf("get all processinfos error %s", err.Error())
		return err
	}

	m.Lock()
	for _, processInfo := range processInfos {
		m.processInfos[processInfo.Id] = processInfo
	}
	m.Unlock()

	return nil
}

func (m *manager) Start() {
	//check process status
	go m.loopCheckProcess()
}

func (m *manager) GetConfig() *config.Config {
	return m.conf
}

func (m *manager) lockObjectKey(key string) {

	m.processRWlock.RLock()
	myLock, ok := m.processLocks[key]
	m.processRWlock.RUnlock()
	if ok {
		myLock.Lock()
		return
	}

	m.processRWlock.Lock()
	myLock, ok = m.processLocks[key]
	if !ok {
		blog.Info("create process lock(%s), current locknum(%d)", key, len(m.processLocks))
		m.processLocks[key] = new(sync.Mutex)
		myLock, _ = m.processLocks[key]
	}
	m.processRWlock.Unlock()

	myLock.Lock()
	return
}

func (m *manager) unLockObjectKey(key string) {
	m.processRWlock.RLock()
	myLock, ok := m.processLocks[key]
	m.processRWlock.RUnlock()

	if !ok {
		blog.Error("process lock(%s) not exist when do unlock", key)
		return
	}
	myLock.Unlock()
}

func (m *manager) HeartBeat(heartbeat *types.HeartBeat) {
	m.Lock()
	defer m.Unlock()

	process, ok := m.processInfos[heartbeat.ProcessId]
	if !ok {
		blog.Errorf("heartbeat process %s not found", heartbeat.ProcessId)
		return
	}

	m.lockObjectKey(process.Id)
	process.ExecutorHeartBeatTime = time.Now().Unix()
	err := m.store.StoreProcessInfo(process)
	if err != nil {
		blog.Errorf("store processInfo %s error %s", process.Id, err.Error())
	}
	m.unLockObjectKey(process.Id)

	return
}

func (m *manager) CreateProcess(processInfo *types.ProcessInfo) error {
	m.Lock()
	defer m.Unlock()

	_, ok := m.processInfos[processInfo.Id]
	if ok {
		blog.Errorf("processinfo %s already existed", processInfo.Id)
		return fmt.Errorf("processinfo %s already existed", processInfo.Id)
	}

	status := &types.ProcessStatusInfo{
		Id:           processInfo.Id,
		Status:       types.ProcessStatusStaging,
		RegisterTime: time.Now().Unix(),
	}
	processInfo.StatusInfo = status
	processInfo.ExecutorHeartBeatTime = time.Now().Unix()

	for _, uri := range processInfo.Uris {
		arr := strings.Split(uri.Value, "/")
		fileName := arr[len(arr)-1]

		uri.PackagesFile = filepath.Join(m.conf.WorkspaceDir, "packages_dir", fileName)
		uri.ExtractDir = filepath.Join(m.conf.WorkspaceDir, "extract_dir", strings.TrimRight(fileName, ".tar.gz"))
	}

	m.processInfos[processInfo.Id] = processInfo
	err := m.store.StoreProcessInfo(processInfo)
	if err != nil {
		blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
		return err
	}

	by, _ := json.Marshal(processInfo)
	blog.Infof("create process %s info %s", processInfo.Id, string(by))
	go m.startProcess(processInfo)

	return nil
}

func (m *manager) InspectProcessStatus(processId string) (*types.ProcessStatusInfo, error) {
	m.RLock()
	defer m.RUnlock()

	process, ok := m.processInfos[processId]
	if !ok {
		return nil, fmt.Errorf("process %s not found", processId)
	}

	return process.StatusInfo, nil
}

func (m *manager) StopProcess(processId string, timeout int) error {
	m.RLock()
	processInfo, ok := m.processInfos[processId]
	m.RUnlock()
	if !ok {
		blog.Errorf("process %s not found", processId)
		return fmt.Errorf("process %s not found", processId)
	}

	m.lockObjectKey(processId)
	defer m.unLockObjectKey(processId)

	if processInfo.StatusInfo.Status == types.ProcessStatusStopped {
		blog.Infof("process %s is done", processInfo.Id)
		return nil
	}

	if processInfo.StatusInfo.Status != types.ProcessStatusRunning {
		blog.Errorf("process %s status %s can't be stopped", processInfo.Id, processInfo.StatusInfo.Status)
		return fmt.Errorf("process %s status %s can't be stopped", processInfo.Id, processInfo.StatusInfo.Status)
	}

	processInfo.StatusInfo.Status = types.ProcessStatusStopping
	err := m.store.StoreProcessInfo(processInfo)
	if err != nil {
		blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
	}

	blog.Infof("stop process %s pid %d start...", processInfo.Id, processInfo.StatusInfo.Pid)
	cmd := exec.Cmd{
		Path: processInfo.StopCmd,
		Dir:  processInfo.WorkDir,
		Env:  processInfo.Envs,
	}
	buf := bytes.NewBuffer(make([]byte, 1024))
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		blog.Errorf("stop process %s work_dir %s exec stopcmd %s stderr %s error %s", processInfo.Id,
			processInfo.WorkDir, processInfo.StopCmd, buf.String(), err.Error())
		//processInfo.StatusInfo.Status = types.ProcessStatusRunning
		if buf.String() != "" {
			processInfo.StatusInfo.Message = buf.String()
		} else {
			processInfo.StatusInfo.Message = err.Error()
		}

		err = m.store.StoreProcessInfo(processInfo)
		if err != nil {
			blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
		}
	}

	ticker := time.NewTicker(time.Second * time.Duration(timeout))
	//loop check proc
ForResp:
	for {
		var proc *os.Process
		var err error

		select {
		case <-ticker.C:
			proc, _, err = m.processIsOk(processInfo)
			if err != nil {
				blog.Errorf("stop process %s not found pid %d process", processInfo.Id, processInfo.StatusInfo.Pid)
				processInfo.StatusInfo.Status = types.ProcessStatusStopped
				processInfo.StatusInfo.Message = fmt.Sprintf("process %s is done", processInfo.Id)
				err = m.store.StoreProcessInfo(processInfo)
				if err != nil {
					blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
				}
				return nil
			}

			err = proc.Kill()
			if err != nil {
				blog.Errorf("enforce kill -9 process %s pid %d error %s", processInfo.Id, processInfo.StatusInfo.Pid, err.Error())
			} else {
				blog.Infof("enforce kill -9 process %s pid %d success", processInfo.Id, processInfo.StatusInfo.Pid)
				proc.Release()
			}
			break ForResp //break loop check proc

		default:
			proc, _, err = m.processIsOk(processInfo)
			if err != nil {
				blog.Errorf("stop process %s not found pid %d process", processInfo.Id, processInfo.StatusInfo.Pid)
				processInfo.StatusInfo.Status = types.ProcessStatusStopped
				processInfo.StatusInfo.Message = fmt.Sprintf("process %s is done", processInfo.Id)
				err = m.store.StoreProcessInfo(processInfo)
				if err != nil {
					blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
				}
				return nil
			}
			blog.Infof("process %s check proc running", processInfo.Id)
			time.Sleep(time.Second)
		}

	}

	//if process pid not exist,then stop success
	_, _, err = m.processIsOk(processInfo)
	if err != nil {
		blog.Infof("stop process %s pid %d success", processInfo.Id, processInfo.StatusInfo.Pid)
		processInfo.StatusInfo.Status = types.ProcessStatusStopped
		processInfo.StatusInfo.Message = fmt.Sprintf("process %s is done", processInfo.Id)
		err = m.store.StoreProcessInfo(processInfo)
		if err != nil {
			blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
		}
		return nil
	}

	processInfo.StatusInfo.Status = types.ProcessStatusRunning
	processInfo.StatusInfo.Message = fmt.Sprintf("stop process %s pid %d failed", processInfo.Id, processInfo.StatusInfo.Pid)
	err = m.store.StoreProcessInfo(processInfo)
	if err != nil {
		blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
	}

	blog.Errorf("stop process %s pid %d failed", processInfo.Id, processInfo.StatusInfo.Pid)
	return fmt.Errorf("stop process %s pid %d failed", processInfo.Id, processInfo.StatusInfo.Pid)
}

func (m *manager) DeleteProcess(processId string) error {
	m.Lock()
	defer m.Unlock()

	processInfo, ok := m.processInfos[processId]
	if !ok {
		blog.Errorf("delete process %s, but not found", processId)
		return nil
	}

	if processInfo.StatusInfo.Status != types.ProcessStatusStopped {
		blog.Errorf("process %s status %s can't be deleted", processInfo.Id, processInfo.StatusInfo.Status)
		return fmt.Errorf("process %s status %s can't be deleted", processInfo.Id, processInfo.StatusInfo.Status)
	}

	if processInfo.Uris != nil && len(processInfo.Uris) > 0 {
		err := os.Remove(processInfo.Uris[0].OutputDir)
		if err != nil {
			blog.Errorf("process %s remove file %s error %s", processInfo.Id, processInfo.Uris[0].OutputDir, err.Error())
		}
	}

	delete(m.processInfos, processInfo.Id)
	blog.Infof("delete process %s success", processInfo.Id)

	/*err := m.store.DeleteProcessInfo(processInfo)
	if err!=nil {
		blog.Errorf("store delete processinfo %s error %s",processInfo.Id,err.Error())
	}*/

	return nil
}

func (m *manager) symlinkWorkdir(processInfo *types.ProcessInfo) error {
	if processInfo.Uris == nil || len(processInfo.Uris) == 0 {
		return nil
	}

	uriPack := processInfo.Uris[0]
	//whether uripack.outputdir exists
	_, err := os.Stat(uriPack.OutputDir)
	if err != nil {
		blog.Errorf("process %s stat file %s error %s", processInfo.Id, uriPack.User, err.Error())
		err = os.MkdirAll(filepath.Dir(uriPack.OutputDir), 0755)
		if err != nil {
			blog.Errorf("process %s mkdir %s error %s", processInfo.Id, filepath.Dir(uriPack.OutputDir), err.Error())
			return err
		}
	} else {
		err = os.Remove(uriPack.OutputDir)
		if err != nil {
			blog.Errorf("process %s remove file %s error %s", processInfo.Id, uriPack.OutputDir, err.Error())
			return err
		}
	}

	err = os.Symlink(uriPack.ExtractDir, uriPack.OutputDir)
	if err != nil {
		blog.Errorf("process %s symlink file %s error %s", processInfo.Id, uriPack.OutputDir, err.Error())
		return err
	}
	blog.Infof("Symlink from %s to %s success", uriPack.OutputDir, uriPack.ExtractDir)

	return nil
}

func (m *manager) startProcess(processInfo *types.ProcessInfo) {
	m.lockObjectKey(processInfo.Id)
	defer m.unLockObjectKey(processInfo.Id)

	err := m.downloadAndTarProcessPackages(processInfo)
	if err != nil {
		processInfo.StatusInfo.Status = types.ProcessStatusStopped
		processInfo.StatusInfo.Message = err.Error()
		err = m.store.StoreProcessInfo(processInfo)
		if err != nil {
			blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
		}
		return
	}

	err = m.symlinkWorkdir(processInfo)
	if err != nil {
		processInfo.StatusInfo.Status = types.ProcessStatusStopped
		processInfo.StatusInfo.Message = err.Error()
		err = m.store.StoreProcessInfo(processInfo)
		if err != nil {
			blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
		}
		return
	}

	cmd := exec.Cmd{
		Path: processInfo.StartCmd,
		Args: processInfo.Argv,
		Dir:  processInfo.WorkDir,
		Env:  processInfo.Envs,
	}

	//lookup proc user, example user00
	if processInfo.User != "" {
		u, err := user.Lookup(processInfo.User)
		if err != nil {
			blog.Errorf("process %s lookup user %s error %s", processInfo.Id, processInfo.User, err.Error())
		} else {
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)},
			}
		}
	}

	buf := bytes.NewBuffer(make([]byte, 1024))
	cmd.Stderr = buf
	err = cmd.Run()
	//_,err := os.StartProcess(processInfo.StartCmd,processInfo.Argv,attr)
	if err != nil {
		blog.Errorf("start process %s startcmd %s stderr %s error %s", processInfo.Id, processInfo.StartCmd, buf.String(), err.Error())
		processInfo.StatusInfo.Status = types.ProcessStatusStopped
		if buf.String() != "" {
			processInfo.StatusInfo.Message = buf.String()
		} else {
			processInfo.StatusInfo.Message = err.Error()
		}
		err = m.store.StoreProcessInfo(processInfo)
		if err != nil {
			blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
		}
		return
	}

	blog.Infof("start process %s success, and status %s", processInfo.Id, types.ProcessStatusStarting)

	processInfo.StatusInfo.Status = types.ProcessStatusStarting
	processInfo.StatusInfo.LastStartTime = time.Now().Unix()
	err = m.store.StoreProcessInfo(processInfo)
	if err != nil {
		blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
	}

	return
}

func (m *manager) loopCheckProcess() {
	for {
		time.Sleep(time.Second * 5)

		m.Lock()
		for _, process := range m.processInfos {
			//todo
			/*if time.Now().Unix()-process.ExecutorHeartBeatTime>ProcessHeartBeatPeriodSeconds {
				if process.StatusInfo.Status==types.ProcessStatusStopped {
					blog.Errorf("process %s status %s heartbeat timeout, and delete it",process.Id,
						process.StatusInfo.Status)
					err := m.store.DeleteProcessInfo(process)
					if err!=nil {
						blog.Errorf("store delete processinfo %s error %s",process.Id,err.Error())
					}
					delete(m.processInfos,process.Id)
					continue
				}

				if process.StatusInfo.Status==types.ProcessStatusRunning {
					blog.Errorf("process %s status %s heartbeat timeout, and stop it",process.Id,
						process.StatusInfo.Status)
					go m.StopProcess(process.Id,600)
					continue
				}
			}*/

			if process.StatusInfo.Status != types.ProcessStatusStarting &&
				process.StatusInfo.Status != types.ProcessStatusRunning {
				continue
			}

			if process.StatusInfo.Status == types.ProcessStatusStarting &&
				(time.Now().Unix()-process.StatusInfo.LastStartTime) < process.StartGracePeriod {
				continue
			}

			go m.checkProcess(process)
		}
		m.Unlock()
	}
}

func (m *manager) checkProcess(processInfo *types.ProcessInfo) {
	m.lockObjectKey(processInfo.Id)
	defer m.unLockObjectKey(processInfo.Id)

	oldStatus := processInfo.StatusInfo.Status
	_, pid, err := m.processIsOk(processInfo)
	if err != nil {
		processInfo.StatusInfo.Status = types.ProcessStatusStopped
		processInfo.StatusInfo.Message = err.Error()
	} else {
		processInfo.StatusInfo.Status = types.ProcessStatusRunning
		processInfo.StatusInfo.Message = fmt.Sprintf("process %s is running", processInfo.Id)
		processInfo.StatusInfo.Pid = pid
	}

	if oldStatus != processInfo.StatusInfo.Status {
		blog.Infof("process %s status from %s change to %s message %s", processInfo.Id, oldStatus,
			processInfo.StatusInfo.Status, processInfo.StatusInfo.Message)
	}

	err = m.store.StoreProcessInfo(processInfo)
	if err != nil {
		blog.Errorf("store processInfo %s error %s", processInfo.Id, err.Error())
	}
}

func (m *manager) processIsOk(processInfo *types.ProcessInfo) (*os.Process, int, error) {
	by, err := ioutil.ReadFile(processInfo.PidFile)
	if err != nil {
		blog.Errorf("read process %s pid file error %s", processInfo.Id, err.Error())
		return nil, 0, err
	}

	pid, err := strconv.Atoi(string(by))
	if err != nil {
		blog.Errorf("process %s conv pid %s to int error %s", processInfo.Id, string(by), err.Error())
		return nil, 0, err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		blog.Errorf("find process %s pid %d error %s", processInfo.Id, pid, err.Error())
		return nil, 0, fmt.Errorf("process %s is done", processInfo.Id)
	}

	procPath, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		blog.Errorf("process %s readlink error %s", processInfo.Id, err.Error())
		return nil, 0, fmt.Errorf("process %s is done", processInfo.Id)
	}

	if processInfo.ProcessName != path.Base(procPath) {
		blog.Errorf("process %s pid %d procName is %s", processInfo.Id, pid, path.Base(procPath))
		return nil, 0, fmt.Errorf("process %s is done", processInfo.Id)
	}

	return proc, pid, nil
}
