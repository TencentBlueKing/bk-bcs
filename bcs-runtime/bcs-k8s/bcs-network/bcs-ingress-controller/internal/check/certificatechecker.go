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

package check

import (
	"context"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/namespacedssl"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// CertificateChecker periodically checks SSL certificate expiry for Ingress resources.
type CertificateChecker struct {
	cli            client.Client
	sslClient      tencentcloud.SSLClient
	namespacedSSL  *namespacedssl.NamespacedSSL
	opts           *option.ControllerOption
	lastBindingSet map[string]struct{}
}

// NewCertificateChecker creates a certificate expiry checker.
func NewCertificateChecker(cli client.Client, sslClient tencentcloud.SSLClient,
	namespacedSSL *namespacedssl.NamespacedSSL, opts *option.ControllerOption) *CertificateChecker {
	return &CertificateChecker{
		cli:            cli,
		sslClient:      sslClient,
		namespacedSSL:  namespacedSSL,
		opts:           opts,
		lastBindingSet: make(map[string]struct{}),
	}
}

// CertificateCheckerRegisterInterval is the CheckRunner interval for CertificateChecker.
const CertificateCheckerRegisterInterval = CheckPer60Min

// ShouldRegisterCertificateChecker reports whether CertificateChecker should be registered.
// Registration requires Tencent Cloud and explicit --certificate_check_enabled=true.
func ShouldRegisterCertificateChecker(cloud string, enabled bool) bool {
	return cloud == constant.CloudTencent && enabled
}

// Run executes one certificate expiry check round.
func (c *CertificateChecker) Run() {
	ingressList := &networkextensionv1.IngressList{}
	if err := c.cli.List(context.TODO(), ingressList); err != nil {
		blog.Errorf("list ingress failed for certificate check, err: %s", err.Error())
		return
	}

	bindings := expandBindings(ingressList.Items)
	currentSet := buildBindingSet(bindings)
	expiryMap, failedCerts := c.queryCertExpiry(bindings)
	c.updateMetrics(bindings, expiryMap, failedCerts)
	c.cleanupStaleMetrics(currentSet)
	c.lastBindingSet = currentSet
}

type nsCertGroup struct {
	namespace string
	certIDs   []string
}

func (c *CertificateChecker) queryCertExpiry(bindings []CertificateBinding) (map[string]int64, map[string]struct{}) {
	expiryMap := make(map[string]int64)
	failedCerts := make(map[string]struct{})

	if c.opts != nil && c.opts.IsNamespaceScope && c.namespacedSSL != nil {
		return c.queryByNamespace(bindings)
	}

	certIDs := collectUniqueCertIDs(bindings)
	if len(certIDs) == 0 {
		return expiryMap, failedCerts
	}
	if c.sslClient == nil {
		for _, id := range certIDs {
			failedCerts[id] = struct{}{}
		}
		return expiryMap, failedCerts
	}

	result, err := c.sslClient.DescribeCertificates(certIDs)
	if err != nil {
		blog.Errorf("DescribeCertificates failed, err: %s", err.Error())
		for _, id := range certIDs {
			failedCerts[id] = struct{}{}
		}
		return expiryMap, failedCerts
	}
	for id, ts := range result {
		expiryMap[id] = ts
	}
	for _, id := range certIDs {
		if _, ok := expiryMap[id]; !ok {
			failedCerts[id] = struct{}{}
			blog.Infof("certificate %s missing or has invalid expiry in API response", id)
		}
	}
	return expiryMap, failedCerts
}

func (c *CertificateChecker) queryByNamespace(bindings []CertificateBinding) (map[string]int64, map[string]struct{}) {
	expiryMap := make(map[string]int64)
	failedCerts := make(map[string]struct{})
	groups := groupCertsByNamespace(bindings)

	for _, group := range groups {
		sslCli, err := c.namespacedSSL.GetNsClient(group.namespace)
		if err != nil {
			blog.Errorf("get ssl client for namespace %s failed, err: %s", group.namespace, err.Error())
			for _, id := range group.certIDs {
				failedCerts[id] = struct{}{}
			}
			continue
		}
		result, err := sslCli.DescribeCertificates(group.certIDs)
		if err != nil {
			blog.Errorf("DescribeCertificates for namespace %s failed, err: %s", group.namespace, err.Error())
			for _, id := range group.certIDs {
				failedCerts[id] = struct{}{}
			}
			continue
		}
		for id, ts := range result {
			expiryMap[id] = ts
		}
		for _, id := range group.certIDs {
			if _, ok := expiryMap[id]; !ok {
				failedCerts[id] = struct{}{}
				blog.Infof("certificate %s in namespace %s missing or has invalid expiry", id, group.namespace)
			}
		}
	}
	return expiryMap, failedCerts
}

func groupCertsByNamespace(bindings []CertificateBinding) []nsCertGroup {
	nsCertMap := make(map[string]map[string]struct{})
	for _, b := range bindings {
		if b.CertID == "" {
			continue
		}
		if _, ok := nsCertMap[b.OwnerNamespace]; !ok {
			nsCertMap[b.OwnerNamespace] = make(map[string]struct{})
		}
		nsCertMap[b.OwnerNamespace][b.CertID] = struct{}{}
	}
	groups := make([]nsCertGroup, 0, len(nsCertMap))
	for ns, certSet := range nsCertMap {
		ids := make([]string, 0, len(certSet))
		for id := range certSet {
			ids = append(ids, id)
		}
		groups = append(groups, nsCertGroup{namespace: ns, certIDs: ids})
	}
	return groups
}

func buildBindingSet(bindings []CertificateBinding) map[string]struct{} {
	set := make(map[string]struct{}, len(bindings))
	for _, b := range bindings {
		set[b.BindingKey()] = struct{}{}
	}
	return set
}

func (c *CertificateChecker) updateMetrics(bindings []CertificateBinding,
	expiryMap map[string]int64, failedCerts map[string]struct{}) {
	now := time.Now()
	metrics.SetCertificateBindingsTotal(float64(len(bindings)))

	for _, b := range bindings {
		labels := b.LabelValues()
		ownerKind, ns, name, certID, role, scope, protocol, port, domain := labels[0], labels[1], labels[2],
			labels[3], labels[4], labels[5], labels[6], labels[7], labels[8]

		if _, failed := failedCerts[b.CertID]; failed {
			metrics.DeleteCertificateDaysUntilExpiry(ownerKind, ns, name, certID, role, scope, protocol, port, domain)
			metrics.SetCertificateQuerySuccess(ownerKind, ns, name, certID, role, scope, protocol, port, domain, 0)
			continue
		}
		endUnix, ok := expiryMap[b.CertID]
		if !ok || endUnix <= 0 {
			blog.Infof("binding %s has no valid expiry for cert %s", b.BindingKey(), b.CertID)
			metrics.DeleteCertificateDaysUntilExpiry(ownerKind, ns, name, certID, role, scope, protocol, port, domain)
			metrics.SetCertificateQuerySuccess(ownerKind, ns, name, certID, role, scope, protocol, port, domain, 0)
			continue
		}
		days := float64(endUnix-now.Unix()) / 86400.0
		metrics.SetCertificateDaysUntilExpiry(ownerKind, ns, name, certID, role, scope, protocol, port, domain, days)
		metrics.SetCertificateQuerySuccess(ownerKind, ns, name, certID, role, scope, protocol, port, domain, 1)
	}
}

func (c *CertificateChecker) cleanupStaleMetrics(currentSet map[string]struct{}) {
	for key := range c.lastBindingSet {
		if _, ok := currentSet[key]; ok {
			continue
		}
		b := bindingFromKey(key)
		if b == nil {
			continue
		}
		labels := b.LabelValues()
		metrics.DeleteCertificateMetrics(labels[0], labels[1], labels[2], labels[3],
			labels[4], labels[5], labels[6], labels[7], labels[8])
	}
}

func bindingFromKey(key string) *CertificateBinding {
	parts := strings.SplitN(key, "|", 7)
	if len(parts) != 7 {
		return nil
	}
	ownerParts := strings.SplitN(parts[0], "/", 3)
	if len(ownerParts) != 3 {
		return nil
	}
	return &CertificateBinding{
		OwnerKind:      ownerParts[0],
		OwnerNamespace: ownerParts[1],
		OwnerName:      ownerParts[2],
		CertID:         parts[1],
		CertRole:       parts[2],
		CertScope:      parts[3],
		Protocol:       parts[4],
		Port:           parts[5],
		Domain:         parts[6],
	}
}
