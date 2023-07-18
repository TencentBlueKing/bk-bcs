/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clusterops

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	k8scorecliset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// ErrServerNotInit error for server not init
	ErrServerNotInit = errors.New("server not init")
)

// K8SOperator operator of k8s
type K8SOperator struct {
	opt   *options.ClusterManagerOptions
	model store.ClusterManagerModel
}

// NewK8SOperator create operator of k8s
func NewK8SOperator(opt *options.ClusterManagerOptions, model store.ClusterManagerModel) *K8SOperator {
	return &K8SOperator{
		opt:   opt,
		model: model,
	}
}

// NewKubeClient get k8s client from kubeConfig file
func NewKubeClient(kubeConfig string) (k8scorecliset.Interface, error) {
	data, err := base64.StdEncoding.DecodeString(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("decode kube config failed: %v", err)
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("build rest config failed: %v", err)
	}

	config.Burst = 200
	config.QPS = 100

	return NewKubeClientByRestConfig(config)
}

// NewKubeClientByRestConfig get k8s client from rest config
func NewKubeClientByRestConfig(config *rest.Config) (k8scorecliset.Interface, error) {
	return k8scorecliset.NewForConfig(config)
}

// GetClusterClient get cluster client
func (ko *K8SOperator) GetClusterClient(clusterID string) (k8scorecliset.Interface, error) {
	cred, found, err := ko.model.GetClusterCredential(context.TODO(), clusterID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("cluster credential not found of %s", clusterID)
	}
	cfg := &rest.Config{QPS: 50, Burst: 50}
	if cred.ConnectMode == modules.BCSConnectModeTunnel {
		if len(ko.opt.ClientCert) != 0 && len(ko.opt.ClientCa) != 0 && len(ko.opt.ClientKey) != 0 {
			cfg.Host = "https://" + ko.opt.Address + ":" + strconv.Itoa(int(ko.opt.HTTPPort)) +
				"/clustermanager/clusters/" + clusterID

			_, certData, keyData, err := loadClusterClientCert(clusterID, ko.opt.ClientCa,
				ko.opt.ClientCert, ko.opt.ClientKey, static.ClientCertPwd)
			if err != nil {
				return nil, err
			}

			cfg.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
				CertData: certData,
				KeyData:  keyData,
			}
		} else {
			cfg.Host = "http://" + ko.opt.Address + ":" + strconv.Itoa(int(ko.opt.HTTPPort)) +
				"/clustermanager/clusters/" + clusterID
			cfg.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
			}
		}
		cliset, err := k8scorecliset.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}
		return cliset, nil
	} else if cred.ConnectMode == modules.BCSConnectModeDirect {
		addressList := strings.Split(cred.ServerAddress, ",")
		if len(addressList) == 0 {
			return nil, fmt.Errorf("error credential server addresses %s of cluster %s", cred.ServerAddress, clusterID)
		}
		// get a random server
		rand.Seed(time.Now().Unix())
		cfg.Host = addressList[rand.Intn(len(addressList))]

		if len(cred.CaCertData) != 0 && len(cred.ClientCert) != 0 && len(cred.ClientKey) != 0 {
			cfg.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
				// CAData:   []byte(cred.CaCertData),
				CertData: []byte(cred.ClientCert),
				KeyData:  []byte(cred.ClientKey),
			}
		} else {
			cfg.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
			}
		}
		cfg.BearerToken = cred.UserToken
		cliset, err := k8scorecliset.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}
		return cliset, nil
	}

	return nil, fmt.Errorf("invalid credential mode %s of cluster %s", cred.ConnectMode, clusterID)
}

func loadClusterClientCert(clusterID, clientCa, clientCert, clientKey string, passwd string) ([]byte, []byte, []byte, error) {
	caData, err := loadCertificates(clientCa, "")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cluster[%s] websocketTunnel LoadCertificates(clientCA) failed: %v", clusterID, err)
	}
	certData, err := loadCertificates(clientCert, "")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cluster[%s] websocketTunnel LoadCertificates(clientCert) failed: %v", clusterID, err)
	}
	keyData, err := loadCertificates(clientKey, passwd)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cluster[%s] websocketTunnel LoadCertificates(clientKey) failed: %v", clusterID, err)
	}

	return []byte(caData), []byte(certData), []byte(keyData), nil
}

// loadCertificates parse cert
func loadCertificates(keyFile, passwd string) (string, error) {
	priKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", err
	}

	if "" != passwd {
		priPem, _ := pem.Decode(priKey)
		if priPem == nil {
			return "", fmt.Errorf("decode private key failed")
		}

		priDecrPem, err := x509.DecryptPEMBlock(priPem, []byte(passwd))
		if err != nil {
			return "", err
		}

		priKey = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priDecrPem,
		})
	}

	return string(priKey), nil
}
