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
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-logbeat-sidecar/config"

	"github.com/fsouza/go-dockerclient"
)

const (
	EnvLogInfoDataid       = "io.tencent.bcs.app.dataid"
	EnvLogInfoStdout       = "io.tencent.bcs.app.stdout"
	EnvLogInfoLogPath      = "io.tencent.bcs.app.logpath"
	EnvLogInfoLogCluster   = "io.tencent.bcs.app.cluster"
	EnvLogInfoLogNamepsace = "io.tencent.bcs.app.namespcae"
)

type SidecarController struct {
	sync.RWMutex

	conf   *config.Config
	client *docker.Client
	//key = containerid, value = ContainerLogConf
	logConfs map[string]*ContainerLogConf
	//log config tempalte
	logTemplate *template.Template
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

	//init log config template
	by, err := ioutil.ReadFile(s.conf.TemplateFile)
	if err != nil {
		return nil, err
	}

	s.logTemplate, err = template.New("LogConfig").Parse(string(by))
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
func produceLogConfParameter(container *docker.Container) (*LogConfParameter, bool) {
	para := &LogConfParameter{}
	for _, env := range container.Config.Env {
		key, val := strings.Split(env, "=")[0], strings.Split(env, "=")[1]
		switch key {
		case EnvLogInfoDataid:
			para.DataId = val
		case EnvLogInfoLogCluster:
			para.ClusterId = val
		case EnvLogInfoLogNamepsace:
			para.Namespace = val
		case EnvLogInfoStdout:
			if strings.ToLower(val) == "true" {
				para.stdout = true
			} else {
				para.stdout = false
			}
		case EnvLogInfoLogPath:
			para.nonstandardLog = val
		}
	}

	//if DataId, Namespace, ClusterId == nil, don't need collect the container log
	if para.DataId == "" || para.Namespace == "" || para.ClusterId == "" {
		blog.Warnf("container %s don't contain %s, %s, %s env",
			container.ID, EnvLogInfoDataid, EnvLogInfoLogNamepsace, EnvLogInfoLogCluster)
		return nil, false
	}

	//if nonstandard log
	if !para.stdout {
		if para.nonstandardLog == "" {
			blog.Warnf("container %s don't contain %s env", EnvLogInfoLogPath)
			return nil, false
		} else {
			para.LogFile = para.nonstandardLog
		}
		//else standard log
	} else {
		para.LogFile = container.LogPath
	}

	para.ContainerId = container.ID
	return para, true
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
		confPath:    fmt.Sprintf("%s/%s-%s.conf", s.conf.LogbeatDir, s.prefixFile, []byte(c.ID)[:12]),
	}
	para, ok := produceLogConfParameter(c)
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

		err = s.logTemplate.Execute(f, para)
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
