/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcslog

import (
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	mapset "github.com/deckarep/golang-set"
	"k8s.io/apimachinery/pkg/labels"
)

// InjectApplicationContent inject log envs to application
func (h *Hooker) InjectApplicationContent(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {
	// get all BcsLogConfig
	bcsLogConfs, err := h.bcsLogConfigLister.BcsLogConfigs(application.NameSpace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range IgnoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(application.ObjectMeta.NameSpace) {
		matchedLogConf := FindBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			injected := h.injectMesosContainers(
				application.ObjectMeta.NameSpace,
				application.ReplicaControllerSpec.Template,
				matchedLogConf)
			application.ReplicaControllerSpec.Template = injected
		}
		return application, nil
	}

	// handle business modules log inject
	var injectedContainers []commtypes.Container

	defaultLogConf := FindDefaultConfigType(bcsLogConfs)
	matchedLogConf := FindMesosMatchedConfigType("application", application.Name, bcsLogConfs)

	if matchedLogConf != nil {
		for _, container := range application.ReplicaControllerSpec.Template.PodSpec.Containers {
			containerMatched := false
			for j, containerConf := range matchedLogConf.Spec.ContainerConfs {
				if container.Name == containerConf.ContainerName {
					containerMatched = true
					injectedContainer := h.injectMesosContainer(
						application.ObjectMeta.NameSpace,
						container, matchedLogConf, j)
					injectedContainers = append(injectedContainers, injectedContainer)
					break
				}
			}
			if !containerMatched {
				if defaultLogConf != nil {
					injectedContainer := h.injectMesosContainer(
						application.ObjectMeta.NameSpace,
						container, defaultLogConf, -1)
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
		injected := h.injectMesosContainers(
			application.ObjectMeta.NameSpace,
			application.ReplicaControllerSpec.Template,
			defaultLogConf)
		application.ReplicaControllerSpec.Template = injected
	}

	return application, nil
}

// InjectDeployContent inject log envs to Deployment
func (h *Hooker) InjectDeployContent(deploy *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	// get all BcsLogConfig
	bcsLogConfs, err := h.bcsLogConfigLister.BcsLogConfigs(deploy.NameSpace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range IgnoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(deploy.ObjectMeta.NameSpace) {
		matchedLogConf := FindBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			injected := h.injectMesosContainers(deploy.ObjectMeta.NameSpace, deploy.Spec.Template, matchedLogConf)
			deploy.Spec.Template = injected
		}
		return deploy, nil
	}

	// handle business modules log inject
	var injectedContainers []commtypes.Container

	defaultLogConf := FindDefaultConfigType(bcsLogConfs)
	matchedLogConf := FindMesosMatchedConfigType("deployment", deploy.Name, bcsLogConfs)

	if matchedLogConf != nil {
		for _, container := range deploy.Spec.Template.PodSpec.Containers {
			containerMatched := false
			for j, containerConf := range matchedLogConf.Spec.ContainerConfs {
				if container.Name == containerConf.ContainerName {
					containerMatched = true
					injectedContainer := h.injectMesosContainer(
						deploy.ObjectMeta.NameSpace, container, matchedLogConf, j)
					injectedContainers = append(injectedContainers, injectedContainer)
					break
				}
			}

			if !containerMatched {
				if defaultLogConf != nil {
					injectedContainer := h.injectMesosContainer(
						deploy.ObjectMeta.NameSpace, container, defaultLogConf, -1)
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
		injected := h.injectMesosContainers(
			deploy.ObjectMeta.NameSpace, deploy.Spec.Template, defaultLogConf) // nolint
		deploy.Spec.Template = injected
	}

	return deploy, nil
}

// injectMesosContainers injects bcs log config to all containers
func (h *Hooker) injectMesosContainers(namespace string, podTemplate *commtypes.PodTemplateSpec, bcsLogConf *bcsv1.BcsLogConfig) *commtypes.PodTemplateSpec { // nolint

	var injectedContainers []commtypes.Container
	for _, container := range podTemplate.PodSpec.Containers {
		injectedContainer := h.injectMesosContainer(namespace, container, bcsLogConf, -1)
		injectedContainers = append(injectedContainers, injectedContainer)
	}

	podTemplate.PodSpec.Containers = injectedContainers
	return podTemplate
}

// injectMesosContainer injects bcs log config to an container
func (h *Hooker) injectMesosContainer(namespace string, container commtypes.Container, bcsLogConf *bcsv1.BcsLogConfig, index int) commtypes.Container { // nolint
	var envs []commtypes.EnvVar

	clusterIDEnv := commtypes.EnvVar{
		Name:  ClusterIDEnvKey,
		Value: bcsLogConf.Spec.ClusterId,
	}
	envs = append(envs, clusterIDEnv)

	namespaceEnv := commtypes.EnvVar{
		Name:  NamespaceEnvKey,
		Value: namespace,
	}
	envs = append(envs, namespaceEnv)

	appIDEnv := commtypes.EnvVar{
		Name:  AppIDEnvKey,
		Value: bcsLogConf.Spec.AppId,
	}
	envs = append(envs, appIDEnv)

	if index >= 0 {
		containerConf := bcsLogConf.Spec.ContainerConfs[index]

		if containerConf.StdDataId != "" {
			stdDataIDEnv := commtypes.EnvVar{
				Name:  StdDataIDEnvKey,
				Value: containerConf.StdDataId,
			}
			envs = append(envs, stdDataIDEnv)
		}

		if containerConf.NonStdDataId != "" {
			nonStdDataIDEnv := commtypes.EnvVar{
				Name:  NonStdDataIDEnvKey,
				Value: containerConf.NonStdDataId,
			}
			envs = append(envs, nonStdDataIDEnv)
		}

		stdoutEnv := commtypes.EnvVar{
			Name:  StdoutEnvKey,
			Value: strconv.FormatBool(containerConf.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if len(containerConf.LogPaths) > 0 {
			logPathEnv := commtypes.EnvVar{
				Name:  LogPathEnvKey,
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
				Name:  LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	} else {
		stdoutEnv := commtypes.EnvVar{
			Name:  StdoutEnvKey,
			Value: strconv.FormatBool(bcsLogConf.Spec.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if bcsLogConf.Spec.StdDataId != "" {
			stdDataIDEnv := commtypes.EnvVar{
				Name:  StdDataIDEnvKey,
				Value: bcsLogConf.Spec.StdDataId,
			}
			envs = append(envs, stdDataIDEnv)
		}

		if bcsLogConf.Spec.NonStdDataId != "" {
			nonStdDataIDEnv := commtypes.EnvVar{
				Name:  NonStdDataIDEnvKey,
				Value: bcsLogConf.Spec.NonStdDataId,
			}
			envs = append(envs, nonStdDataIDEnv)
		}

		if len(bcsLogConf.Spec.LogPaths) > 0 {
			logPathEnv := commtypes.EnvVar{
				Name:  LogPathEnvKey,
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
				Name:  LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	}

	container.Env = append(container.Env, envs...)
	return container
}
