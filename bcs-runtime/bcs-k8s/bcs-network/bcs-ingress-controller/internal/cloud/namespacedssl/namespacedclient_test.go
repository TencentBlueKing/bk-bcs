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

package namespacedssl

import (
	"fmt"
	"testing"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

type stubSSLClient struct {
	id string
}

func (s *stubSSLClient) DescribeCertificates(certIDs []string) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func newTestClient(objs ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8scorev1.SchemeGroupVersion, &k8scorev1.Secret{})
	scheme.AddKnownTypes(networkextensionv1.GroupVersion, &networkextensionv1.ControllerConfig{})
	return k8sfake.NewFakeClientWithScheme(scheme, objs...)
}

type spySSLFunc struct {
	calls  int
	lastNs string
	err    error
}

func (s *spySSLFunc) fn(data map[string][]byte) (tencentcloud.SSLClient, error) {
	s.calls++
	if v, ok := data["ns"]; ok {
		s.lastNs = string(v)
	}
	if s.err != nil {
		return nil, s.err
	}
	return &stubSSLClient{id: "per-ns"}, nil
}

func makeSecret(ns string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:            IDKeySecretName,
			Namespace:       ns,
			ResourceVersion: "1",
		},
		Data: map[string][]byte{"ns": []byte(ns)},
	}
}

func TestNamespacedSSLGetNsClientExempt(t *testing.T) {
	defaultClient := &stubSSLClient{id: "default"}
	k8sCli := newTestClient()
	spy := &spySSLFunc{}
	nc := NewNamespacedSSLForTest(k8sCli, spy.fn, defaultClient, map[string]struct{}{"bcs-system": {}})

	got, err := nc.GetNsClient("bcs-system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != defaultClient {
		t.Fatalf("exempt ns should use default client")
	}
	if spy.calls != 0 {
		t.Fatalf("newSSLFunc must not be called for exempt ns")
	}
}

func TestNamespacedSSLGetNonExempt(t *testing.T) {
	defaultClient := &stubSSLClient{id: "default"}
	k8sCli := newTestClient(makeSecret("user-ns"))
	spy := &spySSLFunc{}
	nc := NewNamespacedSSLForTest(k8sCli, spy.fn, defaultClient, map[string]struct{}{"bcs-system": {}})

	got, err := nc.GetNsClient("user-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == defaultClient {
		t.Fatalf("non-exempt ns should not use default client")
	}
	if spy.calls != 1 {
		t.Fatalf("expected 1 newSSLFunc call, got %d", spy.calls)
	}
}

func TestNamespacedSSLCredentialMissing(t *testing.T) {
	defaultClient := &stubSSLClient{id: "default"}
	k8sCli := newTestClient()
	spy := &spySSLFunc{}
	nc := NewNamespacedSSLForTest(k8sCli, spy.fn, defaultClient, nil)

	_, err := nc.GetNsClient("missing-ns")
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestNamespacedSSLBatchIsolate(t *testing.T) {
	defaultClient := &stubSSLClient{id: "default"}
	nsClient := &stubSSLClient{id: "ns-a"}
	k8sCli := newTestClient(makeSecret("ns-a"))
	nc := NewNamespacedSSLForTest(k8sCli, func(data map[string][]byte) (tencentcloud.SSLClient, error) {
		return nsClient, nil
	}, defaultClient, nil)

	clientA, err := nc.GetNsClient("ns-a")
	if err != nil {
		t.Fatalf("get ns-a client failed: %v", err)
	}
	if clientA != nsClient {
		t.Fatalf("expected per-ns client for ns-a")
	}

	clientB, err := nc.GetNsClient("bcs-system")
	if err == nil && clientB == nsClient {
		t.Fatalf("unexpected client for missing secret ns")
	}
	_ = fmt.Sprintf("%v", clientB)
}
