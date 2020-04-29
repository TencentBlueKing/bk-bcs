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

package mesos

import (
	"strconv"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	bcsv1 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	listers "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common"
	mapset "github.com/deckarep/golang-set"
	"k8s.io/apimachinery/pkg/labels"
)

// LogConfInject implements MesosInject
type LogConfInject struct {
	BcsLogConfigLister listers.BcsLogConfigLister
}

// NewLogConfInject create LogConfInject object
func NewLogConfInject(bcsLogConfLister listers.BcsLogConfigLister) MesosInject {
	mesosInject := &LogConfInject{
		BcsLogConfigLister: bcsLogConfLister,
	}

	return mesosInject
}

// InjectApplicationContent inject log envs to application
func (logConf *LogConfInject) InjectApplicationContent(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {
	// get all BcsLogConfig
	bcsLogConfs, err := logConf.BcsLogConfigLister.BcsLogConfigs(application.NameSpace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range common.IgnoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(application.ObjectMeta.NameSpace) {
		matchedLogConf := common.FindBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			injected := logConf.injectMesosContainers(application.ObjectMeta.NameSpace, application.ReplicaControllerSpec.Template, matchedLogConf) // nolint
			application.ReplicaControllerSpec.Template = injected
		}
		return application, nil
	}

	// handle business modules log inject
	var injectedContainers []commtypes.Container

	defaultLogConf := common.FindDefaultConfigType(bcsLogConfs)
	matchedLogConf := common.FindMesosMatchedConfigType("application", application.Name, bcsLogConfs)

	if matchedLogConf != nil {
		for _, container := range application.ReplicaControllerSpec.Template.PodSpec.Containers {
			containerMatched := false
			for j, containerConf := range matchedLogConf.Spec.ContainerConfs {
				if container.Name == containerConf.ContainerName {
					containerMatched = true
					injectedContainer := logConf.injectMesosContainer(application.ObjectMeta.NameSpace, container, matchedLogConf, j)
					injectedContainers = append(injectedContainers, injectedContainer)
					break
				}
			}
			if !containerMatched {
				if defaultLogConf != nil {
					injectedContainer := logConf.injectMesosContainer(application.ObjectMeta.NameSpace, container, defaultLogConf, -1)
					injectedContainers = append(injectedContainers, injectedContainer)
				} else {
					injectedContainers = append(injectedContainers, container)
				}
			}
		}
		application.ReplicaControllerSpec.Template.PodSpec.Containers = injectedContainers
		return application, nil
	}

	if defaultLogConf != nil {
		injected := logConf.injectMesosContainers(application.ObjectMeta.NameSpace, application.ReplicaControllerSpec.Template, defaultLogConf) // nolint
		application.ReplicaControllerSpec.Template = injected
	}

	return application, nil
}

// InjectDeployContent inject log envs to Deployment
func (logConf *LogConfInject) InjectDeployContent(deploy *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	// get all BcsLogConfig
	bcsLogConfs, err := logConf.BcsLogConfigLister.BcsLogConfigs(deploy.NameSpace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range common.IgnoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(deploy.ObjectMeta.NameSpace) {
		matchedLogConf := common.FindBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			injected := logConf.injectMesosContainers(deploy.ObjectMeta.NameSpace, deploy.Spec.Template, matchedLogConf)
			deploy.Spec.Template = injected
		}
		return deploy, nil
	}

	// handle business modules log inject
	var injectedContainers []commtypes.Container

	defaultLogConf := common.FindDefaultConfigType(bcsLogConfs)
	matchedLogConf := common.FindMesosMatchedConfigType("deployment", deploy.Name, bcsLogConfs)

	if matchedLogConf != nil {
		for _, container := range deploy.Spec.Template.PodSpec.Containers {
			containerMatched := false
			for j, containerConf := range matchedLogConf.Spec.ContainerConfs {
				if container.Name == containerConf.ContainerName {
					containerMatched = true
					injectedContainer := logConf.injectMesosContainer(deploy.ObjectMeta.NameSpace, container, matchedLogConf, j)
					injectedContainers = append(injectedContainers, injectedContainer)
					break
				}
			}

			if !containerMatched {
				if defaultLogConf != nil {
					injectedContainer := logConf.injectMesosContainer(deploy.ObjectMeta.NameSpace, container, defaultLogConf, -1)
					injectedContainers = append(injectedContainers, injectedContainer)
				} else {
					injectedContainers = append(injectedContainers, container)
				}
			}
		}
		deploy.Spec.Template.PodSpec.Containers = injectedContainers
		return deploy, nil
	}

	if defaultLogConf != nil {
		injected := logConf.injectMesosContainers(deploy.ObjectMeta.NameSpace, deploy.Spec.Template, defaultLogConf) // nolint
		deploy.Spec.Template = injected
	}

	return deploy, nil
}

// injectMesosContainers injects bcs log config to all containers
func (logConf *LogConfInject) injectMesosContainers(namespace string, podTemplate *commtypes.PodTemplateSpec, bcsLogConf *bcsv1.BcsLogConfig) *commtypes.PodTemplateSpec { // nolint

	var injectedContainers []commtypes.Container
	for _, container := range podTemplate.PodSpec.Containers {
		injectedContainer := logConf.injectMesosContainer(namespace, container, bcsLogConf, -1)
		injectedContainers = append(injectedContainers, injectedContainer)
	}

	podTemplate.PodSpec.Containers = injectedContainers
	return podTemplate
}

// injectMesosContainer injects bcs log config to an container
func (logConf *LogConfInject) injectMesosContainer(namespace string, container commtypes.Container, bcsLogConf *bcsv1.BcsLogConfig, index int) commtypes.Container { // nolint
	var envs []commtypes.EnvVar

	clusterIdEnv := commtypes.EnvVar{
		Name:  common.ClusterIdEnvKey,
		Value: bcsLogConf.Spec.ClusterId,
	}
	envs = append(envs, clusterIdEnv)

	namespaceEnv := commtypes.EnvVar{
		Name:  common.NamespaceEnvKey,
		Value: namespace,
	}
	envs = append(envs, namespaceEnv)

	appIdEnv := commtypes.EnvVar{
		Name:  common.AppIdEnvKey,
		Value: bcsLogConf.Spec.AppId,
	}
	envs = append(envs, appIdEnv)

	if index >= 0 {
		containerConf := bcsLogConf.Spec.ContainerConfs[index]

		if containerConf.StdDataId != "" {
			stdDataIdEnv := commtypes.EnvVar{
				Name:  common.StdDataIdEnvKey,
				Value: containerConf.StdDataId,
			}
			envs = append(envs, stdDataIdEnv)
		}

		if containerConf.NonStdDataId != "" {
			nonStdDataIdEnv := commtypes.EnvVar{
				Name:  common.NonStdDataIdEnvKey,
				Value: containerConf.NonStdDataId,
			}
			envs = append(envs, nonStdDataIdEnv)
		}

		stdoutEnv := commtypes.EnvVar{
			Name:  common.StdoutEnvKey,
			Value: strconv.FormatBool(containerConf.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if len(containerConf.LogPaths) > 0 {
			logPathEnv := commtypes.EnvVar{
				Name:  common.LogPathEnvKey,
				Value: strings.Join(containerConf.LogPaths, ","),
			}
			envs = append(envs, logPathEnv)
		}

		if len(containerConf.LogTags) > 0 {
			var tags []string
			for k, v := range containerConf.LogTags {
				tag := k + ":" + v
				tags = append(tags, tag)
			}

			logTagEnv := commtypes.EnvVar{
				Name:  common.LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	} else {
		stdoutEnv := commtypes.EnvVar{
			Name:  common.StdoutEnvKey,
			Value: strconv.FormatBool(bcsLogConf.Spec.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if bcsLogConf.Spec.StdDataId != "" {
			stdDataIdEnv := commtypes.EnvVar{
				Name:  common.StdDataIdEnvKey,
				Value: bcsLogConf.Spec.StdDataId,
			}
			envs = append(envs, stdDataIdEnv)
		}

		if bcsLogConf.Spec.NonStdDataId != "" {
			nonStdDataIdEnv := commtypes.EnvVar{
				Name:  common.NonStdDataIdEnvKey,
				Value: bcsLogConf.Spec.NonStdDataId,
			}
			envs = append(envs, nonStdDataIdEnv)
		}

		if len(bcsLogConf.Spec.LogPaths) > 0 {
			logPathEnv := commtypes.EnvVar{
				Name:  common.LogPathEnvKey,
				Value: strings.Join(bcsLogConf.Spec.LogPaths, ","),
			}
			envs = append(envs, logPathEnv)
		}

		if len(bcsLogConf.Spec.LogTags) > 0 {
			var tags []string
			for k, v := range bcsLogConf.Spec.LogTags {
				tag := k + ":" + v
				tags = append(tags, tag)
			}

			logTagEnv := commtypes.EnvVar{
				Name:  common.LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	}

	container.Env = append(container.Env, envs...)
	//blog.Infof("%v", container.Env)
	return container
}
