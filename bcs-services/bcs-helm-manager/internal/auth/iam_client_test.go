/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package auth

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

type stubPermClient struct {
	iam.PermClient
	tenantID string
}

func TestIAMClientCachesByTenant(t *testing.T) {
	t.Cleanup(resetIAMClientCache)

	var created int32
	SetIAMClientFactory(func(tenantID string) iam.PermClient {
		atomic.AddInt32(&created, 1)
		return &stubPermClient{tenantID: tenantID}
	})

	first := IAMClient("default")
	second := IAMClient("default")
	if first != second {
		t.Fatal("same tenant should reuse the cached IAM client")
	}
	if atomic.LoadInt32(&created) != 1 {
		t.Fatalf("factory should be called once for one tenant, got %d", created)
	}

	other := IAMClient("other")
	if other == first {
		t.Fatal("different tenant should get a different IAM client")
	}
	if atomic.LoadInt32(&created) != 2 {
		t.Fatalf("factory should be called once per tenant, got %d", created)
	}
}

func TestIAMClientNormalizesEmptyTenant(t *testing.T) {
	t.Cleanup(resetIAMClientCache)

	var created int32
	SetIAMClientFactory(func(tenantID string) iam.PermClient {
		atomic.AddInt32(&created, 1)
		return &stubPermClient{tenantID: tenantID}
	})

	empty := IAMClient("")
	def := IAMClient(iam.DefaultTenantId)
	if empty != def {
		t.Fatal("empty tenantID should share cache with default tenant")
	}
	if atomic.LoadInt32(&created) != 1 {
		t.Fatalf("empty and default tenant should create one client, got %d", created)
	}
}

func TestIAMClientConcurrentSameTenant(t *testing.T) {
	t.Cleanup(resetIAMClientCache)

	var created int32
	SetIAMClientFactory(func(tenantID string) iam.PermClient {
		atomic.AddInt32(&created, 1)
		return &stubPermClient{tenantID: tenantID}
	})

	var wg sync.WaitGroup
	clients := make([]iam.PermClient, 32)
	for i := range clients {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clients[idx] = IAMClient("default")
		}(i)
	}
	wg.Wait()

	if atomic.LoadInt32(&created) != 1 {
		t.Fatalf("concurrent first access should create one client, got %d", created)
	}
	for i := 1; i < len(clients); i++ {
		if clients[i] != clients[0] {
			t.Fatal("concurrent callers should receive the same cached client")
		}
	}
}
