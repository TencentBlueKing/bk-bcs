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

package types

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type TLSCollectorCfg struct {
	IsTLS        bool   `json:"isTLS"`
	CA           string `json:"ca"`
	ClientCert   string `json:"clientCert"`
	ClientKey    string `json:"clientKey"`
	ClientKeyPwd string `json:"clientKeyPwd"`
}

func (tc TLSCollectorCfg) GetTLSConfig() (c *tls.Config, err error) {
	if !tc.IsTLS {
		return nil, nil
	}

	var caPool *x509.CertPool
	var certificate tls.Certificate
	if caPool, err = tc.GetCAPool(); err != nil {
		return
	}
	if certificate, err = tc.GetCertificate(); err != nil {
		return
	}

	c = &tls.Config{
		RootCAs:            caPool,
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: true,
	}
	return
}

func (tc TLSCollectorCfg) GetCAPool() (r *x509.CertPool, err error) {
	var ca []byte
	if ca, err = tc.GetCA(); err != nil {
		return
	}

	r = x509.NewCertPool()
	if ok := r.AppendCertsFromPEM(ca); !ok {
		err = fmt.Errorf("append ca cert failed")
		return
	}
	return
}

func (tc TLSCollectorCfg) GetCA() ([]byte, error) {
	return base64.StdEncoding.DecodeString(tc.CA)
}

func (tc TLSCollectorCfg) GetClientCert() ([]byte, error) {
	return base64.StdEncoding.DecodeString(tc.ClientCert)
}

func (tc TLSCollectorCfg) GetClientKey() (key []byte, err error) {
	if key, err = base64.StdEncoding.DecodeString(tc.ClientKey); err != nil {
		return
	}

	if len(key) == 0 {
		return
	}

	if tc.ClientKeyPwd != "" {
		priPem, _ := pem.Decode(key)
		if priPem == nil {
			return nil, fmt.Errorf("decode private key failed")
		}

		priDecryptPem, err := x509.DecryptPEMBlock(priPem, []byte(tc.ClientKeyPwd))
		if err != nil {
			return nil, err
		}

		key = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priDecryptPem,
		})
	}
	return
}

func (tc TLSCollectorCfg) GetCertificate() (r tls.Certificate, err error) {
	var cert, key []byte
	if cert, err = tc.GetClientCert(); err != nil {
		return
	}
	if key, err = tc.GetClientKey(); err != nil {
		return
	}

	if len(cert) == 0 || len(key) == 0 {
		return
	}

	return tls.X509KeyPair(cert, key)
}
