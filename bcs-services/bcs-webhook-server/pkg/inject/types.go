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

package inject

import (
	"net/http"

	"bk-bcs/bcs-services/bcs-webhook-server/options"
	internalclientset "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/clientset/versioned"
	listers "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/k8s"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/mesos"
	"k8s.io/client-go/kubernetes"
)

type WebhookServer struct {
	Server     *http.Server
	KubeClient kubernetes.Interface
	ClientSet  *internalclientset.Clientset
	EngineType string //kubernetes or mesos
	Injects    options.InjectOptions

	BcsLogConfigLister    listers.BcsLogConfigLister
	BcsDbPrivConfigLister listers.BcsDbPrivConfigLister

	K8sLogConfInject    k8s.K8sInject
	K8sDbPrivConfInject k8s.K8sInject
	K8sBscpInject       k8s.K8sInject

	MesosLogConfInject    mesos.MesosInject
	MesosDbPrivConfInject mesos.MesosInject
	MesosBscpInject       mesos.MesosInject
}
