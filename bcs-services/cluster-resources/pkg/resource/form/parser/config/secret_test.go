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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

func TestParseSecretDataOpaque(t *testing.T) {
	secretManifest := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				resCsts.EditModeAnnoKey: "form",
			},
			"name":      "secret-test",
			"namespace": "default",
		},
		"type":      resCsts.SecretTypeOpaque,
		"immutable": true,
		"data": map[string]interface{}{
			"username": "YWRtaW5fdXNlcg==",
		},
	}
	excepted := model.SecretData{
		Type:      resCsts.SecretTypeOpaque,
		Immutable: true,
		Opaque: []model.OpaqueData{
			{
				Key:   "username",
				Value: "admin_user",
			},
		},
	}
	actual := model.SecretData{}
	ParseSecretData(secretManifest, &actual)
	assert.Equal(t, excepted, actual)
}

func TestParseSecretDataDocker(t *testing.T) {
	// dockerconfigjson
	secretManifest := map[string]interface{}{
		"type": resCsts.SecretTypeDocker,
		"data": map[string]interface{}{
			".dockerconfigjson": "eyJhdXRocyI6eyJkb2NrZXIuaW8iOnsidXNlcm5hbWUiOiJhZG1pbl91c2VyIiwicGFzc3dvcmQiOiIifX19",
		},
	}
	excepted := model.SecretData{
		Type: resCsts.SecretTypeDocker,
		Docker: model.DockerRegistryData{
			Registry: "docker.io",
			Username: "admin_user",
			// not set value avoid code scan warning
			Password: "",
		},
	}
	actual := model.SecretData{}
	ParseSecretData(secretManifest, &actual)
	assert.Equal(t, excepted, actual)
}

func TestParseSecretDataBasicAuth(t *testing.T) {
	secretManifest := map[string]interface{}{
		"type": resCsts.SecretTypeBasicAuth,
		"data": map[string]interface{}{
			"username": "YWRtaW5fdXNlcjE=",
			"password": "",
		},
	}
	excepted := model.SecretData{
		Type: resCsts.SecretTypeBasicAuth,
		BasicAuth: model.BasicAuthData{
			Username: "admin_user1",
			Password: "",
		},
	}
	actual := model.SecretData{}
	ParseSecretData(secretManifest, &actual)
	assert.Equal(t, excepted, actual)
}

func TestParseSecretDataSSHAuth(t *testing.T) {
	secretManifest := map[string]interface{}{
		"type": resCsts.SecretTypeSSHAuth,
		"data": map[string]interface{}{
			"ssh-privatekey": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpBCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==",
			"ssh-publickey":  "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCkIKLS0tLS1FTkQgUlNBIFBVQkxJQyBLRVktLS0tLQo=",
		},
	}
	excepted := model.SecretData{
		Type: resCsts.SecretTypeSSHAuth,
		SSHAuth: model.SSHAuthData{
			PublicKey:  "-----BEGIN RSA PUBLIC KEY-----\nB\n-----END RSA PUBLIC KEY-----\n",
			PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nA\n-----END RSA PRIVATE KEY-----\n",
		},
	}
	actual := model.SecretData{}
	ParseSecretData(secretManifest, &actual)
	assert.Equal(t, excepted, actual)
}

func TestParseSecretDataTLS(t *testing.T) {
	secretManifest := map[string]interface{}{
		"type": resCsts.SecretTypeTLS,
		"data": map[string]interface{}{
			"tls.key": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpBCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==",
			"tls.crt": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCkMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
		},
	}
	excepted := model.SecretData{
		Type: resCsts.SecretTypeTLS,
		TLS: model.TLSData{
			PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nA\n-----END RSA PRIVATE KEY-----\n",
			Cert:       "-----BEGIN CERTIFICATE-----\nC\n-----END CERTIFICATE-----\n",
		},
	}
	actual := model.SecretData{}
	ParseSecretData(secretManifest, &actual)
	assert.Equal(t, excepted, actual)
}

func TestParseSecretDataSAToken(t *testing.T) {
	secretManifest := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"kubernetes.io/service-account.name": "default-x",
			},
		},
		"type": resCsts.SecretTypeSAToken,
		"data": map[string]interface{}{
			"namespace": "ZGVmYXVsdA==",
			"token":     "",
			"ca.crt":    "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCkMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
		},
	}
	excepted := model.SecretData{
		Type: resCsts.SecretTypeSAToken,
		SAToken: model.SATokenData{
			Namespace: "default",
			SAName:    "default-x",
			Token:     "",
			Cert:      "-----BEGIN CERTIFICATE-----\nC\n-----END CERTIFICATE-----\n",
		},
	}
	actual := model.SecretData{}
	ParseSecretData(secretManifest, &actual)
	assert.Equal(t, excepted, actual)
}
