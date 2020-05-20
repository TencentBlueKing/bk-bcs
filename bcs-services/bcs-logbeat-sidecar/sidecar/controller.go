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
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-logbeat-sidecar/config"
	"bk-bcs/bcs-services/bcs-logbeat-sidecar/types"
	bkbcsv1 "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	ContainerLabelK8sContainerName = "io.kubernetes.container.name"
	ContainerLabelK8sPodName       = "io.kubernetes.pod.name"
	ContainerLabelK8sPodNameSpace  = "io.kubernetes.pod.namespace"
)

type SidecarController struct {
	sync.RWMutex

	conf   *config.Config
	client *docker.Client
	//key = containerid, value = ContainerLogConf
	logConfs map[string]*ContainerLogConf
	//log config prefix file name
	prefixFile string

	//pod Lister
	podLister corev1.PodLister
	//apiextensions clientset
	extensionClientset *apiextensionsclient.Clientset
	//BcsLogConfig Lister
	bcsLogConfigLister   bkbcsv1.BcsLogConfigLister
	bcsLogConfigInformer cache.SharedIndexInformer
}

type ContainerLogConf struct {
	containerId string
	confPath    string
	//needCollect bool
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
		blog.Errorf("new dockerclient %s failed: %s", conf.DockerSock, err.Error())
		return nil, err
	}

	//init kubeconfig
	err = s.initKubeconfig()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SidecarController) Start() {
	go s.listenerDockerEvent()
	//go s.tickerSyncContainerLogConfs()
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
		case "destroy":
			s.Lock()
			s.deleteContainerLogConf(msg.ID)
			s.Unlock()
		}
	}
}

func (s *SidecarController) syncLogConfs() {
	//list all running containers
	apiContainers, err := s.client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		blog.Errorf("docker ListContainers failed: %s", err.Error())
		return
	}

	for _, apiC := range apiContainers {
		c, err := s.client.InspectContainer(apiC.ID)
		if err != nil {
			blog.Errorf("docker InspectContainer %s failed: %s", apiC.ID, err.Error())
			continue
		}

		s.produceContainerLogConf(c)
	}

	//remove invalid logconfig file
	s.removeInvalidLogConfigFile()
}

func (s *SidecarController) removeInvalidLogConfigFile() {
	files, err := ioutil.ReadDir(s.conf.LogbeatDir)
	if err != nil {
		blog.Errorf("ReadDir %s failed: %s", s.conf.LogbeatDir, err.Error())
		return
	}

	for _, o := range files {
		confKey := fmt.Sprintf("%s/%s", s.conf.LogbeatDir, o.Name())
		s.RLock()
		_, ok := s.logConfs[confKey]
		s.RUnlock()
		if ok {
			continue
		}
		err := os.Remove(confKey)
		if err != nil {
			blog.Errorf("remove invalid logconfig file %s error %s", confKey, err.Error())
		} else {
			blog.Infof("remove invalid logconfig file %s success", confKey)
		}
	}
}

/*// if need to collect the container logs, return true
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
}*/

func (s *SidecarController) getContainerLogConfKey(containerId string) string {
	return fmt.Sprintf("%s/%s-%s.yaml", s.conf.LogbeatDir, s.prefixFile, []byte(containerId)[:12])
}

func (s *SidecarController) produceContainerLogConf(c *docker.Container) {
	key := s.getContainerLogConfKey(c.ID)
	y, ok := s.produceLogConfParameterV2(c)
	//the container don't match any BcsLogConfig
	if !ok {
		s.Lock()
		defer s.Unlock()
		_, ok := s.logConfs[key]
		//if the container have logconfig, then delete it
		if ok {
			s.deleteContainerLogConf(c.ID)
			delete(s.logConfs, key)
		}
		blog.Infof("container %s don't need collect log", c.ID)
		return
	}

	logConf := &ContainerLogConf{
		containerId: c.ID,
		confPath:    key,
	}
	by, _ := yaml.Marshal(y)
	blog.Infof("container %s need been collected log, and LogConfig(%s)", c.ID, string(by))
	f, err := os.Create(logConf.confPath)
	if err != nil {
		blog.Errorf("container %s open file %s failed: %s", c.ID, logConf.confPath, err.Error())
		return
	}
	defer f.Close()

	_, err = f.Write(by)
	if err != nil {
		blog.Errorf("container %s tempalte execute failed: %s", c.ID, err.Error())
		return
	}
	blog.Infof("produce container %s log config %s success", c.ID, logConf.confPath)
	s.Lock()
	s.logConfs[key] = logConf
	s.Unlock()
	return
}

func (s *SidecarController) deleteContainerLogConf(containerId string) {
	key := s.getContainerLogConfKey(containerId)
	logConf, ok := s.logConfs[key]
	if !ok {
		blog.Infof("container %s don't have LogConfig, then ignore", containerId)
		return
	}
	err := os.Remove(logConf.confPath)
	if err != nil {
		blog.Errorf("remove log config %s error %s", logConf.confPath, err.Error())
		return
	}
	delete(s.logConfs, key)
	blog.Infof("delete container %s log config success", containerId)
}

// if need to collect the container logs, return true
// else return false
func (s *SidecarController) produceLogConfParameterV2(container *docker.Container) (*types.Yaml, bool) {
	//if container is network, ignore
	name := container.Config.Labels[ContainerLabelK8sContainerName]
	if name == "POD" || name == "" {
		blog.Infof("container %s is network container, ignore", container.ID)
		return nil, false
	}
	podName := container.Config.Labels[ContainerLabelK8sPodName]
	podNameSpace := container.Config.Labels[ContainerLabelK8sPodNameSpace]
	pod, err := s.podLister.Pods(podNameSpace).Get(podName)
	if err != nil {
		blog.Errorf("container %s fetch pod(%s:%s) error %s", container.ID, podName, podNameSpace, err.Error())
		return nil, false
	}

	para := types.Local{
		ExtMeta: make(map[string]string),
	}
	logConf := s.getPodLogConfigCrd(container, pod)
	//if logConf==nil, container not match BcsLogConfig
	if logConf == nil {
		blog.Warnf("container %s don't match BcsLogConfig", container.ID)
		return nil, false
	}

	para.ExtMeta["io_tencent_bcs_cluster"] = logConf.Spec.ClusterId
	para.ExtMeta["io_tencent_bcs_namespace"] = pod.Namespace
	para.ExtMeta["io_tencent_bcs_server_name"] = pod.OwnerReferences[0].Name
	para.ExtMeta["io_tencent_bcs_type"] = pod.OwnerReferences[0].Kind
	para.ExtMeta["io_tencent_bcs_appid"] = pod.Annotations["io.tencent.bcs.app.appid"]
	para.ExtMeta["io_tencent_bcs_projectid"] = pod.Annotations["io.tencent.paas.projectid"]
	para.ExtMeta["container_id"] = container.ID
	para.ExtMeta["container_hostname"] = container.Config.Hostname
	para.ToJson = true
	if len(logConf.Spec.ContainerConfs) > 0 {
		for _, conf := range logConf.Spec.ContainerConfs {
			if conf.ContainerName == name {
				para.StdoutDataid = conf.StdDataId
				para.NonstandardDataid = conf.NonStdDataId
				para.NonstandardPaths = conf.LogPaths
				para.LogTags = conf.LogTags
				break
			}
		}
	} else {
		para.StdoutDataid = logConf.Spec.StdDataId
		para.NonstandardDataid = logConf.Spec.NonStdDataId
		para.NonstandardPaths = logConf.Spec.LogPaths
		para.LogTags = logConf.Spec.LogTags
	}
	//whether report pod labels to log tags
	if logConf.Spec.PodLabels {
		for k, v := range pod.Labels {
			para.ExtMeta[k] = v
		}
	}
	//custom log tags
	for k, v := range para.LogTags {
		para.ExtMeta[k] = v
	}

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
		inLocal := para
		for _, f := range para.NonstandardPaths {
			inLocal.Paths = append(inLocal.Paths, fmt.Sprintf("/proc/%d/root%s", container.State.Pid, f))
		}
		i, _ := strconv.Atoi(para.NonstandardDataid)
		inLocal.DataId = i
		y.Local = append(y.Local, inLocal)
	}

	return y, true
}
