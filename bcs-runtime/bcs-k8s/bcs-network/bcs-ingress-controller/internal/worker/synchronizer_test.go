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

package worker

import (
	"errors"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/mock"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestEventHandler test event handler function
func TestEventHandler(t *testing.T) {
	listener1 := networkextensionv1.Listener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Listener",
			APIVersion: "networkextension.bkbcs.tencent.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lis-1",
			Namespace: "ns-1",
		},
		Spec: networkextensionv1.ListenerSpec{
			LoadbalancerID: "lb-id",
			Port:           8000,
			EndPort:        0,
			Protocol:       "tcp",
		},
	}

	listener2 := networkextensionv1.Listener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Listener",
			APIVersion: "networkextension.bkbcs.tencent.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lis-2",
			Namespace: "ns-2",
		},
		Spec: networkextensionv1.ListenerSpec{
			LoadbalancerID: "lb-id",
			Port:           8000,
			EndPort:        8001,
			Protocol:       "tcp",
		},
	}

	testCases := []struct {
		name      string
		eventType EventType
		lis       networkextensionv1.Listener
		segLis    networkextensionv1.Listener
		region    string
		lisID     string
		lisErr    error
		segLisID  string
		segLisErr error
		hasErr    bool
	}{
		{
			name:      "test event add",
			eventType: EventAdd,
			lis:       listener1,
			segLis:    listener2,
			region:    "testregion",
			lisID:     "lb-xxxxx",
			lisErr:    nil,
			segLisID:  "lb-xxxxx",
			segLisErr: nil,
			hasErr:    false,
		},
		{
			name:      "empty listener id",
			eventType: EventAdd,
			lis:       listener1,
			segLis:    listener2,
			region:    "testregion",
			lisID:     "",
			lisErr:    errors.New("error"),
			segLisID:  "lb-xxxxx",
			segLisErr: nil,
			hasErr:    true,
		},
		{
			name:      "test event add with seg error",
			eventType: EventAdd,
			lis:       listener1,
			segLis:    listener2,
			region:    "testregion",
			lisID:     "lb-xxxxx",
			lisErr:    nil,
			segLisID:  "",
			segLisErr: errors.New("error"),
			hasErr:    true,
		},
		{
			name:      "test event delete",
			eventType: EventDelete,
			lis:       listener1,
			segLis:    listener2,
			region:    "testregion",
			lisID:     "lb-xxxxx",
			lisErr:    nil,
			segLisID:  "lb-xxxxx",
			segLisErr: nil,
			hasErr:    false,
		},
		{
			name:      "test event delete with error",
			eventType: EventDelete,
			lis:       listener1,
			segLis:    listener2,
			region:    "testregion",
			lisID:     "",
			lisErr:    errors.New("error"),
			segLisID:  "lb-xxxxx",
			segLisErr: nil,
			hasErr:    true,
		},
		{
			name:      "test event delete with seg error",
			eventType: EventDelete,
			lis:       listener1,
			segLis:    listener2,
			region:    "testregion",
			lisID:     "lb-xxxxx",
			lisErr:    nil,
			segLisID:  "",
			segLisErr: errors.New("error"),
			hasErr:    true,
		},
	}

	for index := range testCases {
		t.Run(testCases[index].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCloud := mock.NewMockLoadBalance(ctrl)
			mockCloud.
				EXPECT().
				EnsureListener(testCases[index].region, &testCases[index].lis).
				Return(testCases[index].lisID, testCases[index].lisErr).
				AnyTimes()
			mockCloud.
				EXPECT().
				DeleteListener(testCases[index].region, &testCases[index].lis).
				Return(testCases[index].lisErr).
				AnyTimes()
			mockCloud.
				EXPECT().
				EnsureSegmentListener(testCases[index].region, &testCases[index].segLis).
				Return(testCases[index].segLisID, testCases[index].segLisErr).
				AnyTimes()
			mockCloud.
				EXPECT().
				DeleteSegmentListener(testCases[index].region, &testCases[index].segLis).
				Return(testCases[index].segLisErr).
				AnyTimes()

			newScheme := runtime.NewScheme()
			newScheme.AddKnownTypes(networkextensionv1.GroupVersion, &listener1)
			opt := EventHandlerOption{
				Region:   testCases[index].region,
				LbID:     "lbID",
				LbClient: mockCloud,
				K8sCli: k8sfake.NewFakeClientWithScheme(
					newScheme, &testCases[index].lis, &testCases[index].segLis),
			}
			eventHandler := NewEventHandler(opt)
			eventHandler.eventQueue = workqueue.NewRateLimitingQueue(
				workqueue.NewItemExponentialFailureRateLimiter(
					100*time.Millisecond,
					2*time.Second))
			eventHandler.PushQueue(k8stypes.NamespacedName{
				Namespace: testCases[index].lis.GetNamespace(),
				Name:      testCases[index].lis.GetName(),
			})
			eventHandler.PushQueue(k8stypes.NamespacedName{
				Name:      testCases[index].segLis.GetName(),
				Namespace: testCases[index].segLis.GetNamespace(),
			})
			go eventHandler.RunQueueRecving()
			time.Sleep(100 * time.Millisecond)
			err := eventHandler.doHandleMulti()
			time.Sleep(200 * time.Millisecond)
			if !testCases[index].hasErr && err != nil {
				t.Errorf("test failed, err %s", err.Error())
			}
			if !testCases[index].hasErr && err != nil {
				t.Errorf("test failed, err %s", err.Error())
			}
			if testCases[index].hasErr && len(eventHandler.eventRecvCache.List()) == 0 {
				t.Errorf("expected requeue, but no item requeued")
			}
			if !testCases[index].hasErr && len(eventHandler.eventRecvCache.List()) != 0 {
				t.Errorf("expected no requeue, but get item requeued")
			}
			ctrl.Finish()
		})
	}
}
