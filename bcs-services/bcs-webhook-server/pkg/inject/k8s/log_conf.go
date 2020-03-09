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

package k8s

import (
	"bk-bcs/bcs-common/common/blog"
	bcsv1 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	listers "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"strconv"
	"strings"
)

// LogConfInject implements K8sInject
type LogConfInject struct {
	BcsLogConfigLister listers.BcsLogConfigLister
}

// NewLogConfInject create LogConfInject object
func NewLogConfInject(bcsLogConfLister listers.BcsLogConfigLister) K8sInject {
	k8sInject := &LogConfInject{
		BcsLogConfigLister: bcsLogConfLister,
	}

	return k8sInject
}

// InjectContent inject log envs to pod
func (logConf *LogConfInject) InjectContent(pod *corev1.Pod) ([]PatchOperation, error) {

	var patch []PatchOperation

	bcsLogConfs, err := logConf.BcsLogConfigLister.BcsLogConfigs(pod.ObjectMeta.Namespace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range common.IgnoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(pod.ObjectMeta.Namespace) {
		matchedLogConf := common.FindBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			for i, container := range pod.Spec.Containers {
				patchedContainer := logConf.injectK8sContainer(pod.Namespace, &container, matchedLogConf, -1)
				patch = append(patch, replaceContainer(i, *patchedContainer))
			}
		}
		return patch, nil
	}

	//handle business modules' log inject
	defaultLogConf := common.FindDefaultConfigType(bcsLogConfs)
	matchedLogConf := common.FindK8sMatchedConfigType(pod, bcsLogConfs)
	if matchedLogConf != nil {
		for i, container := range pod.Spec.Containers {
			containerMatched := false
			for j, containerConf := range matchedLogConf.Spec.ContainerConfs {
				if container.Name == containerConf.ContainerName {
					containerMatched = true
					patchedContainer := logConf.injectK8sContainer(pod.Namespace, &container, matchedLogConf, j)
					patch = append(patch, replaceContainer(i, *patchedContainer))
					break
				}
			}
			if !containerMatched {
				if defaultLogConf != nil {
					patchedContainer := logConf.injectK8sContainer(pod.Namespace, &container, defaultLogConf, -1)
					patch = append(patch, replaceContainer(i, *patchedContainer))
				}
			}
		}
	} else {
		if defaultLogConf != nil {
			for i, container := range pod.Spec.Containers {
				patchedContainer := logConf.injectK8sContainer(pod.Namespace, &container, defaultLogConf, -1)
				patch = append(patch, replaceContainer(i, *patchedContainer))
			}
		}
	}

	return patch, nil
}

func (logConf *LogConfInject) injectK8sContainer(namespace string, container *corev1.Container, bcsLogConf *bcsv1.BcsLogConfig, index int) *corev1.Container { // nolint

	patchedContainer := container.DeepCopy()

	var envs []corev1.EnvVar

	clusterIdEnv := corev1.EnvVar{
		Name:  common.ClusterIdEnvKey,
		Value: bcsLogConf.Spec.ClusterId,
	}
	envs = append(envs, clusterIdEnv)

	namespaceEnv := corev1.EnvVar{
		Name:  common.NamespaceEnvKey,
		Value: namespace,
	}
	envs = append(envs, namespaceEnv)

	appIdEnv := corev1.EnvVar{
		Name:  common.AppIdEnvKey,
		Value: bcsLogConf.Spec.AppId,
	}
	envs = append(envs, appIdEnv)

	if index >= 0 {
		containerConf := bcsLogConf.Spec.ContainerConfs[index]

		if containerConf.StdDataId != "" {
			stdDataIdEnv := corev1.EnvVar{
				Name:  common.StdDataIdEnvKey,
				Value: containerConf.StdDataId,
			}
			envs = append(envs, stdDataIdEnv)
		}

		if containerConf.NonStdDataId != "" {
			nonStdDataIdEnv := corev1.EnvVar{
				Name:  common.NonStdDataIdEnvKey,
				Value: containerConf.NonStdDataId,
			}
			envs = append(envs, nonStdDataIdEnv)
		}

		stdoutEnv := corev1.EnvVar{
			Name:  common.StdoutEnvKey,
			Value: strconv.FormatBool(containerConf.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if len(containerConf.LogPaths) > 0 {
			logPathEnv := corev1.EnvVar{
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

			logTagEnv := corev1.EnvVar{
				Name:  common.LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	} else {
		stdoutEnv := corev1.EnvVar{
			Name:  common.StdoutEnvKey,
			Value: strconv.FormatBool(bcsLogConf.Spec.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if bcsLogConf.Spec.StdDataId != "" {
			stdDataIdEnv := corev1.EnvVar{
				Name:  common.StdDataIdEnvKey,
				Value: bcsLogConf.Spec.StdDataId,
			}
			envs = append(envs, stdDataIdEnv)
		}

		if bcsLogConf.Spec.NonStdDataId != "" {
			nonStdDataIdEnv := corev1.EnvVar{
				Name:  common.NonStdDataIdEnvKey,
				Value: bcsLogConf.Spec.NonStdDataId,
			}
			envs = append(envs, nonStdDataIdEnv)
		}

		if len(bcsLogConf.Spec.LogPaths) > 0 {
			logPathEnv := corev1.EnvVar{
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

			logTagEnv := corev1.EnvVar{
				Name:  common.LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	}

	patchedContainer.Env = append(patchedContainer.Env, envs...)

	return patchedContainer
}

func replaceContainer(index int, patchedContainer corev1.Container) (patch PatchOperation) {
	patch = PatchOperation{
		Op:    "replace",
		Path:  fmt.Sprintf("/spec/containers/%v", index),
		Value: patchedContainer,
	}
	return patch
}
