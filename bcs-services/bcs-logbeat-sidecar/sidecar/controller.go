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
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/types"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	bkbcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"

	docker "github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// ContainerLabelK8sContainerName container name key in containers labels
	ContainerLabelK8sContainerName = "io.kubernetes.container.name"
	// ContainerLabelK8sPodName pod name key in containers labels
	ContainerLabelK8sPodName = "io.kubernetes.pod.name"
	// ContainerLabelK8sPodNameSpace pod namespace key in containers labels
	ContainerLabelK8sPodNameSpace = "io.kubernetes.pod.namespace"
)

// SidecarController controls the behaviour of BcsLogConfig CRD
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

// ContainerLogConf record the log config for container
type ContainerLogConf struct {
	confPath string
	data     []byte
	yamlData *types.Yaml
}

// LogConfParameter is no longer used
type LogConfParameter struct {
	LogFile     string
	DataID      string
	ContainerID string
	ClusterID   string
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

// NewSidecarController returns a new bcslogconfigs controller
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

	//mkdir logconfig dir
	err = os.MkdirAll(conf.LogbeatDir, os.ModePerm)
	if err != nil {
		blog.Errorf("mkdir %s failed: %s", conf.LogbeatDir, err.Error())
		return nil, err
	}
	s.initLogConfigs()
	//init kubeconfig
	err = s.initKubeconfig()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Start starts the controller
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
			blog.V(3).Infof("receive docker event action %s container %s", msg.Action, msg.ID)
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
		blog.Warnf("ReadDir %s failed: %s", s.conf.LogbeatDir, err.Error())
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

func (s *SidecarController) initLogConfigs() {
	fileList, err := ioutil.ReadDir(s.conf.LogbeatDir)
	if err != nil {
		blog.Errorf("initLogConfigs readdir %s failed: %s", s.conf.LogbeatDir, err.Error())
		return
	}

	for _, f := range fileList {
		key := fmt.Sprintf("%s/%s", s.conf.LogbeatDir, f.Name())
		by, err := ioutil.ReadFile(key)
		if err != nil {
			blog.Errorf("read file %s failed: %s", key, err.Error())
			continue
		}

		conf := &ContainerLogConf{
			confPath: key,
			data:     by,
		}
		s.logConfs[key] = conf
	}
}

func (s *SidecarController) getContainerLogConfKey(containerID string) string {
	return fmt.Sprintf("%s/%s-%s.yaml", s.conf.LogbeatDir, s.prefixFile, []byte(containerID)[:12])
}

func (s *SidecarController) produceContainerLogConf(c *docker.Container) {
	key := s.getContainerLogConfKey(c.ID)
	y, ok := s.produceLogConfParameterV2(c)
	by, _ := yaml.Marshal(y)
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
		return
	}
	//if log config exist, and not changed
	s.RLock()
	logConf, _ := s.logConfs[key]
	s.RUnlock()
	if logConf != nil {
		if string(by) == string(logConf.data) {
			blog.Infof("container %s log config %s not changed", c.ID, logConf.confPath)
			if logConf.yamlData == nil {
				logConf.yamlData = y
				if err := y.Metric.Renew(); err != nil {
					blog.Errorf("Renew metric with label (%+v) failed: %s", *y.Metric, err.Error())
				}
			}
			return
		} else {
			blog.Infof("container %s log config %s changed, from(%s)->to(%s)", c.ID, logConf.confPath, string(logConf.data), string(by))
		}
		blog.Infof("container %s log config %s changed, from(%s)->to(%s)", c.ID, logConf.confPath, string(logConf.data), string(by))
	} else {
		blog.Infof("container %s log config %s will created, and LogConfig(%s)", c.ID, key, string(by))
	}

	newlogConf := &ContainerLogConf{
		confPath: key,
		data:     by,
		yamlData: y,
	}
	f, err := os.Create(newlogConf.confPath)
	if err != nil {
		blog.Errorf("container %s open file %s failed: %s", c.ID, newlogConf.confPath, err.Error())
		return
	}
	defer f.Close()

	_, err = f.Write(by)
	if err != nil {
		blog.Errorf("container %s tempalte execute failed: %s", c.ID, err.Error())
		return
	}
	blog.Infof("produce container %s log config %s success", c.ID, newlogConf.confPath)
	// Set/Update metric
	if logConf == nil || logConf.yamlData == nil {
		err := y.Metric.Set(1)
		if err != nil {
			blog.Errorf("Set metric with label (%+v) with value (1) failed: %s", *logConf.yamlData.Metric, err.Error())
		}
	} else {
		err := logConf.yamlData.Metric.Update(y.Metric)
		if err != nil {
			blog.Errorf("Update metric from label (%+v) to label (%+v) failed: %s", *logConf.yamlData.Metric, *y.Metric, err.Error())
		}
	}
	s.Lock()
	s.logConfs[key] = newlogConf
	s.Unlock()
	return
}

func (s *SidecarController) deleteContainerLogConf(containerID string) {
	key := s.getContainerLogConfKey(containerID)
	logConf, ok := s.logConfs[key]
	if !ok {
		blog.Infof("container %s don't have LogConfig, then ignore", containerID)
		return
	}
	err := os.Remove(logConf.confPath)
	if err != nil {
		blog.Errorf("remove log config %s error %s", logConf.confPath, err.Error())
		return
	}
	if logConf.yamlData != nil {
		logConf.yamlData.Metric.Delete()
	}
	delete(s.logConfs, key)
	blog.Infof("delete container %s log config success", containerID)
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

	logConf := s.getPodLogConfigCrd(container, pod)
	//if logConf==nil, container not match BcsLogConfig
	if logConf == nil {
		return nil, false
	}

	para := types.Local{
		ExtMeta:          make(map[string]string),
		NonstandardPaths: make([]string, 0),
	}
	para.ExtMeta["io_tencent_bcs_cluster"] = logConf.Spec.ClusterId
	para.ExtMeta["io_tencent_bcs_pod"] = pod.Name
	para.ExtMeta["io_tencent_bcs_namespace"] = pod.Namespace
	para.ExtMeta["io_tencent_bcs_server_name"] = pod.OwnerReferences[0].Name
	para.ExtMeta["io_tencent_bcs_type"] = pod.OwnerReferences[0].Kind
	para.ExtMeta["io_tencent_bcs_appid"] = logConf.Spec.AppId
	para.ExtMeta["io_tencent_bcs_projectid"] = pod.Labels["io.tencent.paas.projectid"]
	para.ExtMeta["container_id"] = container.ID
	para.ExtMeta["container_hostname"] = container.Config.Hostname
	para.ToJSON = true
	containerRootPath := s.getContainerRootPath(container)
	var matchedLogConfig bcsv1.ContainerConf
	if len(logConf.Spec.ContainerConfs) > 0 {
		for _, conf := range logConf.Spec.ContainerConfs {
			if conf.ContainerName == name {
				conf.DeepCopyInto(&matchedLogConfig)
				break
			}
		}
	} else {
		matchedLogConfig.StdDataId = logConf.Spec.StdDataId
		matchedLogConfig.NonStdDataId = logConf.Spec.NonStdDataId
		matchedLogConfig.LogPaths = logConf.Spec.LogPaths
		matchedLogConfig.HostPaths = logConf.Spec.HostPaths
		matchedLogConfig.LogTags = logConf.Spec.LogTags
	}
	// generate intermediate config
	para.StdoutDataid = matchedLogConfig.StdDataId
	para.NonstandardDataid = matchedLogConfig.NonStdDataId
	for _, f := range matchedLogConfig.LogPaths {
		if !filepath.IsAbs(f) {
			blog.Errorf("log path specified as \"%s\" is not an absolute path", f)
			continue
		}
		para.NonstandardPaths = append(para.NonstandardPaths, fmt.Sprintf("%s%s", containerRootPath, f))
	}
	for _, f := range matchedLogConfig.HostPaths {
		if !filepath.IsAbs(f) {
			blog.Errorf("host path specified as \"%s\" is not an absolute path", f)
			continue
		}
		para.NonstandardPaths = append(para.NonstandardPaths, f)
	}
	para.LogTags = matchedLogConfig.LogTags
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
		inLocal.DataID = i
		y.Local = append(y.Local, inLocal)
	}
	//if nonstandard Log
	if para.NonstandardDataid != "" && len(para.NonstandardPaths) > 0 {
		inLocal := para
		inLocal.Paths = para.NonstandardPaths
		i, _ := strconv.Atoi(para.NonstandardDataid)
		inLocal.DataID = i
		y.Local = append(y.Local, inLocal)
	}

	// construct log file metric info
	y.Metric = &metric.LogFileInfoType{
		ClusterID:         strings.ToLower(logConf.Spec.ClusterId),
		CRDName:           logConf.GetName(),
		CRDNamespace:      logConf.GetNamespace(),
		HostIP:            pod.Status.HostIP,
		ContainerID:       container.ID,
		PodName:           pod.GetName(),
		PodNamespace:      pod.GetNamespace(),
		WorkloadType:      pod.OwnerReferences[0].Kind,
		WorkloadName:      pod.OwnerReferences[0].Name,
		WorkloadNamespace: pod.GetNamespace(),
	}

	return y, true
}

// getContainerRootPath return the root path of the container
// Usually it begins with /data/bcs/lib/docker/overlay2/{hashid}/merged
// If the container does not use OverlayFS, it will return /proc/{procid}/root
func (s *SidecarController) getContainerRootPath(container *docker.Container) string {
	switch container.Driver {
	// case "overlay2":
	// 	return container.GraphDriver.Data["MergedDir"]
	default:
		// blog.Warnf("Container %s has driver %s not overlay2", container.ID, container.Driver)
		return fmt.Sprintf("/proc/%d/root", container.State.Pid)
	}
}
