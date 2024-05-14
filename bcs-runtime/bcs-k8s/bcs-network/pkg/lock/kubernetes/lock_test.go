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

package kubernetes

import (
	"sync"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/lock"

	k8skube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestKubeLock(b *testing.T) {

	blog.InitLogs(conf.LogConfig{
		LogDir:          "./logs",
		LogMaxSize:      500,
		LogMaxNum:       10,
		AlsoToStdErr:    true,
		Verbosity:       5,
		StdErrThreshold: "2",
	})
	defer blog.CloseLogs()
	kubeconfigPath := "/root/.kube/config"
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		b.Errorf("build restConfig by file %s failed, err %s", kubeconfigPath, err.Error())
	}
	k8sClient1, err := k8skube.NewForConfig(restConfig)
	if err != nil {
		b.Errorf("build kubeClient failed, err %s", err.Error())
	}
	cmStore1 := &ConfigmapStore{
		prefix:    "lock-test",
		namespace: "default",
		cmClient:  k8sClient1.CoreV1(),
	}
	k8sClient2, err := k8skube.NewForConfig(restConfig)
	if err != nil {
		b.Errorf("build kubeClient failed, err %s", err.Error())
	}
	cmStore2 := &ConfigmapStore{
		prefix:    "lock-test",
		namespace: "default",
		cmClient:  k8sClient2.CoreV1(),
	}
	locker1 := &Locker{
		lockerName:      "locker1",
		store:           cmStore1,
		timeoutDuration: 20 * time.Second,
		renewDuration:   2 * time.Second,
		retryDuration:   1 * time.Second,
		locks:           make(map[string]*kubeLock),
	}
	locker2 := &Locker{
		lockerName:      "locker2",
		store:           cmStore2,
		timeoutDuration: 20 * time.Second,
		renewDuration:   5 * time.Second,
		retryDuration:   1 * time.Second,
		locks:           make(map[string]*kubeLock),
	}
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	var wg sync.WaitGroup
	for _, key := range keys {
		tmpKey := key
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := locker1.Lock(tmpKey, lock.LockTTL(10*time.Second)); err != nil {
				b.Errorf("locker1 lock key %s failed, err %s", tmpKey, err.Error())
				return
			}
			b.Logf("locker1 hold %s, time %s", tmpKey, time.Now().String())
			time.Sleep(10 * time.Millisecond)
			b.Logf("locker1 released %s, time %s", tmpKey, time.Now().String())
			locker1.Unlock(tmpKey)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := locker1.Lock(tmpKey, lock.LockTTL(10*time.Second)); err != nil {
				b.Errorf("locker1 lock key %s failed, err %s", tmpKey, err.Error())
				return
			}
			b.Logf("locker1 hold %s, time %s", tmpKey, time.Now().String())
			time.Sleep(10 * time.Millisecond)
			b.Logf("locker1 released %s, time %s", tmpKey, time.Now().String())
			locker1.Unlock(tmpKey)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := locker2.Lock(tmpKey, lock.LockTTL(10*time.Second)); err != nil {
				b.Errorf("locker2 lock key %s failed, err %s", tmpKey, err.Error())
				return
			}
			b.Logf("locker2 hold %s, time %s", tmpKey, time.Now().String())
			time.Sleep(10 * time.Millisecond)
			b.Logf("locker2 released %s, time %s", tmpKey, time.Now().String())
			locker2.Unlock(tmpKey)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := locker2.Lock(tmpKey, lock.LockTTL(10*time.Second)); err != nil {
				b.Errorf("locker1 lock key %s failed, err %s", tmpKey, err.Error())
				return
			}
			b.Logf("locker2 hold %s, time %s", tmpKey, time.Now().String())
			time.Sleep(10 * time.Millisecond)
			b.Logf("locker2 released %s, time %s", tmpKey, time.Now().String())
			locker2.Unlock(tmpKey)
		}()
	}
	wg.Wait()
	b.Error()
}
