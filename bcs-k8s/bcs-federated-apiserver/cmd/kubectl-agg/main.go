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

package main

import (
	"os"

	kubectlagg "github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/kubectl-agg"

	"k8s.io/klog"
)

func main() {
	var o kubectlagg.AggPodOptions

	// Parse the command line args.
	err := kubectlagg.ParseKubectlArgs(os.Args, &o)
	if err != nil {
		klog.Errorln("ParseKubectlArgs error.")
		return
	}

	// Create a new FederatedApiServer clientSet.
	clientSet, err := kubectlagg.NewFedApiServerClientSet()
	if err != nil {
		klog.Errorln("new clientSet error.")
		return
	}

	// Get member cluster's Pod list from kubeFedApiServer.
	pods, err := kubectlagg.GetPodAggregationList(clientSet, &o)
	if err != nil {
		klog.Errorln("GetPodAggregationList error.")
		return
	}

	// Output the member cluster's Pod lists.
	kubectlagg.PrintPodAggregation(&o, pods)
}
