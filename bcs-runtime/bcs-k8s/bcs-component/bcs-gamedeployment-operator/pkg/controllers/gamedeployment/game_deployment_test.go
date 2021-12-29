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

package gamedeployment

import (
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"testing"
)

type fixture struct {
	t *testing.T

	client *fake.Clientset
	// Objects to put in the store.
	dLister   []*apps.Deployment
	rsLister  []*apps.ReplicaSet
	podLister []*v1.Pod

	// Actions expected to happen on the client. Objects from here are also
	// preloaded into NewSimpleFake.
	actions []core.Action
	objects []runtime.Object
}

// Test simple gamedeployment

// Test RollingUpdate

// Test InPlaceUpdate

// Test canary step

// Test pod index

// Test PreDeleteHook

// Test delete gamedeployment

// Test scale up gamedeployment

// Test scale down gamedeployment