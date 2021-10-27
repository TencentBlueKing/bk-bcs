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

package expectations

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdate(t *testing.T) {
	controllerKey := "default/controller-test"
	revisions := []string{"rev-0", "rev-1"}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
			Labels: map[string]string{
				"revision": "none",
			},
		},
	}
	c := NewUpdateExpectations(func(p metav1.Object) string { return p.GetLabels()["revision"] })

	// no pod in cache
	if satisfied, _ := c.SatisfiedExpectations(controllerKey, revisions[0]); !satisfied {
		t.Fatalf("expected no pod for revision %v", revisions[0])
	}

	// update to rev-0 and then observed
	tmpPod := pod.DeepCopy()
	c.ExpectUpdated(controllerKey, revisions[0], tmpPod)
	if satisfied, _ := c.SatisfiedExpectations(controllerKey, revisions[0]); satisfied {
		t.Fatalf("expected pod updated for revision %v, got false", revisions[0])
	}

	tmpPod.Labels["revision"] = revisions[0]
	c.ObserveUpdated(controllerKey, revisions[0], tmpPod)
	if satisfied, _ := c.SatisfiedExpectations(controllerKey, revisions[0]); !satisfied {
		t.Fatalf("expected no pod for revision %v", revisions[0])
	}

	// rev-0 up to rev-1
	c.ExpectUpdated(controllerKey, revisions[0], tmpPod)
	if satisfied, _ := c.SatisfiedExpectations(controllerKey, revisions[1]); !satisfied {
		t.Fatalf("expected cache clean when revision updated")
	}
	tmpPod.Labels["revision"] = revisions[1]
	c.ObserveUpdated(controllerKey, revisions[0], tmpPod)
	if satisfied, _ := c.SatisfiedExpectations(controllerKey, revisions[1]); !satisfied {
		t.Fatalf("expected cache clean when revision updated")
	}

	// delete controllerKey
	c.DeleteExpectations(controllerKey)
	if satisfied, _ := c.SatisfiedExpectations(controllerKey, revisions[1]); !satisfied {
		t.Fatalf("expected no pod for revision %v", revisions[0])
	}

}
