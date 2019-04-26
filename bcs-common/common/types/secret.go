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

//SecretDataItem secret detail
type SecretDataItem struct {
	//Path    string `json:"path,omitempty"` //mesos only
	Content string `json:"content"`
}

//BcsSecretType type for secret
type BcsSecretType string

const (
	BcsSecretTypeOpaque              BcsSecretType = "Opaque"
	BcsSecretTypeServiceAccountToken BcsSecretType = "kubernetes.io/service-account-token"
	BcsSecretTypeDockercfg           BcsSecretType = "kubernetes.io/dockercfg"
	BcsSecretTypeDockerConfigJson    BcsSecretType = "kubernetes.io/dockerconfigjson"
	BcsSecretTypeBasicAuth           BcsSecretType = "kubernetes.io/basic-auth"
	BcsSecretTypeSSHAuth             BcsSecretType = "kubernetes.io/ssh-auth"
	BcsSecretTypeTLS                 BcsSecretType = "kubernetes.io/tls"
)

//BcsSecret bcs secret definition
type BcsSecret struct {
	TypeMeta `json:",inline"`
	//AppMeta    `json:",inline"`
	ObjectMeta `json:"metadata"`
	Type       BcsSecretType             `json:"type,omitempty"` //k8s only
	Data       map[string]SecretDataItem `json:"datas"`
}
