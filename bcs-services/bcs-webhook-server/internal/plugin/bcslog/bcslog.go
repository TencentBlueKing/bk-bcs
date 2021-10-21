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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bk-bcs/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/informers/externalversions"
	listers "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/listers/bk-bcs/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/types"
)

// Hooker webhook for bcslog
type Hooker struct {
	stopCh             chan struct{}
	opt                *Options
	bcsLogConfigLister listers.BcsLogConfigLister
}

// Init implements webhook plugin interface
func (h *Hooker) Init(configFilePath string) error {
	fileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("read bcs log config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("read bcs log config file %s failed, err %s", configFilePath, err.Error())
	}
	h.opt = &Options{}
	if err = json.Unmarshal(fileBytes, h.opt); err != nil {
		blog.Errorf("decode bcs log config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("decode bcs log config file %s failed, err %s", configFilePath, err.Error())
	}
	if err = h.opt.Validate(); err != nil {
		return err
	}
	h.stopCh = make(chan struct{})

	return nil
}

func (h *Hooker) initKubeClient() error {
	cfg, err := clientcmd.BuildConfigFromFlags(h.opt.KubeMaster, h.opt.Kubeconfig)
	if err != nil {
		return fmt.Errorf("building kubeconfig failed, err %s", err.Error())
	}
	// init extension kubeclient to create bcs log crd
	extensionClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("buildling extension clientset failed, err %s", err.Error())
	}
	bcslogCrdCreated, err := h.createBcsLogConfig(extensionClient)
	if err != nil {
		return fmt.Errorf("create bcs log crd failed, err %s", err.Error())
	}
	blog.Infof("created BcsLogConfig crd: %t", bcslogCrdCreated)

	// init client for list bcs log config
	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("build internal clientset failed, err %s", err.Error())
	}
	factory := informers.NewSharedInformerFactory(clientset, 0)
	bcsLogConfigInformer := factory.Bkbcs().V1().BcsLogConfigs()
	h.bcsLogConfigLister = bcsLogConfigInformer.Lister()

	go factory.Start(h.stopCh)

	blog.Infof("Waiting for BcsLogConfig inormer caches to sync")
	blog.Infof("sleep 1 seconds to wait for BcsLogConfig crd to be ready")
	time.Sleep(1 * time.Second)
	if ok := clientGoCache.WaitForCacheSync(h.stopCh, bcsLogConfigInformer.Informer().HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	return nil
}

// create crd of BcsLogConf
func (h *Hooker) createBcsLogConfig(clientset apiextensionsclient.Interface) (bool, error) {
	bcsLogConfigPlural := "bcslogconfigs"

	bcsLogConfigFullName := "bcslogconfigs" + "." + bcsv1.SchemeGroupVersion.Group

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsLogConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv1.SchemeGroupVersion.Group,   // BcsLogConfigsGroup,
			Version: bcsv1.SchemeGroupVersion.Version, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   bcsLogConfigPlural,
				Kind:     reflect.TypeOf(bcsv1.BcsLogConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv1.BcsLogConfigList{}).Name(),
			},
		},
	}

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("crd is already exists: %s", err)
			return false, nil
		}
		blog.Errorf("create crd failed: %s", err)
		return false, err
	}
	return true, nil
}

// AnnotationKey implements webhook plugin interface
func (h *Hooker) AnnotationKey() string {
	return ""
}

// Handle implements webhook plugin interface
func (h *Hooker) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	started := time.Now()
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(BcsLogPluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}

	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}

	// do inject
	patches, err := h.createPatch(pod)
	if err != nil {
		blog.Errorf("create path failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(BcsLogPluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("encoding patches failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(BcsLogPluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	reviewResponse := v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	metrics.ReportBcsWebhookServerPluginLantency(BcsLogPluginName, metrics.StatusSuccess, started)

	return &reviewResponse
}

func (h *Hooker) createPatch(pod *corev1.Pod) ([]types.PatchOperation, error) {
	var patch []types.PatchOperation

	bcsLogConfs, err := h.bcsLogConfigLister.BcsLogConfigs(pod.ObjectMeta.Namespace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range IgnoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(pod.ObjectMeta.Namespace) {
		matchedLogConf := FindBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			for i, container := range pod.Spec.Containers {
				patchedContainer := h.injectK8sContainer(pod.Namespace, &container, matchedLogConf, -1)
				patch = append(patch, replaceContainer(i, *patchedContainer))
			}
		}
		return patch, nil
	}

	//handle business modules' log inject
	defaultLogConf := FindDefaultConfigType(bcsLogConfs)
	matchedLogConf := FindK8sMatchedConfigType(pod, bcsLogConfs)
	if matchedLogConf != nil {
		for i, container := range pod.Spec.Containers {
			containerMatched := false
			for j, containerConf := range matchedLogConf.Spec.ContainerConfs {
				if container.Name == containerConf.ContainerName {
					containerMatched = true
					patchedContainer := h.injectK8sContainer(pod.Namespace, &container, matchedLogConf, j)
					patch = append(patch, replaceContainer(i, *patchedContainer))
					break
				}
			}
			if !containerMatched {
				if defaultLogConf != nil {
					patchedContainer := h.injectK8sContainer(pod.Namespace, &container, defaultLogConf, -1)
					patch = append(patch, replaceContainer(i, *patchedContainer))
				}
			}
		}
	} else {
		if defaultLogConf != nil {
			for i, container := range pod.Spec.Containers {
				patchedContainer := h.injectK8sContainer(pod.Namespace, &container, defaultLogConf, -1)
				patch = append(patch, replaceContainer(i, *patchedContainer))
			}
		}
	}

	return patch, nil
}

func (h *Hooker) injectK8sContainer(
	namespace string, container *corev1.Container,
	bcsLogConf *bcsv1.BcsLogConfig, index int) *corev1.Container {

	patchedContainer := container.DeepCopy()
	var envs []corev1.EnvVar
	clusterIDEnv := corev1.EnvVar{
		Name:  ClusterIDEnvKey,
		Value: bcsLogConf.Spec.ClusterId,
	}
	envs = append(envs, clusterIDEnv)

	namespaceEnv := corev1.EnvVar{
		Name:  NamespaceEnvKey,
		Value: namespace,
	}
	envs = append(envs, namespaceEnv)

	appIDEnv := corev1.EnvVar{
		Name:  AppIDEnvKey,
		Value: bcsLogConf.Spec.AppId,
	}
	envs = append(envs, appIDEnv)

	if index >= 0 {
		containerConf := bcsLogConf.Spec.ContainerConfs[index]

		if containerConf.StdDataId != "" {
			stdDataIDEnv := corev1.EnvVar{
				Name:  StdDataIDEnvKey,
				Value: containerConf.StdDataId,
			}
			envs = append(envs, stdDataIDEnv)
		}

		if containerConf.NonStdDataId != "" {
			nonStdDataIDEnv := corev1.EnvVar{
				Name:  NonStdDataIDEnvKey,
				Value: containerConf.NonStdDataId,
			}
			envs = append(envs, nonStdDataIDEnv)
		}

		stdoutEnv := corev1.EnvVar{
			Name:  StdoutEnvKey,
			Value: strconv.FormatBool(containerConf.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if len(containerConf.LogPaths) > 0 {
			logPathEnv := corev1.EnvVar{
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

			logTagEnv := corev1.EnvVar{
				Name:  LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	} else {
		stdoutEnv := corev1.EnvVar{
			Name:  StdoutEnvKey,
			Value: strconv.FormatBool(bcsLogConf.Spec.Stdout),
		}
		envs = append(envs, stdoutEnv)

		if bcsLogConf.Spec.StdDataId != "" {
			stdDataIDEnv := corev1.EnvVar{
				Name:  StdDataIDEnvKey,
				Value: bcsLogConf.Spec.StdDataId,
			}
			envs = append(envs, stdDataIDEnv)
		}

		if bcsLogConf.Spec.NonStdDataId != "" {
			nonStdDataIDEnv := corev1.EnvVar{
				Name:  NonStdDataIDEnvKey,
				Value: bcsLogConf.Spec.NonStdDataId,
			}
			envs = append(envs, nonStdDataIDEnv)
		}

		if len(bcsLogConf.Spec.LogPaths) > 0 {
			logPathEnv := corev1.EnvVar{
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

			logTagEnv := corev1.EnvVar{
				Name:  LogTagEnvKey,
				Value: strings.Join(tags, ","),
			}
			envs = append(envs, logTagEnv)
		}
	}

	patchedContainer.Env = append(patchedContainer.Env, envs...)

	return patchedContainer
}

func replaceContainer(index int, patchedContainer corev1.Container) types.PatchOperation {
	return types.PatchOperation{
		Op:    "replace",
		Path:  fmt.Sprintf("/spec/containers/%v", index),
		Value: patchedContainer,
	}
}

// Close implements webhook plugin interface
func (h *Hooker) Close() error {
	h.stopCh <- struct{}{}
	return nil
}
