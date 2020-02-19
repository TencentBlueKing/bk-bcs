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

package statefulsetplus

import (
	"fmt"

	stsplus "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/clientset/internalclientset"
	stspluslisters "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/listers/tkex/v1alpha1"

	"github.com/golang/glog"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/retry"
)

// StatefulSetPlusStatusUpdaterInterface is an interface used to update the StatefulSetPlusStatus associated with a StatefulSetPlus.
// For any use other than testing, clients should create an instance using NewRealStatefulSetPlusStatusUpdater.
type StatefulSetPlusStatusUpdaterInterface interface {
	// UpdateStatefulSetPlusStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateStatefulSetPlusStatus(set *stsplus.StatefulSetPlus, status *stsplus.StatefulSetPlusStatus) error
}

// NewRealStatefulSetPlusStatusUpdater returns a StatefulSetPlusStatusUpdaterInterface that updates the Status of a StatefulSetPlus,
// using the supplied client and setLister.
func NewRealStatefulSetPlusStatusUpdater(
	tkexClient tkexclientset.Interface,
	setLister stspluslisters.StatefulSetPlusLister) StatefulSetPlusStatusUpdaterInterface {
	return &realStatefulSetPlusStatusUpdater{tkexClient, setLister}
}

type realStatefulSetPlusStatusUpdater struct {
	tkexClient tkexclientset.Interface
	setLister  stspluslisters.StatefulSetPlusLister
}

func (ssu *realStatefulSetPlusStatusUpdater) UpdateStatefulSetPlusStatus(
	set *stsplus.StatefulSetPlus,
	status *stsplus.StatefulSetPlusStatus) error {
	// Debug Info
	glog.V(3).Infof("Update %s/%s StatefulSetPlus Status: %+v", set.Namespace, set.Name, status)
	// don't wait due to limited number of clients, but backoff after the default number of steps
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		set.Status = *status
		_, updateErr := ssu.tkexClient.TkexV1alpha1().StatefulSetPluses(set.Namespace).UpdateStatus(set)
		if updateErr == nil {
			return nil
		}
		if updated, err := ssu.setLister.StatefulSetPluses(set.Namespace).Get(set.Name); err == nil {
			// make a copy so we don't mutate the shared cache
			set = updated.DeepCopy()
		} else {
			utilruntime.HandleError(fmt.Errorf("error getting updated StatefulSetPlus %s/%s from lister: %v", set.Namespace, set.Name, err))
		}

		return updateErr
	})
}

var _ StatefulSetPlusStatusUpdaterInterface = &realStatefulSetPlusStatusUpdater{}
