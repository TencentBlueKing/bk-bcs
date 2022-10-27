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

package formdata

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// CMComplex ...
var CMComplex = model.CM{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.CM,
		Name:       "cm-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.CMData{
		Immutable: true,
		Items: []model.OpaqueData{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key2",
				Value: "value2\nvalue3\nvalue4",
			},
		},
	},
}

// SecretOpaque ...
var SecretOpaque = model.Secret{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Secret,
		Name:       "secret-opaque-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.SecretData{
		Type:      resCsts.SecretTypeOpaque,
		Immutable: true,
		Opaque: []model.OpaqueData{
			{"username", "admin_user"},
		},
	},
}

// SecretSocker ...
var SecretSocker = model.Secret{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Secret,
		Name:       "secret-docker-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.SecretData{
		Type: resCsts.SecretTypeDocker,
		Docker: model.DockerRegistryData{
			Registry: "docker.io",
			Username: "admin_user",
			Password: "......",
		},
	},
}

// SecretBasicAuth ...
var SecretBasicAuth = model.Secret{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Secret,
		Name:       "secret-basic-auth-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.SecretData{
		Type: resCsts.SecretTypeBasicAuth,
		BasicAuth: model.BasicAuthData{
			Username: "admin_user1",
		},
	},
}

// SecretSSHAuth ...
var SecretSSHAuth = model.Secret{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Secret,
		Name:       "secret-ssh-auth-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.SecretData{
		Type: resCsts.SecretTypeSSHAuth,
		SSHAuth: model.SSHAuthData{
			PublicKey:  "-----BEGIN RSA PUBLIC KEY-----\nB\n-----END RSA PUBLIC KEY-----\n",
			PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nA\n-----END RSA PRIVATE KEY-----\n",
		},
	},
}

// SecretTLS ...
var SecretTLS = model.Secret{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Secret,
		Name:       "secret-tls-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.SecretData{
		Type: resCsts.SecretTypeTLS,
		TLS: model.TLSData{
			PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nA\n-----END RSA PRIVATE KEY-----\n",
			Cert:       "-----BEGIN CERTIFICATE-----\nC\n-----END CERTIFICATE-----\n",
		},
	},
}

// SecretSAToken ...
var SecretSAToken = model.Secret{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Secret,
		Name:       "secret-sa-token-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Data: model.SecretData{
		Type: resCsts.SecretTypeSAToken,
		SAToken: model.SATokenData{
			Namespace: "default",
			SAName:    "default-x",
			Cert:      "-----BEGIN CERTIFICATE-----\nC\n-----END CERTIFICATE-----\n",
		},
	},
}
