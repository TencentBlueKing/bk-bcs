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

package listenercontroller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListenerHelper do some listener operation in go routine
type ListenerHelper struct {
	ctx    context.Context
	client client.Client

	toDeleteListeners map[string]networkextensionv1.Listener
	sync.Mutex
}

// NewListenerHelper return listener helper
func NewListenerHelper(cli client.Client) *ListenerHelper {
	helper := &ListenerHelper{
		ctx:               context.Background(),
		client:            cli,
		toDeleteListeners: make(map[string]networkextensionv1.Listener),
		Mutex:             sync.Mutex{},
	}
	go helper.run()
	return helper
}

// SetDeleteListeners delete listeners
func (l *ListenerHelper) SetDeleteListeners(listeners []networkextensionv1.Listener) {
	l.Lock()
	defer l.Unlock()
	for _, listener := range listeners {
		if listener.DeletionTimestamp != nil {
			blog.V(4).Infof("listener %s/%s has been deleted, skip", listener.GetNamespace(), listener.GetName())
			continue
		}
		key := fmt.Sprintf("%s/%s", listener.GetNamespace(), listener.GetName())
		l.toDeleteListeners[key] = listener
	}
}

func (l *ListenerHelper) run() {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			l.doDeleteListeners()
		case <-l.ctx.Done():
			blog.Infof("ListenerHelper stopped")
			return
		}
	}
}

func (l *ListenerHelper) doDeleteListeners() {
	listenerList := make([]networkextensionv1.Listener, 0, len(l.toDeleteListeners))
	l.Lock()
	// 取出listener 再删除， 避免长时间占用锁
	for _, listener := range l.toDeleteListeners {
		listenerList = append(listenerList, listener)
	}
	l.toDeleteListeners = make(map[string]networkextensionv1.Listener)
	l.Unlock()

	if len(listenerList) != 0 {
		blog.Infof("delete listener(%d)", len(listenerList))
	}
	for _, listener := range listenerList {
		blog.Infof("delete listener %s/%s", listener.GetNamespace(), listener.GetName())
		if err := l.client.Delete(context.TODO(), &listener); err != nil {
			blog.Errorf("delete listener'%s/%s' failed, err: %s", listener.GetNamespace(), listener.GetName(), err.Error())
		}
	}
}
