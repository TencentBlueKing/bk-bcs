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

package pkg

import (
	"flag"
)

// ControllerOption options for controller
type ControllerOption struct {
	IngressControllerNamespace   string
	IngresControllerWorkloadName string

	MaxCloudUpdateConcurrent int

	Mode                     string
	ListenerNameValidateMode string

	Cloud           string
	BcsClusterID    string
	TcpUdpPortReuse bool

	StoreCloudSecretName   string
	KeyStoreCloudSecretKey string
	KeyStoreCloudSecretID  string

	CloudSecretID  string
	CloudSecretKey string
	CloudDomain    string
}

// BindFromCommandLine 读取命令行参数并绑定
func (op *ControllerOption) BindFromCommandLine() {
	flag.StringVar(&op.IngressControllerNamespace, "namespace", "bcs-system", "namespace of bcs ingress controller")
	flag.StringVar(&op.IngresControllerWorkloadName, "workloadname", "bcsingresscontroller", "workload name of bcs ingress controller")

	flag.StringVar(&op.ListenerNameValidateMode, "listener-validate-mode", "STRICT", "[CLOSE,NORMAL,STRICT]")
	flag.IntVar(&op.MaxCloudUpdateConcurrent, "max-cloud-update-concurrent", 10, "max cloud update concurrent")

	flag.Parse()
}
