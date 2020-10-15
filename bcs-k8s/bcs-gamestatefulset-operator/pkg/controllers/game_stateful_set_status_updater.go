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

package gamestatefulset

import (
	"fmt"

	stsplus "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/clientset/internalclientset"
	stspluslisters "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/listers/tkex/v1alpha1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

// GameStatefulSetStatusUpdaterInterface is an interface used to update the GameStatefulSetStatus associated with a GameStatefulSet.
// For any use other than testing, clients should create an instance using NewRealGameStatefulSetStatusUpdater.
type GameStatefulSetStatusUpdaterInterface interface {
	// UpdateGameStatefulSetStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateGameStatefulSetStatus(set *stsplus.GameStatefulSet, status *stsplus.GameStatefulSetStatus) error
}

// NewRealGameStatefulSetStatusUpdater returns a GameStatefulSetStatusUpdaterInterface that updates the Status of a GameStatefulSet,
// using the supplied client and setLister.
func NewRealGameStatefulSetStatusUpdater(
	tkexClient tkexclientset.Interface,
	setLister stspluslisters.GameStatefulSetLister) GameStatefulSetStatusUpdaterInterface {
	return &realGameStatefulSetStatusUpdater{tkexClient, setLister}
}

// realGameStatefulSetStatusUpdater updater implementation
type realGameStatefulSetStatusUpdater struct {
	tkexClient tkexclientset.Interface
	setLister  stspluslisters.GameStatefulSetLister
}

// UpdateGameStatefulSetStatus update gamesatefulset status
func (ssu *realGameStatefulSetStatusUpdater) UpdateGameStatefulSetStatus(
	set *stsplus.GameStatefulSet,
	status *stsplus.GameStatefulSetStatus) error {
	// Debug Info
	klog.V(3).Infof("Update %s/%s GameStatefulSet Status: %+v", set.Namespace, set.Name, status)
	// don't wait due to limited number of clients, but backoff after the default number of steps
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		set.Status = *status
		_, updateErr := ssu.tkexClient.TkexV1alpha1().GameStatefulSets(set.Namespace).UpdateStatus(set)
		if updateErr == nil {
			return nil
		}
		if updated, err := ssu.setLister.GameStatefulSets(set.Namespace).Get(set.Name); err == nil {
			// make a copy so we don't mutate the shared cache
			set = updated.DeepCopy()
		} else {
			utilruntime.HandleError(fmt.Errorf("error getting updated GameStatefulSet %s/%s from lister: %v", set.Namespace, set.Name, err))
		}

		return updateErr
	})
}

var _ GameStatefulSetStatusUpdaterInterface = &realGameStatefulSetStatusUpdater{}
