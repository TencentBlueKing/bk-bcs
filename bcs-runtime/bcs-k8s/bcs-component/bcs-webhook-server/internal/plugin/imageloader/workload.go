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

package imageloader

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/apis/tkex/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var (
	workloads = make(map[string]Workload)
)

// Workload xxx
type Workload interface {
	// Name return name of the workload for imageloader handler to dispatch.
	Name() string
	// Init inits the workload(start informer or sth.).
	Init(*imageLoader) error
	// LoadImageBeforeUpdate is called when the corresponding workload instance being updated.
	LoadImageBeforeUpdate(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse
	// JobDoneHook is called when image-load job is done(either success or fail).
	JobDoneHook(namespace, name string, event *corev1.Event) error
	// WaitForCacheSync waits the cache informer of the workload to be synced.
	WaitForCacheSync(chan struct{}) bool
}

// InitWorkloads xxx
// Init inits all workloads.
func InitWorkloads(i *imageLoader) (map[string]Workload, error) {
	// DOTO add other workloads
	if strings.Contains(i.config.Workload, tkexv1alpha1.KindGameDeployment) {
		// add bcsgd
		bcsgd := &bcsgdWorkload{}
		err := bcsgd.Init(i)
		if err != nil {
			return nil, fmt.Errorf("init bcsgd workload failed: %v", err)
		}
		workloads[bcsgd.Name()] = bcsgd
	}

	if strings.Contains(i.config.Workload, tkexv1alpha1.KindGameStatefulSet) {
		// add bcsgs
		bcsgs := &bcsgsWorkload{}
		err := bcsgs.Init(i)
		if err != nil {
			return nil, fmt.Errorf("init bcsgs workload failed: %v", err)
		}
		workloads[bcsgs.Name()] = bcsgs
	}

	return workloads, nil
}

func workloadsWaitForCacheSync(stopCh chan struct{}) bool {
	for n, w := range workloads {
		if !w.WaitForCacheSync(stopCh) {
			blog.Errorf("workload %s cache sync failed", n)
			return false
		}
	}
	return true
}
