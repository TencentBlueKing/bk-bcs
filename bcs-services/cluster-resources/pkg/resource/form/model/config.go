/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

// CM ConfigMap 表单化建模
type CM struct {
	Metadata Metadata `structs:"metadata"`
	Data     CMData   `structs:"data"`
}

// CMData xxx
type CMData struct {
	Items []OpaqueData `structs:"items"`
}

// Secret 表单化建模
type Secret struct {
	Metadata Metadata   `structs:"metadata"`
	Data     SecretData `structs:"data"`
}

// SecretData xxx
type SecretData struct {
	Type      string             `structs:"type"`
	Opaque    []OpaqueData       `structs:"opaque"`
	Docker    DockerRegistryData `structs:"docker"`
	BasicAuth BasicAuthData      `structs:"basicAuth"`
	SSHAuth   SSHAuthData        `structs:"sshAuth"`
	TLS       TLSData            `structs:"tls"`
}

// OpaqueData Key-Value 类型数据
type OpaqueData struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// DockerRegistryData Docker 配置信息类型数据
type DockerRegistryData struct {
	RepoAddr string `structs:"repoAddr"`
	Username string `structs:"username"`
	Password string `structs:"password"`
}

// BasicAuthData xxx
type BasicAuthData struct {
	Username string `structs:"username"`
	Password string `structs:"password"`
}

// SSHAuthData xxx
type SSHAuthData struct {
	PublicKey  string `structs:"publicKey"`
	PrivateKey string `structs:"privateKey"`
}

// TLSData xxx
type TLSData struct {
	PublicKey string `structs:"publicKey"`
	Cert      string `structs:"cert"`
}
