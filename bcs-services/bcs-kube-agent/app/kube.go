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

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	"k8s.io/api/core/v1" // nolint
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// NOCC:gas/crypto(工具误报:不包含凭证,只是获取凭证的路径)
	tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token" // nolint
)

// NewKubeClientContext init kube client context
func NewKubeClientContext(kube string) (*KubeClientContext, error) {
	kubeCli := &KubeClientContext{kubeConfig: kube}

	err := kubeCli.setRestConfig()
	if err != nil {
		return nil, err
	}

	err = kubeCli.initKubeClient()
	if err != nil {
		return nil, err
	}

	return kubeCli, nil
}

// KubeClientContext stores k8s client and factory
type KubeClientContext struct {
	kubeConfig string
	cfg        *rest.Config
	kubeClient kubernetes.Interface
}

func (k *KubeClientContext) setRestConfig() error {
	cfg, err := clientcmd.BuildConfigFromFlags("", k.kubeConfig)
	if err != nil {
		return fmt.Errorf("error getting k8s cluster config: %s", err.Error())
	}

	err = handleEmptyKubeConfig(k.kubeConfig, cfg)
	if err != nil {
		return err
	}
	k.cfg = cfg

	return nil
}

func (k *KubeClientContext) initKubeClient() error {
	if k.cfg == nil {
		err := k.setRestConfig()
		if err != nil {
			return err
		}
	}

	kubeClient, err := kubernetes.NewForConfig(k.cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}
	k.kubeClient = kubeClient

	return nil
}

// GetRestConfig get rest config
func (k *KubeClientContext) GetRestConfig() *rest.Config {
	return k.cfg
}

// GetApiserverAddresses get apiserver address
func (k *KubeClientContext) GetApiserverAddresses() (string, error) {
	if k.kubeClient == nil {
		return "", fmt.Errorf("kubeClientContext kubeClient not init")
	}

	var (
		ep              *v1.Endpoints
		err             error
		apiserverPort   int32
		endpointsList   []string
		serverAddresses string
	)

	err = retry.Do(func() error {
		endpoints, errLocal := k.kubeClient.CoreV1().Endpoints(defaultNamespace).Get(
			context.TODO(), clusterServiceName, metav1.GetOptions{})
		if errLocal != nil {
			blog.Errorf("kubeClientContext GetApiserverAddresses failed: %v", errLocal.Error())
			return errLocal
		}

		ep = endpoints
		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Millisecond*500))
	if err != nil {
		return "", err
	}

	for _, subset := range ep.Subsets {
		if len(subset.Addresses) == 0 {
			continue
		}

		// here we only use the apiserver secure-port
		for _, port := range subset.Ports {
			if port.Name == "https" {
				apiserverPort = port.Port
				break
			}
		}
		masterNodes, errLocal := getMasterNodes(k.kubeClient)
		if errLocal != nil {
			return "", errLocal
		}
		for _, node := range masterNodes {
			nodeIP, errLocal := getNodeInternalIP(node)
			if errLocal != nil {
				blog.Warnf("get node internal ip failed, err %s", err.Error())
				continue
			}
			errLocal = pingEndpoint(net.JoinHostPort(nodeIP, strconv.Itoa(int(apiserverPort))))
			if errLocal != nil {
				blog.Warnf("ping endpoint failed, err %s", err.Error())
			} else {
				endpoint := "https://" + net.JoinHostPort(nodeIP, strconv.Itoa(int(apiserverPort)))
				endpointsList = append(endpointsList, endpoint)
			}
		}
	}
	sort.Strings(endpointsList)
	serverAddresses = strings.Join(endpointsList, ",")

	return serverAddresses, nil
}

// Run starts update rest config && kubeClient Periodically
func (k *KubeClientContext) Run(ctx context.Context) {
	// update jwt token if in cluster config && re init kube client Periodically
	if k.kubeConfig == "" {
		wait.UntilWithContext(ctx, func(context.Context) {
			err := handleEmptyKubeConfig(k.kubeConfig, k.cfg)
			if err != nil {
				blog.Errorf("kubeClientContext handleEmptyKubeConfig failed: %v", err)
				return
			}

			blog.V(5).Infof("KubeClientContext run CA: %v", string(k.cfg.CAData))
			blog.V(5).Infof("KubeClientContext run Token: %v", k.cfg.BearerToken)
			blog.V(5).Infof("kubeClientContext update JWT token success")
		}, time.Second*20)

		wait.UntilWithContext(ctx, func(context.Context) {
			err := k.initKubeClient()
			if err != nil {
				blog.Errorf("kubeClientContext initKubeClient failed: %v", err)
				return
			}

			blog.V(5).Infof("kubeClientContext re init kubeClient success")
		}, time.Hour*24)
	}
}

func handleEmptyKubeConfig(config string, cfg *rest.Config) error {
	if config == "" {
		// since go-client 9.0.0, the restclient.Config returned by BuildConfigFromFlags doesn't have BearerToken,
		// so manually get the BearerToken
		token, err := ioutil.ReadFile(tokenFile) // nolint
		if err != nil {
			return fmt.Errorf("error getting the BearerToken: %s", err.Error())
		}
		cfg.BearerToken = string(token)
		if err := populateCAData(cfg); err != nil {
			return fmt.Errorf("error populating ca data: %s", err.Error())
		}
	}

	return nil
}

func populateCAData(cfg *rest.Config) error {
	bytes, err := ioutil.ReadFile(cfg.CAFile)
	if err != nil {
		return err
	}
	cfg.CAData = bytes
	return nil
}

func getNodeInternalIP(node v1.Node) (string, error) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", fmt.Errorf("node %s internal ip is not found", node.GetName())
}

// getMasterNodes xxx
// get the k8s cluster master node
func getMasterNodes(kubeClient kubernetes.Interface) ([]v1.Node, error) {
	var retNodes []v1.Node
	masterNodes, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range masterNodes.Items {
		if isMasterNode(node.Labels) {
			retNodes = append(retNodes, node)
		}
	}
	return retNodes, nil
}

// isMasterNode check master node
func isMasterNode(labels map[string]string) bool {
	_, ok1 := labels[masterRole]
	_, ok2 := labels[controlPlanRole]
	if ok1 || ok2 {
		return true
	}

	return false
}

// pingEndpoint xxx
// probe the health of the apiserver address for 3 times
func pingEndpoint(host string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = dialTLS2(host, time.Second*30)
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			blog.Infof("Error connecting the apiserver %s. Retrying...: %s", host, err.Error())
			time.Sleep(time.Second)
			continue
		} else if err != nil {
			blog.Errorf("Error connecting the apiserver %s: %s", host, err.Error())
			return err
		}
		return err
	}
	return err
}

func dialTLS(host string) error { // nolint:unused
	conf := &tls.Config{
		// NOCC:gas/tls(设计如此:此处需要跳过验证)
		InsecureSkipVerify: true, // nolint
	}
	conn, err := tls.Dial("tcp", host, conf)
	if err != nil {
		return err
	}
	defer conn.Close() // nolint
	return nil
}

func dialTLS2(host string, timeout time.Duration) error {
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", host, &tls.Config{
		// NOCC:gas/tls(设计如此:此处需要跳过验证)
		InsecureSkipVerify: true, // nolint
	})
	if err != nil {
		return err
	}
	defer conn.Close() // nolint
	return nil
}
