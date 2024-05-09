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

package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"sort"

	json "github.com/json-iterator/go"
	"github.com/rotisserie/eris"

	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ETCDAuthConfig ETCDAuthConfig
type ETCDAuthConfig struct {
	Addrs               []string `json:"addrs"`
	TLSInsecure         bool     `json:"tlsInsecure"`
	GatewayTLSSecretRef string   `json:"gatewayTLSSecretRef"`
	Username            string   `json:"username"`
	Password            string   `json:"password"`
}

// ServiceConfig for etcd discovery service
type ServiceConfig struct {
	ETCDAuthConfig
	tlsConfig           *tls.Config `json:"-"` // NOCC:vet/vet(设计如此)
	noAuth              bool        `json:"-"` // NOCC:vet/vet(设计如此)
	ServicePortOverride *int64      `json:"servicePortOverride,omitempty"`
	portOverride        int64       `json:"-"` // NOCC:vet/vet(设计如此)
	DisableIPv6         bool        `json:"disableIPv6"`
	IPv6Only            bool        `json:"ipv6Only"`
}

// Validate validate service config
func (c *ServiceConfig) Validate(kubeclient client.Client, namespace string) (bool, error) {
	if len(c.Addrs) == 0 {
		return false, eris.Errorf("Must specify etcd addresses")
	}
	sort.Strings(c.Addrs)
	if len(c.GatewayTLSSecretRef) != 0 {
		secret := corev1.Secret{}
		err := kubeclient.Get(
			context.Background(),
			k8stypes.NamespacedName{Namespace: namespace, Name: c.GatewayTLSSecretRef},
			&secret,
		)
		if err != nil {
			return false, err
		}
		cert := secret.Data["tls.crt"]
		key := secret.Data["tls.key"]
		cacert := secret.Data["ca.crt"]
		if cert != nil && key != nil {
			tlsCert, err := tls.X509KeyPair(cert, key)
			if err != nil {
				return false, err
			}
			c.tlsConfig = &tls.Config{
				// NOCC:gas/tls(设计如此)
				InsecureSkipVerify: c.TLSInsecure,
				Certificates:       []tls.Certificate{tlsCert},
			}
			if cacert != nil {
				caPool := x509.NewCertPool()
				if ok := caPool.AppendCertsFromPEM(cacert); !ok {
					return false, eris.Errorf("append ca cert failed")
				}
				c.tlsConfig.RootCAs = caPool
			}
		} else {
			return false, eris.Errorf("Get cert from secret(%s/%s) failed: no 'tls.crt' or 'tls.key' fields",
				namespace, c.GatewayTLSSecretRef)
		}
	}
	if c.tlsConfig == nil && len(c.Username) == 0 && len(c.Password) == 0 {
		c.noAuth = true
	}
	if c.ServicePortOverride == nil {
		return true, nil
	}
	c.portOverride = *c.ServicePortOverride
	if c.portOverride <= 0 && c.portOverride > 65535 {
		return false, eris.Errorf("Service Port should in range of (0, 65535]")
	}
	if c.DisableIPv6 && c.IPv6Only {
		return false, eris.Errorf("should not set 'ipv6Only' when 'disableIPv6' has been set")
	}
	return true, nil
}

// String return serilized auth config
func (c *ETCDAuthConfig) String() string {
	if c == nil {
		return ""
	}
	by, _ := json.Marshal(c)
	return string(by)
}
