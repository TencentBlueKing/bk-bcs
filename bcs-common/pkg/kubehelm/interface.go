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

package kubehelm

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
)

//GlobalFlags helm parameters
type GlobalFlags struct {
	KubeApiserver string
	KubeToken     string
	Kubeconfig    string
}

//ParseParameters parse helm parameters
func (f *GlobalFlags) ParseParameters() (string, error) {
	var parameters string
	if f.KubeApiserver != "" && f.KubeToken != "" {
		file, err := ioutil.TempFile("/tmp", "kubeconfig")
		if err != nil {
			return "", nil
		}
		defer file.Close()

		f.Kubeconfig = file.Name()
		config := clientcmdapi.Config{
			APIVersion: "v1",
			Kind:       "Config",
			Clusters:   make([]clientcmdapi.NamedCluster, 0),
			AuthInfos:  make([]clientcmdapi.NamedAuthInfo, 0),
			Contexts:   make([]clientcmdapi.NamedContext, 0),
		}
		cluster := clientcmdapi.NamedCluster{
			Name: "cluster",
			Cluster: clientcmdapi.Cluster{
				Server:                f.KubeApiserver,
				InsecureSkipTLSVerify: true,
			},
		}
		config.Clusters = append(config.Clusters, cluster)
		authInfo := clientcmdapi.NamedAuthInfo{
			Name: "helm",
			AuthInfo: clientcmdapi.AuthInfo{
				Token: f.KubeToken,
			},
		}
		config.AuthInfos = append(config.AuthInfos, authInfo)
		context := clientcmdapi.NamedContext{
			Name: "cluster-context",
			Context: clientcmdapi.Context{
				Cluster:  cluster.Name,
				AuthInfo: authInfo.Name,
			},
		}
		config.Contexts = append(config.Contexts, context)
		config.CurrentContext = context.Name
		by, _ := yaml.Marshal(config)
		_, err = file.Write(by)
		if err != nil {
			return "", err
		}
	}
	if f.Kubeconfig != "" {
		parameters += fmt.Sprintf(" --kubeconfig %s", f.Kubeconfig)
	}
	return parameters, nil
}

//InstallFlags chart parameters
type InstallFlags struct {
	//setParam --set hub=docker.io/istio tag=1.5.4
	SetParam map[string]string
	Chart    string
	Name     string
}

//ParseParameters parse chart parameters
func (f *InstallFlags) ParseParameters() string {
	var parameters string
	if f.Name != "" {
		parameters += fmt.Sprintf(" %s", f.Name)
	}
	if f.Chart != "" {
		parameters += fmt.Sprintf(" %s", f.Chart)
	}
	for k, v := range f.SetParam {
		parameters += fmt.Sprintf(" --set %s=%s", k, v)
	}

	return parameters
}

// KubeHelm kube helm interface
type KubeHelm interface {
	//install
	//setParam --set hub=docker.io/istio tag=1.5.4
	InstallChart(inf InstallFlags, glf GlobalFlags) error
}
