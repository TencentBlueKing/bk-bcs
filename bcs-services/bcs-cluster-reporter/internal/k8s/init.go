/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// GetClientsetByConfig cluster k8s client
func GetClientsetByConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

// GetRestClientGetterByConfig restClient
func GetRestClientGetterByConfig(config *rest.Config) *RESTClientGetter {
	return &RESTClientGetter{
		clientconfig: clientcmd.NewDefaultClientConfig(BuildKubeconfig(config), nil),
	}
}

// GetRestConfigByConfig for config
func GetRestConfigByConfig(filePath string) (*rest.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", filePath)
	if err != nil {
		return nil, err
	}
	kubeconfig.QPS = 300
	kubeconfig.Burst = 600
	return kubeconfig, nil
}

// BuildKubeconfig build cluster config
func BuildKubeconfig(config *rest.Config) clientcmdapi.Config {
	kubeConfig := clientcmdapi.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		Clusters:       make(map[string]*clientcmdapi.Cluster),
		AuthInfos:      make(map[string]*clientcmdapi.AuthInfo),
		Contexts:       make(map[string]*clientcmdapi.Context),
		CurrentContext: "default-context",
	}

	kubeConfig.Clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                config.Host,
		CertificateAuthority:  config.CAFile,
		InsecureSkipTLSVerify: config.Insecure,
	}

	kubeConfig.Contexts["default-context"] = &clientcmdapi.Context{
		Cluster:   "default-cluster",
		Namespace: "default",
		AuthInfo:  "default",
	}
	kubeConfig.AuthInfos["default"] = &clientcmdapi.AuthInfo{
		TokenFile: config.BearerTokenFile,
		Username:  config.Username,
		Token:     config.BearerToken,
	}

	return kubeConfig
}

// RESTClientGetter xxx
type RESTClientGetter struct {
	clientconfig clientcmd.ClientConfig
}

// ToRESTConfig xxx
func (r *RESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.clientconfig.ClientConfig()
}

// ToDiscoveryClient discovery client
func (r *RESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	restconfig, err := r.clientconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(restconfig)
	if err != nil {
		return nil, err
	}
	// legacy httpheader仅设置application/json
	dc = dc.WithLegacy().(*discovery.DiscoveryClient)
	return memory.NewMemCacheClient(dc), nil
}

// ToRESTMapper xxx
func (r *RESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(dc), nil
}

// ToRawKubeConfigLoader xxx
func (r *RESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return r.clientconfig
}
