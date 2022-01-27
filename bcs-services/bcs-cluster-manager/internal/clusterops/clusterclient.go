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
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	k8scorecliset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// GetClusterClient get cluster client
func (ko *K8SOperator) GetClusterClient(clusterID string) (k8scorecliset.Interface, error) {
	cred, found, err := ko.model.GetClusterCredential(context.TODO(), clusterID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("cluster credential not found of %s", clusterID)
	}
	cfg := &rest.Config{}
	if cred.ConnectMode == modules.BCSConnectModeTunnel {
		if len(ko.opt.ClientCert) != 0 && len(ko.opt.ClientCa) != 0 && len(ko.opt.ClientKey) != 0 {
			cfg.Host = "https://" + ko.opt.Address + ":" + strconv.Itoa(int(ko.opt.HTTPPort)) +
				"/clustermanager/clusters/" + clusterID

			cfg.TLSClientConfig = rest.TLSClientConfig{
				Insecure: false,
				CertFile: ko.opt.ClientCert,
				CAFile:   ko.opt.ClientCa,
				KeyFile:  ko.opt.ClientKey,
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
		cfg.TLSClientConfig = rest.TLSClientConfig{
			Insecure: false,
			CAData:   []byte(cred.CaCertData),
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
