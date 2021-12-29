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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/scheme"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/revision"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
	"reflect"
	"testing"
)

func TestGetActiveRevisions(t *testing.T) {
	_ = gdscheme.AddToScheme(scheme.Scheme)
	revisionControl := revision.NewRevisionControl()
	var collisionCount int32

	// initialize test data
	deploy1 := test.NewGameDeployment(1)
	// because revision will hash the spec.template, so we need to change the spec.template
	deploy1.Spec.Template.Labels["test"] = "test1"
	dRev1, err := revisionControl.NewRevision(deploy1, 1, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	deploy2 := test.NewGameDeployment(2)
	deploy2.Spec.Template.Labels["test"] = "test2"
	dRev2, err := revisionControl.NewRevision(deploy2, 2, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	deploy3 := test.NewGameDeployment(3)
	deploy3.Spec.Template.Labels["test"] = "test3"
	dRev3, err := revisionControl.NewRevision(deploy3, 3, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	dRev4, err := revisionControl.NewRevision(deploy2, 4, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		deploy       *v1alpha1.GameDeployment
		revisions    []*apps.ControllerRevision
		podRevisions sets.String

		exceptedCurrentRevision *apps.ControllerRevision
		exceptedUpdateRevision  *apps.ControllerRevision
		exceptedCollisionCount  int32
		exceptedError           error
	}{
		{ // the equivalent revision is the latest revision
			name:                    "the equivalent revision is the latest revision",
			deploy:                  deploy3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2, dRev3},
			podRevisions:            map[string]sets.Empty{dRev2.Name: {}},
			exceptedCurrentRevision: dRev2,
			exceptedUpdateRevision:  dRev3,
			exceptedCollisionCount:  0,
			exceptedError:           nil,
		},
		{ // the equivalent revision isn't the latest revision
			name:                    "the equivalent revision isn't the latest revision",
			deploy:                  deploy2,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2, dRev3},
			podRevisions:            map[string]sets.Empty{dRev3.Name: {}},
			exceptedCurrentRevision: dRev3,
			exceptedUpdateRevision:  dRev4,
			exceptedCollisionCount:  0,
			exceptedError:           nil,
		},
		{ // haven't equivalent revision
			name:                    "haven't equivalent revision",
			deploy:                  deploy3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2},
			podRevisions:            map[string]sets.Empty{dRev1.Name: {}},
			exceptedCurrentRevision: dRev1,
			exceptedUpdateRevision:  dRev3,
			exceptedCollisionCount:  0,
			exceptedError:           nil,
		},
		{ // when initializing, the latest revision is the current revision
			name:                    "when initializing",
			deploy:                  deploy3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2},
			podRevisions:            map[string]sets.Empty{},
			exceptedCurrentRevision: dRev3,
			exceptedUpdateRevision:  dRev3,
			exceptedCollisionCount:  0,
			exceptedError:           nil,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			stop := make(chan struct{})
			defer close(stop)
			informerFactory.Start(stop)
			informer := informerFactory.Apps().V1().ControllerRevisions()
			informerFactory.WaitForCacheSync(stop)
			for i := range s.revisions {
				informer.Informer().GetIndexer().Add(s.revisions[i])
			}
			controllerHistory := history.NewFakeHistory(informer)
			control := &defaultGameDeploymentControl{revisionControl: revisionControl, controllerHistory: controllerHistory}

			currentRevision, updateRevision, collisionCount, err := control.getActiveRevisions(s.deploy, s.revisions, s.podRevisions)
			if err != s.exceptedError {
				t.Errorf("expected error %v, got %v", s.exceptedError, err)
			}
			if !reflect.DeepEqual(currentRevision, s.exceptedCurrentRevision) {
				t.Errorf("expected current revision %v, got %v", s.exceptedCurrentRevision, currentRevision)
			}
			if !reflect.DeepEqual(updateRevision, s.exceptedUpdateRevision) {
				t.Errorf("expected update revision %v, got %v", s.exceptedUpdateRevision, updateRevision)
			}
			if collisionCount != s.exceptedCollisionCount {
				t.Errorf("expected collision count %v, got %v", s.exceptedCollisionCount, collisionCount)
			}
		})
	}
}
