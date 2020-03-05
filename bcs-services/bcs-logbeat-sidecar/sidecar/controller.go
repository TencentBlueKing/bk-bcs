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

package sidecar

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-logbeat-sidecar/config"
	"bk-bcs/bcs-services/bcs-logbeat-sidecar/types"

	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
)

const (
	//report container stdout dataid
	EnvLogInfoStdoutDataid = "io_tencent_bcs_app_stdout_dataid_v2"
	//report container nonstandard log dataid
	EnvLogInfoNonstandardDataid = "io_tencent_bcs_app_nonstandard_dataid_v2"
	//if true, then stdout; else custom logs file
	EnvLogInfoStdout = "io_tencent_bcs_app_stdout_v2"
	//if stdout=false, log file path
	EnvLogInfoLogPath = "io_tencent_bcs_app_logpath_v2"
	//clusterid
	EnvLogInfoLogCluster = "io_tencent_bcs_app_cluster_v2"
	//namespace
	EnvLogInfoLogNamepsace = "io_tencent_bcs_app_namespcae_v2"
	//custom labels, log tags
	//example: kv1:val1,kv2:val2,kv3:val3...
	EnvLogInfoLogLabel = "io_tencent_bcs_app_label_v2"
	//application or deployment't name
	EnvLogInfoLogServerName = "io_tencent_bcs_controller_name"
	//enum: Application、Deployment...
	EnvLogInfoLogType = "io_tencent_bcs_controller_type"
)

type SidecarController struct {
	sync.RWMutex

	conf   *config.Config
	client *docker.Client
	//key = containerid, value = ContainerLogConf
	logConfs map[string]*ContainerLogConf
	//log config prefix file name
	prefixFile string
}

type ContainerLogConf struct {
	containerId string
	confPath    string
	needCollect bool
}

type LogConfParameter struct {
	LogFile     string
	DataId      string
	ContainerId string
	ClusterId   string
	Namespace   string
	//application or deployment't name
	ServerName string
	//application or deployment
	ServerType string
	//custom label
	CustemLabel string

	stdout         bool
	nonstandardLog string
}

func NewSidecarController(conf *config.Config) (*SidecarController, error) {
	var err error
	s := &SidecarController{
		conf:       conf,
		logConfs:   make(map[string]*ContainerLogConf),
		prefixFile: conf.PrefixFile,
	}

	//init docker client
	s.client, err = docker.NewClient(conf.DockerSock)
	if err != nil {
		return nil, err
	}

	err = s.syncLogConfs()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SidecarController) Start() {
	go s.listenerDockerEvent()
	go s.tickerSyncContainerLogConfs()
}

//start listen docker api event
//when create container, and produce container log config
//when stop container, and delete container log config
func (s *SidecarController) listenerDockerEvent() {
	listener := make(chan *docker.APIEvents)
	err := s.client.AddEventListener(listener)
	if err != nil {
		blog.Errorf("listen docker event error %s", err.Error())
		os.Exit(1)
	}

	defer func() {
		err = s.client.RemoveEventListener(listener)
		if err != nil {
			blog.Errorf("remove docker event error  %s", err.Error())
		}
	}()

	for {
		var msg *docker.APIEvents
		select {
		case msg = <-listener:
			blog.Infof("receive docker event action %s container %s", msg.Action, msg.ID)
		}

		switch msg.Action {
		//start container
		case "start":
			c, err := s.client.InspectContainer(msg.ID)
			if err != nil {
				blog.Errorf("inspect container %s error %s", msg.ID, err.Error())
				break
			}
			s.produceContainerLogConf(c)

		// stop container
		case "stop":
			s.deleteContainerLogConf(msg.ID)
		}
	}
}

func (s *SidecarController) tickerSyncContainerLogConfs() {
	ticker := time.NewTicker(time.Minute * 10)

	for {
		select {
		case <-ticker.C:
			err := s.syncLogConfs()
			if err != nil {
				blog.Errorf("sync log config failed: %s", err.Error())
			}
		}
	}
}

func (s *SidecarController) syncLogConfs() error {
	blog.Infof("sync all container log configs start...")
	//list all running containers
	apiContainers, err := s.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, apiC := range apiContainers {
		c, err := s.client.InspectContainer(apiC.ID)
		if err != nil {
			return err
		}

		s.produceContainerLogConf(c)
	}
	blog.Infof("sync all container log configs done")
	return nil
}

// if need to collect the container logs, return true
// else return false
func produceLogConfParameter(container *docker.Container) (*types.Yaml, bool) {
	para := types.Local{
		ExtMeta: make(map[string]string),
	}
	for _, env := range container.Config.Env {
		key, val := strings.Split(env, "=")[0], strings.Split(env, "=")[1]
		switch key {
		case EnvLogInfoStdoutDataid:
			para.StdoutDataid = val
		case EnvLogInfoNonstandardDataid:
			para.NonstandardDataid = val
		case EnvLogInfoLogCluster:
			para.ExtMeta["io_tencent_bcs_cluster"] = val
		case EnvLogInfoLogNamepsace:
			para.ExtMeta["io_tencent_bcs_namespace"] = val
		case EnvLogInfoLogPath:
			para.NonstandardPaths = val
		case EnvLogInfoLogServerName:
			para.ExtMeta["io_tencent_bcs_server_name"] = val
		case EnvLogInfoLogType:
			para.ExtMeta["io_tencent_bcs_type"] = val
		case EnvLogInfoLogLabel:
			array := strings.Split(val, ",")
			for _, o := range array {
				kvs := strings.Split(o, ":")
				if len(kvs) != 2 {
					blog.Infof("container %s env %s value %s is invalid", container.ID, EnvLogInfoLogLabel, val)
					continue
				}
				para.ExtMeta[kvs[0]] = kvs[1]
			}
		}
	}

	//if DataId, Namespace, ClusterId == nil, don't need collect the container log
	if para.StdoutDataid == "" && para.NonstandardDataid == "" {
		blog.Warnf("container %s don't contain %s or %s env",
			container.ID, EnvLogInfoStdoutDataid, EnvLogInfoNonstandardDataid)
		return nil, false
	}
	//container id
	para.ExtMeta["container_id"] = container.ID
	para.ToJson = true
	y := &types.Yaml{Local: make([]types.Local, 0)}
	//if stdout container log
	if para.StdoutDataid != "" {
		inLocal := para
		inLocal.Paths = []string{container.LogPath}
		i, _ := strconv.Atoi(para.StdoutDataid)
		inLocal.DataId = i
		y.Local = append(y.Local, inLocal)
	}
	//if nonstandard Log
	if para.NonstandardDataid != "" {
		array := strings.Split(para.NonstandardPaths, ",")
		inLocal := para
		for _, f := range array {
			inLocal.Paths = append(inLocal.Paths, fmt.Sprintf("/proc/%d/root%s", container.State.Pid, f))
		}
		i, _ := strconv.Atoi(para.NonstandardDataid)
		inLocal.DataId = i
		y.Local = append(y.Local, inLocal)
	}
	//if len(files)==0, then invalid
	if len(y.Local) == 0 {
		blog.Warnf("container %s env(%s, %s) is invalid", container.ID, EnvLogInfoStdout, EnvLogInfoLogPath)
		return nil, false
	}
	return y, true
}

func (s *SidecarController) produceContainerLogConf(c *docker.Container) {
	s.RLock()
	_, ok := s.logConfs[c.ID]
	s.RUnlock()
	if ok {
		blog.V(3).Infof("container %s already under SidecarController manager", c.ID)
		return
	}

	logConf := &ContainerLogConf{
		containerId: c.ID,
		confPath:    fmt.Sprintf("%s/%s-%s.yaml", s.conf.LogbeatDir, s.prefixFile, []byte(c.ID)[:12]),
	}
	y, ok := produceLogConfParameter(c)
	if !ok {
		blog.Warnf("container %s don't need collect log file", c.ID)
		logConf.needCollect = false
		s.Lock()
		s.logConfs[c.ID] = logConf
		s.Unlock()
		return
	}

	logConf.needCollect = true
	_, err := os.Stat(logConf.confPath)
	//if confpath not exist, then create it
	if err != nil {
		f, err := os.Create(logConf.confPath)
		if err != nil {
			blog.Errorf("container %s open file %s failed: %s", c.ID, logConf.confPath, err.Error())
			return
		}
		defer f.Close()

		by, _ := yaml.Marshal(y)
		_, err = f.Write(by)
		if err != nil {
			blog.Errorf("container %s tempalte execute failed: %s", c.ID, err.Error())
		}
		blog.Infof("produce container %s log config %s success", c.ID, logConf.confPath)
	} else {
		blog.Infof("container %s log config %s already exist, then don't need create it", c.ID, logConf.confPath)
	}

	s.Lock()
	s.logConfs[c.ID] = logConf
	s.Unlock()
	return
}

func (s *SidecarController) deleteContainerLogConf(containerId string) {
	s.RLock()
	logConf, ok := s.logConfs[containerId]
	s.RUnlock()
	if !ok {
		return
	}

	if logConf.needCollect {
		err := os.Remove(logConf.confPath)
		if err != nil {
			blog.Errorf("remove log config %s error %s", logConf.confPath, err.Error())
		}
	}

	s.Lock()
	delete(s.logConfs, containerId)
	s.Unlock()
	blog.Infof("delete container %s log config success", containerId)
}
