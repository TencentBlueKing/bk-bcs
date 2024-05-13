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

// Package eventer defines watch the event of pod create failed
package eventer

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

type EventHookFunc func(event *kubeapi.Event)

// WatchEventInterface defines the interface the watch events of pod
type WatchEventInterface interface {
	Init() error
	Start(ctx context.Context)

	RegisterEventHook(hookKind string, id string, f EventHookFunc)
	UnRegisterEventHook(hookKind string, id string)
}

type kubeEventer struct {
	eventClient     corev1.EventInterface
	resourceVersion string
	watcher         watch.Interface

	hookMap *sync.Map
}

// NewKubeEventer create the instance of kubeEventer
func NewKubeEventer(client *kubernetes.Clientset) WatchEventInterface {
	eventClient := client.CoreV1().Events(kubeapi.NamespaceAll)
	return &kubeEventer{
		eventClient: eventClient,
		hookMap:     &sync.Map{},
	}
}

// Init will create the client that used to watch the events of pod
func (e *kubeEventer) Init() error {
	if err := e.createEventWatcher(); err != nil {
		return errors.Wrapf(err, "init kube eventer failed")
	}
	return nil
}

// Start the watcher of pod event, it will re-watch if watcher unexpected closed
func (e *kubeEventer) Start(ctx context.Context) {
	blog.Infof("event watch start successfully.")
	var err error
	for {
		if err = e.watch(ctx); err == nil {
			return
		}
		blog.Errorf("event watch closed with err: %s", err.Error())
		e.createWatcher(ctx)
	}
}

// RegisterEventHook register event hook for eventKind
func (e *kubeEventer) RegisterEventHook(hookKind, id string, f EventHookFunc) {
	eventHookMapVal, _ := e.hookMap.LoadOrStore(hookKind, &sync.Map{})
	eventHookMap, ok := eventHookMapVal.(*sync.Map)
	if !ok {
		blog.Errorf("unknown hook type, val: %+v", eventHookMapVal)

		// 如果处理到未知类型，则先用空的Map覆盖
		eventHookMap = &sync.Map{}
		e.hookMap.Store(hookKind, eventHookMap)
	}
	eventHookMap.Store(id, f)
}

// UnRegisterEventHook unregister event hook of eventKind
func (e *kubeEventer) UnRegisterEventHook(eventKind, id string) {
	eventHookMapVal, _ := e.hookMap.LoadOrStore(eventKind, &sync.Map{})
	eventHookMap, ok := eventHookMapVal.(*sync.Map)
	if !ok {
		blog.Errorf("unknown hook type, val: %+v", eventHookMapVal)

		// 如果处理到未知类型，则先用空的Map覆盖
		eventHookMap = &sync.Map{}
		e.hookMap.Store(eventKind, eventHookMap)
	}

	eventHookMap.Delete(id)
}

func (e *kubeEventer) watch(ctx context.Context) error {
	watchChannel := e.watcher.ResultChan()
	for {
		select {
		case event, ok := <-watchChannel:
			if !ok {
				return errors.Errorf("event watch channel is closed")
			}

			if event.Type == kubewatch.Error {
				if status, ok := event.Object.(*metav1.Status); ok {
					return errors.Errorf("event watch occurred err: %v", status)
				}
				return errors.Errorf("event watch received unexpected err: %v", event.Object)
			}
			if obj, ok := event.Object.(*kubeapi.Event); ok {
				go e.handleRegister(obj)
			} else {
				blog.Errorf("event watch received wrong object: %v", obj)
			}
		case <-ctx.Done():
			e.watcher.Stop()
			blog.Warnf("event watch received context done")
			return nil
		}
	}
}

const (
	// podCreateFailedReason 用来判断 Pod 创建失败的事件，目前这个判断仅针对 GameWorkload 创建
	// 出来的 Pod。原生的 Workload 创建出来的 Pod 不会到达 controller 的 webhook，不会触发 Port 从
	// Cache 中分配的逻辑。
	podFailedCreateReason = "FailedCreate"

	// HookKindPodCreateFailed pod eventKind
	HookKindPodCreateFailed = "HookKindPodCreateFailed"
	// KindGameDeployment GameWorkload 类型
	KindGameDeployment = "GameDeployment"
	// KindGameStatefulSet GameWorkload 类型
	KindGameStatefulSet = "GameStatefulSet"

	// HookKindLBEnsured service eventKind
	HookKindLBEnsured = "HookKindLBService"
	// KindService service 类型
	KindService = "Service"
)

func (e *kubeEventer) handleRegister(event *kubeapi.Event) {
	kind := event.InvolvedObject.Kind
	var hookKind string

	if event.Reason == podFailedCreateReason && (kind == KindGameDeployment || kind == KindGameStatefulSet) {
		hookKind = HookKindPodCreateFailed
	} else if kind == KindService {
		hookKind = HookKindLBEnsured
	} else {
		return
	}

	hookMapVal, exist := e.hookMap.Load(hookKind)
	if !exist {
		return
	}
	hookMap, ok := hookMapVal.(*sync.Map)
	if !ok {
		blog.Errorf("unknown type, val:%+v", hookMapVal)
		return
	}

	// 遍历并执行所有注册 Pod FailedCreate 事件的函数
	hookMap.Range(func(key, value interface{}) bool {
		f, ok := value.(EventHookFunc)
		if ok {
			f(event)
		}
		return true
	})
}

const (
	createWatcherDuration = 5
)

func (e *kubeEventer) createWatcher(ctx context.Context) {
	var err error
	ticker := time.NewTicker(createWatcherDuration * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.watcher.Stop()
			if err = e.createEventWatcher(); err != nil {
				blog.Errorf("event watch create watcher failed: %s", err.Error())
				continue
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

func (e *kubeEventer) createEventWatcher() error {
	events, err := e.eventClient.List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return errors.Wrapf(err, "list events failed")
	}
	e.resourceVersion = events.ResourceVersion
	if e.watcher, err = e.eventClient.Watch(context.Background(),
		metav1.ListOptions{
			Watch:           true,
			ResourceVersion: e.resourceVersion,
		}); err != nil {
		return errors.Wrapf(err, "event watch create watcher failed")
	}
	return nil
}
