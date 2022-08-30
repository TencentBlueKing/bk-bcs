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

package types

// HaProxyRequest xxx
type HaProxyRequest struct {
	User      string `json:"user"`
	SetID     string `json:"set_id"`
	AppID     int    `json:"app_id"`
	Name      string `json:"name"`
	NameSpace string `json:"namespace"`
	Nbproc    int    `json:"nbproc"`
	Replicas  int    `json:"replicas"`
	Services  []struct {
		Mode        string `json:"mode"`
		ServiceName string `json:"k8s_svc_name"`
		ListenPort  string `json:"listen_port"`
		SecretName  string `json:"secret_name"`
		SslCert     string `json:"ssl_cert"`
	} `json:"services"`
}
