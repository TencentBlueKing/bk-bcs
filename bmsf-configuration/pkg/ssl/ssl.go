/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

// TLSConfig is inner tls config.
type TLSConfig struct {
	// InsecureSkipVerify controls whether a client verifies the
	// server's certificate chain and host name.
	// If InsecureSkipVerify is true, TLS accepts any certificate
	// presented by the server and any host name in that certificate.
	// In this mode, TLS is susceptible to man-in-the-middle attacks.
	// This should be used only for testing.
	InsecureSkipVerify bool

	// CertFile authentication certificate file.
	CertFile string

	// KeyFile authentication key file.
	KeyFile string

	// CAFile authentication root certificate file.
	CAFile string

	// Password authentication key file password.
	Password string
}

// ClientTLSConfVerify load and verify CA/CERT/KEY files of ssl as client side.
func ClientTLSConfVerify(caFile, certFile, keyFile, passwd string) (*tls.Config, error) {
	caPool, err := loadCa(caFile)
	if err != nil {
		return nil, err
	}

	cert, err := loadCertificates(certFile, keyFile, passwd)
	if err != nil {
		return nil, err
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caPool,
		Certificates:       []tls.Certificate{*cert},
	}

	return conf, nil
}

// ServerTLSConfVerify load and verify CA/CERT/KEY files of ssl as server side.
func ServerTLSConfVerify(caFile, certFile, keyFile, passwd string) (*tls.Config, error) {
	caPool, err := loadCa(caFile)
	if err != nil {
		return nil, err
	}

	cert, err := loadCertificates(certFile, keyFile, passwd)
	if err != nil {
		return nil, err
	}

	conf := &tls.Config{
		ClientCAs:    caPool,
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return conf, nil
}

func loadCa(caFile string) (*x509.CertPool, error) {
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(ca); ok != true {
		return nil, fmt.Errorf("append ca cert failed")
	}

	return caPool, nil
}

func loadCertificates(certFile, keyFile, passwd string) (*tls.Certificate, error) {
	priKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	if "" != passwd {
		priPem, _ := pem.Decode(priKey)
		if priPem == nil {
			return nil, fmt.Errorf("decode private key failed")
		}

		priDecrPem, decErr := x509.DecryptPEMBlock(priPem, []byte(passwd))
		if decErr != nil {
			return nil, decErr
		}

		priKey = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priDecrPem,
		})
	}

	certData, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(certData, priKey)
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}
