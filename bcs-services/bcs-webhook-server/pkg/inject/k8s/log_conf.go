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
	"fmt"
	"strconv"

	"bk-bcs/bcs-common/common/blog"
	bcsv2 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v2"
	listers "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v2"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common"
	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	bcsLogConfs, err := logConf.BcsLogConfigLister.List(labels.Everything())
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
				patchedContainer := logConf.injectK8sContainer(pod.Namespace, &container, matchedLogConf)
				patch = append(patch, replaceContainer(i, *patchedContainer))
			}
		}
		return patch, nil
	}

	for i, container := range pod.Spec.Containers {
		matchedLogConf := common.FindMatchedConfigType(container.Name, bcsLogConfs)
		if matchedLogConf != nil {
			patchedContainer := logConf.injectK8sContainer(pod.Namespace, &container, matchedLogConf)
			patch = append(patch, replaceContainer(i, *patchedContainer))
		}
	}

	return patch, nil
}

func (logConf *LogConfInject) injectK8sContainer(namespace string, container *corev1.Container, bcsLogConf *bcsv2.BcsLogConfig) *corev1.Container {

	patchedContainer := container.DeepCopy()

	var envs []corev1.EnvVar
	dataIdEnv := corev1.EnvVar{
		Name:  common.DataIdEnvKey,
		Value: bcsLogConf.Spec.DataId,
	}
	envs = append(envs, dataIdEnv)

	appIdEnv := corev1.EnvVar{
		Name:  common.AppIdEnvKey,
		Value: bcsLogConf.Spec.AppId,
	}
	envs = append(envs, appIdEnv)

	stdoutEnv := corev1.EnvVar{
		Name:  common.StdoutEnvKey,
		Value: strconv.FormatBool(bcsLogConf.Spec.Stdout),
	}
	envs = append(envs, stdoutEnv)

	logPathEnv := corev1.EnvVar{
		Name:  common.LogPathEnvKey,
		Value: bcsLogConf.Spec.LogPath,
	}
	envs = append(envs, logPathEnv)

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

	patchedContainer.Env = envs

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
