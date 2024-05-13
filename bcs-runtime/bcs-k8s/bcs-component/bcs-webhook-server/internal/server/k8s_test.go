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

package server

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	_ "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/fake"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/options"
)

func TestK8SHook(t *testing.T) {
	opt := &options.ServerOption{
		EngineType: options.EngineTypeKubernetes,
		Plugins:    "fake,fake",
	}
	pm := pluginmanager.NewManager(opt.EngineType, opt.PluginDir)
	pluginNames := strings.Split(opt.Plugins, ",")
	if err := pm.InitPlugins(pluginNames); err != nil {
		t.Error(err)
	}
	server := &WebhookServer{
		Opt:        opt,
		EngineType: opt.EngineType,
		PluginMgr:  pm,
	}

	testCases := []struct {
		title    string
		inputPod *corev1.Pod
		allowed  bool
		outPod   *corev1.Pod
	}{
		{
			title: "test1",
			inputPod: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
			},
			allowed: true,
			outPod: &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "fakeimgae",
						},
						{
							Image: "fakeimgae",
						},
					},
				},
			},
		},
	}

	for index, testCase := range testCases {
		t.Logf("test %d, %s", index, testCase.title)
		podBytes, err := json.Marshal(testCase.inputPod)
		if err != nil {
			t.Error(err)
			return
		}
		resp := server.doK8sHook(v1beta1.AdmissionReview{
			Request: &v1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "Pod",
				},
				Operation: v1beta1.Create,
				Object: runtime.RawExtension{
					Raw: podBytes,
				},
			},
		})
		if testCase.allowed != resp.Allowed {
			t.Errorf("expect allowed %t but get %t", testCase.allowed, resp.Allowed)
			return
		}
		patches, err := jsonpatch.DecodePatch(resp.Patch)
		if err != nil {
			t.Error(err)
			return
		}
		modified, err := patches.Apply(podBytes)
		if err != nil {
			t.Error(err)
			return
		}
		tmpPod := &corev1.Pod{}
		err = json.Unmarshal(modified, tmpPod)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(tmpPod.Spec, testCase.outPod.Spec) {
			t.Errorf("expect %+v, but get %+v", testCase.outPod, tmpPod)
			return
		}
	}
}
