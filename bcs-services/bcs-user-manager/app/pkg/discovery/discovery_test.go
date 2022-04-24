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

package discovery

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
)

const (
	serviceDomain = "clustermanager.bkbcs.tencent.com"
)

func NewRegistry() registry.Registry {
	etcdEndpoints := "http://127.0.0.1:2379"

	registry := etcd.NewRegistry(
		registry.Addrs(etcdEndpoints),
		registry.Secure(false),
	)
	registry.Init()

	return registry
}

func TestServiceDiscovery_GetRandomServiceInstance(t *testing.T) {
	r := NewRegistry()

	sd := NewServiceDiscovery(serviceDomain, r)
	err := sd.Start()
	if err != nil {
		t.Fatalf("start service discovery failed: %v", err)
	}

	go func() {
		for {
			<-time.After(3 * time.Second)

			node, err := sd.GetRandomServiceInstance()
			if err != nil {
				t.Fatalf("GetRandomServiceInstance failed: %v", err)
			}
			fmt.Printf("node %+v\n", node)
		}
	}()

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signalChan:
		sd.Stop()
	}
}
