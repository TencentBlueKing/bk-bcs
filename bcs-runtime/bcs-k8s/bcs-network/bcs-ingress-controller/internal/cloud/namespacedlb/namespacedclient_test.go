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

package namespacedlb

import (
	"context"
	"strings"
	"testing"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// stubLB is a minimal cloud.LoadBalance whose only purpose is to be
// identity-comparable in tests; its methods are no-ops.
type stubLB struct{ id string }

func (s *stubLB) DescribeLoadBalancer(region, lbID, name, protocolLayer string) (
	*cloud.LoadBalanceObject, error) {
	return nil, nil
}
func (s *stubLB) DescribeLoadBalancerWithNs(ns, region, lbID, name, protocolLayer string) (
	*cloud.LoadBalanceObject, error) {
	return nil, nil
}
func (s *stubLB) IsNamespaced() bool { return false }
func (s *stubLB) EnsureListener(region string, li *networkextensionv1.Listener) (string, error) {
	return "", nil
}
func (s *stubLB) DeleteListener(region string, li *networkextensionv1.Listener) error { return nil }
func (s *stubLB) EnsureMultiListeners(region, lbID string, lis []*networkextensionv1.Listener) (
	map[string]cloud.Result, error) {
	return nil, nil
}
func (s *stubLB) DeleteMultiListeners(region, lbID string, lis []*networkextensionv1.Listener) error {
	return nil
}
func (s *stubLB) EnsureSegmentListener(region string, li *networkextensionv1.Listener) (string, error) {
	return "", nil
}
func (s *stubLB) EnsureMultiSegmentListeners(region, lbID string,
	lis []*networkextensionv1.Listener) (map[string]cloud.Result, error) {
	return nil, nil
}
func (s *stubLB) DeleteSegmentListener(region string, li *networkextensionv1.Listener) error {
	return nil
}
func (s *stubLB) DescribeBackendStatus(region, ns string, lbIDs []string) (
	map[string][]*cloud.BackendHealthStatus, error) {
	return nil, nil
}

// newTestClient returns a fake k8s client with the schemes NamespacedLB needs.
func newTestClient(objs ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(k8scorev1.SchemeGroupVersion, &k8scorev1.Secret{})
	scheme.AddKnownTypes(networkextensionv1.GroupVersion, &networkextensionv1.ControllerConfig{})
	return k8sfake.NewFakeClientWithScheme(scheme, objs...)
}

// spyLBFunc builds a newLBFunc that records how many times it was called and the
// namespaces it was asked about (via the credentials it received); every invocation
// returns a freshly-allocated stubLB so pointer identity distinguishes per-ns clients
// from the exempt defaultClient.
type spyLBFunc struct {
	calls  int
	err    error
	lastNs string
}

func (s *spyLBFunc) fn(data map[string][]byte, _ client.Client,
	_ eventer.WatchEventInterface) (cloud.LoadBalance, error) {
	s.calls++
	if v, ok := data["ns"]; ok {
		s.lastNs = string(v)
	}
	if s.err != nil {
		return nil, s.err
	}
	return &stubLB{id: "per-ns"}, nil
}

// newNamespacedLBForTest constructs a NamespacedLB without the background
// reload goroutine so tests have deterministic state.
func newNamespacedLBForTest(k8sCli client.Client,
	newLB func(map[string][]byte, client.Client, eventer.WatchEventInterface) (cloud.LoadBalance, error),
	defaultClient cloud.LoadBalance, exempt map[string]struct{}) *NamespacedLB {
	return &NamespacedLB{
		k8sClient:                  k8sCli,
		nsClientSet:                make(map[string]cloud.LoadBalance),
		ncClientResourceVersionMap: make(map[string]string),
		newLBFunc:                  newLB,
		defaultClient:              defaultClient,
		exemptNamespaces:           exempt,
	}
}

// makeSecret returns a k8s Secret carrying a per-ns marker so spyLBFunc can log which ns it saw.
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

// ---------- Tests ----------

// TestIsExempt verifies the exemption helper combines defaultClient and the set correctly.
func TestIsExempt(t *testing.T) {
	defaultLB := &stubLB{id: "default"}

	cases := []struct {
		name             string
		defaultClient    cloud.LoadBalance
		exemptNamespaces map[string]struct{}
		ns               string
		want             bool
	}{
		{"nil default client disables exemption", nil,
			map[string]struct{}{"bcs-system": {}}, "bcs-system", false},
		{"nil exempt map disables exemption", defaultLB, nil, "bcs-system", false},
		{"empty exempt map disables exemption", defaultLB,
			map[string]struct{}{}, "bcs-system", false},
		{"ns listed returns true", defaultLB,
			map[string]struct{}{"bcs-system": {}, "kube-system": {}}, "bcs-system", true},
		{"ns not listed returns false", defaultLB,
			map[string]struct{}{"bcs-system": {}}, "user-ns", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			nc := &NamespacedLB{
				defaultClient:    c.defaultClient,
				exemptNamespaces: c.exemptNamespaces,
			}
			if got := nc.isExempt(c.ns); got != c.want {
				t.Errorf("isExempt(%q) = %v, want %v", c.ns, got, c.want)
			}
		})
	}
}

// TestGetNsClient_ExemptReturnsDefault ensures exempt namespaces bypass k8sClient.Get
// entirely and always return the default (global-credentials) client.
func TestGetNsClientExempt(t *testing.T) {
	defaultLB := &stubLB{id: "default"}
	// No Secret / ControllerConfig exists for "bcs-system" so if the exempt path
	// failed to short-circuit, initNsClient would return a "not found" error.
	k8sCli := newTestClient()

	spy := &spyLBFunc{}
	nc := newNamespacedLBForTest(k8sCli, spy.fn, defaultLB,
		map[string]struct{}{"bcs-system": {}})

	got, err := nc.getNsClient("bcs-system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != defaultLB {
		t.Errorf("got %+v, want defaultLB", got)
	}
	if _, cached := nc.nsClientSet["bcs-system"]; cached {
		t.Errorf("exempt namespace must not be cached in nsClientSet")
	}
	if spy.calls != 0 {
		t.Errorf("newLBFunc must not be invoked for exempt ns, got %d calls", spy.calls)
	}
}

// TestGetNsClient_NonExemptFetchesSecret ensures non-exempt namespaces still
// go through the per-namespace secret lookup and cache the resulting client.
func TestGetNsClientNonExempt(t *testing.T) {
	defaultLB := &stubLB{id: "default"}
	k8sCli := newTestClient(makeSecret("user-ns"))
	spy := &spyLBFunc{}

	nc := newNamespacedLBForTest(k8sCli, spy.fn, defaultLB,
		map[string]struct{}{"bcs-system": {}})

	got, err := nc.getNsClient("user-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == defaultLB {
		t.Errorf("non-exempt ns must NOT reuse defaultLB")
	}
	if spy.calls != 1 {
		t.Errorf("newLBFunc should be invoked exactly once, got %d", spy.calls)
	}
	if spy.lastNs != "user-ns" {
		t.Errorf("newLBFunc saw ns %q, want user-ns", spy.lastNs)
	}
	if _, cached := nc.nsClientSet["user-ns"]; !cached {
		t.Errorf("non-exempt ns should be cached after first lookup")
	}

	// second call should hit the cache and not invoke newLBFunc again
	if _, err = nc.getNsClient("user-ns"); err != nil {
		t.Fatalf("second call unexpected err: %v", err)
	}
	if spy.calls != 1 {
		t.Errorf("second call should hit cache; total calls want 1, got %d", spy.calls)
	}
}

// TestGetNsClient_NonExemptMissingSecretReturnsError confirms the pre-existing
// error behavior is preserved for non-exempt namespaces when no secret exists.
func TestGetNsClientMissingSecret(t *testing.T) {
	k8sCli := newTestClient()
	spy := &spyLBFunc{}
	nc := newNamespacedLBForTest(k8sCli, spy.fn, &stubLB{id: "default"},
		map[string]struct{}{"bcs-system": {}})

	_, err := nc.getNsClient("user-ns")
	if err == nil {
		t.Fatal("expected error when no secret/controllerconfig exists in non-exempt ns")
	}
	if !strings.Contains(err.Error(), "not found secret or controllerConfig") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestGetNsClient_BackwardCompat ensures that constructing NamespacedLB without
// the new exemption configuration (nil defaults) preserves the original behavior.
func TestGetNsClient_BackwardCompat(t *testing.T) {
	k8sCli := newTestClient(makeSecret("ns1"))
	spy := &spyLBFunc{}

	// defaultClient and exemptNamespaces both nil -> every call goes through per-ns path
	nc := newNamespacedLBForTest(k8sCli, spy.fn, nil, nil)

	got, err := nc.getNsClient("ns1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil client")
	}
	if spy.calls != 1 {
		t.Errorf("newLBFunc should still be invoked exactly once, got %d", spy.calls)
	}

	// even if the caller happens to ask for a ns named "bcs-system", with no
	// exempt set configured, we still go through the normal path (and thus fail).
	_, err = nc.getNsClient("bcs-system")
	if err == nil {
		t.Errorf("without exemption configured, bcs-system should hit the normal path "+
			"and fail because its secret is missing; got nil err")
	}
}

// TestReloadNsClient_SkipsExempt confirms that exempt namespaces that somehow
// ended up in nsClientSet are skipped by reloadNsClient (defensive behavior);
// more importantly, no reload happens for exempt ns when the cache is empty.
func TestReloadNsClient_SkipsExempt(t *testing.T) {
	defaultLB := &stubLB{id: "default"}
	// secret exists for user-ns but NOT for bcs-system; if reload tried to read
	// bcs-system's secret it would log an error (and leave the cached entry
	// untouched). We assert spy.calls does not increase for the exempt ns.
	k8sCli := newTestClient(makeSecret("user-ns"))
	spy := &spyLBFunc{}

	nc := newNamespacedLBForTest(k8sCli, spy.fn, defaultLB,
		map[string]struct{}{"bcs-system": {}})

	// prime nsClientSet with a user-ns entry (normal path) so reload has work to do
	if _, err := nc.getNsClient("user-ns"); err != nil {
		t.Fatalf("priming user-ns failed: %v", err)
	}
	spyCallsAfterPrime := spy.calls

	// defensively insert an entry for bcs-system directly; reload must skip it
	nc.nsClientSet["bcs-system"] = defaultLB

	nc.reloadNsClient()

	// Same resource version -> user-ns should NOT trigger a rebuild,
	// and bcs-system must be skipped entirely.
	if spy.calls != spyCallsAfterPrime {
		t.Errorf("reloadNsClient triggered unexpected newLBFunc calls: before=%d after=%d",
			spyCallsAfterPrime, spy.calls)
	}
	// The defensive entry should remain pointing to defaultLB (untouched).
	if nc.nsClientSet["bcs-system"] != defaultLB {
		t.Errorf("bcs-system entry was mutated by reloadNsClient")
	}
}

// TestNewNamespacedLB_SignatureAcceptsNil smoke-tests that the new constructor
// signature accepts nil defaultClient and nil exempt set without panicking,
// and that the background goroutine is cancellable via ctx.
func TestNewNamespacedLBNilParams(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	k8sCli := newTestClient()
	nc := NewNamespacedLB(ctx, k8sCli, nil,
		func(map[string][]byte, client.Client,
			eventer.WatchEventInterface) (cloud.LoadBalance, error) {
			return &stubLB{id: "per-ns"}, nil
		}, nil, nil)
	if nc == nil {
		t.Fatal("NewNamespacedLB returned nil")
	}
	if nc.isExempt("whatever") {
		t.Errorf("isExempt should be false when defaults are nil")
	}
}
