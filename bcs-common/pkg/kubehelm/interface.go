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

import "fmt"

type GlobalFlags struct {
	KubeApiserver string
	KubeToken string
	Kubeconfig string
}

func (f *GlobalFlags) ParseParameters()string{
	var parameters string
	if f.KubeApiserver!="" {
		parameters += fmt.Sprintf(" --kube-apiserver %s", f.KubeApiserver)
	}
	if f.KubeToken!="" {
		parameters += fmt.Sprintf(" --kube-token %s", f.KubeToken)
	}
	if f.Kubeconfig!="" {
		parameters += fmt.Sprintf(" --kubeconfig %s", f.Kubeconfig)
	}
	return parameters
}

type InstallFlags struct {
	//setParam --set hub=docker.io/istio tag=1.5.4
	SetParam map[string]string
	Chart string
	Name string
}

func (f *InstallFlags) ParseParameters()string{
	var parameters string
	if f.Name!="" {
		parameters += fmt.Sprintf(" %s", f.Name)
	}
	if f.Chart!="" {
		parameters += fmt.Sprintf(" %s", f.Chart)
	}
	for k,v :=range f.SetParam {
		parameters += fmt.Sprintf(" --set %s=%s", k,v)
	}

	return parameters
}

type KubeHelm interface {
	//install
	//setParam --set hub=docker.io/istio tag=1.5.4
	InstallChart(inf InstallFlags, glf GlobalFlags)error
}
