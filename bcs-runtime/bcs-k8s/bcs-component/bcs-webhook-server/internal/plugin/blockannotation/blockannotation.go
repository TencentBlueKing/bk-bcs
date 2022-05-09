/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blockannotation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

const (
	pluginName = "annoblocker"
)

func init() {
	p := &AnnotationBlocker{
		nsMap: make(map[string]struct{}),
	}
	pluginmanager.Register(pluginName, p)
}

// AnnotationBlocker blocker for certain annotation
type AnnotationBlocker struct {
	conf     *Config
	nsMap    map[string]struct{}
	blockers []*BlockUnit
}

// AnnotationKey returns key of the annoblocker plugin for hook server to identify
func (ab *AnnotationBlocker) AnnotationKey() string {
	return ab.conf.AnnotationKey
}

// Init init plugin
func (ab *AnnotationBlocker) Init(configFilePath string) error {
	fileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
	}
	conf := &Config{}
	if err := json.Unmarshal(fileBytes, conf); err != nil {
		return fmt.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
	}
	ab.conf = conf

	var blockers []*BlockUnit
	for _, condition := range conf.Conditions {
		blockers = append(blockers, NewBlockUnit(
			condition.Reference,
			condition.Operator,
			condition.FailPolicy))
	}
	ab.blockers = blockers
	for _, ns := range conf.Namespaces {
		ab.nsMap[ns] = struct{}{}
	}
	return nil
}

// Handle handle webhook request
func (ab *AnnotationBlocker) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
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
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	// Deal with potential empty fileds, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}
	value, isHooked := ab.hookRequired(pod)
	if !isHooked {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
		}
	}
	for _, blocker := range ab.blockers {
		isBlocked := blocker.IsBlock(value)
		if isBlocked {
			blog.Errorf("pod %s/%s is blocked by %v", pod.GetName(), pod.GetNamespace(), blocker)
			metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
			return pluginutil.ToAdmissionResponse(
				fmt.Errorf("pod %s/%s is blocked by %v", pod.GetName(), pod.GetNamespace(), blocker))
		}
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func (ab *AnnotationBlocker) hookRequired(pod *corev1.Pod) (string, bool) {
	if _, ok := ab.nsMap[pod.GetNamespace()]; !ok {
		blog.V(3).Infof("Pod %s/%s has no expected ns %v", pod.Name, pod.Namespace, ab.conf.Namespaces)
		return "", false
	}
	value, ok := pod.Annotations[ab.conf.AnnotationKey]
	if !ok {
		blog.V(3).Infof("Pod %s/%s has no expected annoation key", pod.Name, pod.Namespace)
		return "", false
	}
	return value, true
}

// Close close plugin
func (ab *AnnotationBlocker) Close() error {
	return nil
}
