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
 */

// Package imageacceleration xxx
package imageacceleration

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/imageacceleration/cachemanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
)

// Handler defines the handler of image acceleration
type Handler struct {
	client       *kubernetes.Clientset
	cacheManager cachemanager.CacheInterface
}

// Init the webhook plugin, plugin should read config from file configFilePath
func (h *Handler) Init(configFilePath string) error {
	client, err := initK8sClient()
	if err != nil {
		return errors.Wrapf(err, "init k8s client for image accerleration failed")
	}
	h.client = client
	cacheManager := cachemanager.NewCacheManager(client)
	if err := cacheManager.Init(); err != nil {
		return errors.Wrapf(err, "init cache manager failed")
	}
	h.cacheManager = cacheManager
	return nil
}

// AnnotationKey get annotation key for webhook plugin
func (h *Handler) AnnotationKey() string {
	return ""
}

// Handle do hook function
func (h *Handler) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if req.Operation != v1beta1.Create {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	started := time.Now()

	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("image acceleration unmarshal '%s' to pod object failed: %s",
			string(req.Object.Raw), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}
	if !h.injectRequired(pod) {
		return pluginutil.ToAdmissionAllowedResponse()
	}

	patches, err := h.injectToPod(pod)
	if err != nil {
		blog.Errorf("image acceleration inject to pod '%s/%s' failed: %s", pod.Namespace, pod.Name, err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionAllowedResponse()
	}
	if len(patches) == 0 {
		return pluginutil.ToAdmissionAllowedResponse()
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("image acceleration encoding patches failed: %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionAllowedResponse()
	}

	metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusSuccess, started)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func (h *Handler) injectRequired(pod *corev1.Pod) bool {
	cm, err := h.cacheManager.GetConfigMap(pod.Namespace, configMapName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			blog.Errorf("image acceleration get configmap '%s/%s' failed: %s",
				pod.Namespace, configMapName, err.Error())
		} else {
			blog.Warnf("image acceleration configmap '%s/%s' not exist", pod.Namespace, configMapName)
		}
		return false
	}
	v := cm.Data[configMapKeyEnabled]
	return v == "true"
}

func (h *Handler) injectToPod(pod *corev1.Pod) ([]types.PatchOperation, error) {
	cm, err := h.cacheManager.GetConfigMap(pod.Namespace, configMapName)
	if err != nil {
		return nil, errors.Wrapf(err, "image acceleration get configmap failed")
	}
	configmapMapping := h.parseConfigMapping(cm)
	if len(configmapMapping) == 0 {
		return nil, nil
	}
	h.handleImagePullSecret(configmapMapping, pod)

	results := make([]types.PatchOperation, 0)
	for i := range pod.Spec.Containers {
		c := pod.Spec.Containers[i]
		arr := strings.Split(c.Image, "/")
		registry := arr[0]
		v, ok := configmapMapping[registry]
		if !ok || v == "" {
			continue
		}
		arr[0] = v
		results = append(results, types.PatchOperation{
			Path:  fmt.Sprintf(containerImageKey, i),
			Op:    patchOpReplace,
			Value: strings.Join(arr, "/"),
		})
	}
	return results, nil
}

func (h *Handler) parseConfigMapping(cm *corev1.ConfigMap) map[string]string {
	v, ok := cm.Data[configMapKeyMapping]
	if !ok || v == "" {
		return nil
	}
	configmapMapping := make(map[string]string)
	mapping := strings.Split(v, "\n")
	for _, item := range mapping {
		if item = strings.TrimSpace(item); item == "" {
			continue
		}
		kv := strings.Split(item, "=")
		if len(kv) != 2 {
			blog.Warnf("image acceleration parse 'mapping' key for '%s' value '%s' length not 2",
				cm.Namespace, configMapName, item)
			continue
		}
		configmapMapping[kv[0]] = kv[1]
	}
	return configmapMapping
}

// Close called when webhook plugin exit
func (h *Handler) Close() error {
	h.cacheManager.Close()
	return nil
}

// InjectApplicationContent implements mesos plugin interface
func (h *Handler) InjectApplicationContent(application *commtypes.ReplicaController) (
	*commtypes.ReplicaController, error) {
	return nil, nil
}

// InjectDeployContent implements mesos plugin interface
func (h *Handler) InjectDeployContent(deploy *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	return nil, nil
}
