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

// Package steps include all steps for federation manager
package steps

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
)

// ParamsNotFoundError params not found error
func ParamsNotFoundError(taskId, paramKey string) error {
	return fmt.Errorf("task[%s] not exist param: %s", taskId, paramKey)
}

// FormatFederationClusterAddress format federation cluster host
func FormatFederationClusterAddress(address string) string {
	return fmt.Sprintf("https://%s:443", address)
}

// CheckClusterConnection check cluster connection
func CheckClusterConnection(address string) error {
	// todo create certification for bcs-unified-apiserver when install bcs-unified-apiserver and register to cluster manager
	cfg := &rest.Config{
		Host:            address,
		BearerToken:     "xxxxx", //useless but cannot be empty
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("create k8s clientset failed, err: %v", err)
	}

	_, err = clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("check kubeconfig connection failed, err: %v", err)
	}
	return nil
}

// GetBcsUnifiedApiserverAddress check unified apiserver loadbalancer
func GetBcsUnifiedApiserverAddress(clusterId string) (string, error) {
	// get clb ip
	clb, err := cluster.GetClusterClient().GetLoadbalancerIp(&cluster.ResourceGetOptions{
		ClusterId:    clusterId,
		Namespace:    helm.GetHelmClient().GetFederationCharts().Apiserver.ReleaseNamespace,
		ResourceName: helm.GetHelmClient().GetFederationCharts().Apiserver.ReleaseName,
	})
	if err != nil {
		return "", fmt.Errorf("get bcs-unified-apiserver lb address failed, err: %v", err)
	}
	address := FormatFederationClusterAddress(clb)

	return address, nil
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateRandomStr generate random str contains 0-9a-z
func GenerateRandomStr(length int) (string, error) {
	if length <= 0 {
		length = 16
	}

	token := make([]byte, length)
	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果出现错误，使用固定字符
			return "", err
		} else {
			token[i] = charset[num.Int64()]
		}
	}

	return string(token), nil
}
