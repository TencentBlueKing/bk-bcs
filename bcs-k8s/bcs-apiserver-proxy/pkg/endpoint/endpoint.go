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

package endpoint

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	masterLabel       = "node-role.kubernetes.io/master"
	defaultNamespace  = "default"
	defaultKubernetes = "kubernetes"
	schemeHTTPS       = "https"
	schemeHTTP        = "http"
)

var (
	// ErrEndpointsClientNotInited show endpointsClient not inited
	ErrEndpointsClientNotInited = errors.New("endpointsClient not inited")
)

// EndpointsHealthOptions show health check scheme&path
type EndpointsHealthOptions struct {
	Scheme string
	Path   string
}

var (
	defaultHealthScheme = schemeHTTPS
	defaultHealthPath   = "/healthz"

	defaultInterval = time.Second * 3
)

// ClusterEndpointsIP is a interface for sync kubernetes master endpointIPs
type ClusterEndpointsIP interface {
	GetClusterEndpoints() ([]utils.EndPoint, error)
	SyncClusterEndpoints(ctx context.Context)
	Stop()
}

// NewEndpointsClient init endpoints client
func NewEndpointsClient(opts ...EndpointsClientOption) (ClusterEndpointsIP, error) {
	defaultOptions := &EndpointsClientOptions{
		K8sConfig: K8sConfig{
			Mater:      "",
			KubeConfig: "",
		},
		HealthConfig: EndpointsHealthOptions{
			Scheme: defaultHealthScheme,
			Path:   defaultHealthPath,
		},
		Interval: defaultInterval,
	}

	for _, opt := range opts {
		opt(defaultOptions)
	}

	clientSet, err := defaultOptions.K8sConfig.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	ec := &endpointsClient{
		healthOptions: defaultOptions.HealthConfig,
		interval:      defaultOptions.Interval,
		debug:         defaultOptions.Debug,

		Mutex:           sync.Mutex{},
		clientSet:       clientSet,
		masterEndpoints: []utils.EndPoint{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	ec.ctx = ctx
	ec.cancel = cancel

	return ec, nil
}

// EndpointsClientOption func for set EndpointsClientOptions
type EndpointsClientOption func(options *EndpointsClientOptions)

// EndpointsClientOptions conf options
type EndpointsClientOptions struct {
	K8sConfig    K8sConfig
	HealthConfig EndpointsHealthOptions
	Interval     time.Duration
	Debug        bool
}

// WithK8sConfig set k8sConfig
func WithK8sConfig(ks8Config K8sConfig) EndpointsClientOption {
	return func(opts *EndpointsClientOptions) {
		opts.K8sConfig = ks8Config
	}
}

// WithHealthConfig set health check options
func WithHealthConfig(healthConfig EndpointsHealthOptions) EndpointsClientOption {
	return func(opts *EndpointsClientOptions) {
		opts.HealthConfig = healthConfig
	}
}

// WithInterval set interval
func WithInterval(interval time.Duration) EndpointsClientOption {
	return func(opts *EndpointsClientOptions) {
		opts.Interval = interval
	}
}

// WithDebug set debug for unit test
func WithDebug(debug bool) EndpointsClientOption {
	return func(opts *EndpointsClientOptions) {
		opts.Debug = debug
	}
}

type endpointsClient struct {
	healthOptions EndpointsHealthOptions
	interval      time.Duration

	sync.Mutex
	clientSet       kubernetes.Interface
	masterEndpoints []utils.EndPoint

	debug  bool
	ctx    context.Context
	cancel context.CancelFunc
}

// SyncClusterEndpoints sync cluster master endpointIPs
func (ec *endpointsClient) SyncClusterEndpoints(ctx context.Context) {
	if ec == nil {
		blog.Errorf("SyncClusterEndpoints failed; %v", ErrEndpointsClientNotInited)
		return
	}

	select {
	case <-ctx.Done():
		blog.Errorf("external context quit: %v", ctx.Err())
		return
	case <-ec.ctx.Done():
		blog.Errorf("ec context quit: %v", ec.ctx.Err())
		return
	default:
	}

	quitCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-quitCtx.Done():
			blog.Errorf("external context quit: %v", quitCtx.Err())
		case <-ec.ctx.Done():
			blog.Errorf("EndpointsClient context quit: %v", ec.ctx.Err())
			cancel()
		}
	}()

	coldStart := make(chan struct{}, 1)
	coldStart <- struct{}{}

	ticker := time.NewTicker(ec.interval)
	defer ticker.Stop()

	for {
		select {
		case <-quitCtx.Done():
			blog.Infof("EndpointsClient quit: %v", quitCtx.Err())
			return
		case <-coldStart:
		case <-ticker.C:
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					blog.Errorf("EndpointsClient panic: %v", r)
				}
			}()

			// get apiServer Endpoints
			clusterEndpoints, err := ec.getAPIServerEndpoints()
			if err != nil {
				blog.Errorf("getAPIServerEndpoints failed: %v", err)
				return
			}

			ec.Mutex.Lock()
			ec.masterEndpoints = clusterEndpoints
			ec.Mutex.Unlock()
		}()
	}

}

// GetClusterEndpoints get cluster endpointIPs
func (ec *endpointsClient) GetClusterEndpoints() ([]utils.EndPoint, error) {
	if ec == nil {
		return nil, ErrEndpointsClientNotInited
	}

	ec.Mutex.Lock()
	clusterEndpoints := ec.masterEndpoints
	ec.Mutex.Unlock()

	return clusterEndpoints, nil
}

// Stop close sync
func (ec *endpointsClient) Stop() {
	if ec == nil {
		return
	}

	ec.cancel()
}

func (ec *endpointsClient) getMaterNodes() ([]corev1.Node, error) {
	if ec == nil {
		return nil, ErrEndpointsClientNotInited
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clusterNodes, err := ec.clientSet.CoreV1().Nodes().List(timeoutCtx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	masterNodes := []corev1.Node{}
	for _, node := range clusterNodes.Items {
		if _, ok := node.Labels[masterLabel]; ok {
			masterNodes = append(masterNodes, node)
		}
	}

	return masterNodes, nil
}

func (ec *endpointsClient) getAPIServerEndpoints() ([]utils.EndPoint, error) {
	if ec == nil {
		return nil, ErrEndpointsClientNotInited
	}
	timeoutCtx, cancel := context.WithTimeout(ec.ctx, 5*time.Second)
	defer cancel()

	var (
		apiServerPort      uint32
		apiserverEndpoints []utils.EndPoint
	)

	// healthCheck client
	healthCheck, err := NewHealthConfig(ec.healthOptions.Scheme, ec.healthOptions.Path)
	if err != nil {
		blog.Errorf("NewHealthConfig failed: %v", err)
		return nil, err
	}

	endpoints, err := ec.clientSet.CoreV1().Endpoints(defaultNamespace).Get(timeoutCtx,
		defaultKubernetes, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, subset := range endpoints.Subsets {
		if len(subset.Addresses) == 0 {
			continue
		}

		// get apiServer secure-port
		for _, port := range subset.Ports {
			if port.Name == schemeHTTPS {
				apiServerPort = uint32(port.Port)
				break
			}
		}

		masterNodes, err := ec.getMaterNodes()
		if err != nil {
			return nil, err
		}

		for _, node := range masterNodes {
			nodeIP, err := getNodeIP(node)
			if err != nil {
				blog.Errorf("getNodeInternalIP failed: %v", err)
				continue
			}

			if ec.debug {
				apiserverEndpoints = append(apiserverEndpoints, utils.EndPoint{
					IP:   nodeIP,
					Port: apiServerPort,
				})
				continue
			}

			health := healthCheck.IsHTTPAPIHealth(nodeIP, apiServerPort)
			if health {
				apiserverEndpoints = append(apiserverEndpoints, utils.EndPoint{
					IP:   nodeIP,
					Port: apiServerPort,
				})
			}
		}
	}

	return apiserverEndpoints, nil
}

func getNodeIP(node corev1.Node) (string, error) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address, nil
		}
	}

	errMsg := fmt.Sprintf("node %s internalIP not found!", node.GetName())
	return "", errors.New(errMsg)
}
