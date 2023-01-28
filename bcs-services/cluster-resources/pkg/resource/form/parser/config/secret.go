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
	"encoding/base64"
	"encoding/json"

	"github.com/fatih/structs"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseSecret ConfigMap manifest -> formData
func ParseSecret(manifest map[string]interface{}) map[string]interface{} {
	secret := model.Secret{}
	common.ParseMetadata(manifest, &secret.Metadata)
	ParseSecretData(manifest, &secret.Data)
	return structs.Map(secret)
}

// ParseSecretData ...
func ParseSecretData(manifest map[string]interface{}, spec *model.SecretData) {
	spec.Type = mapx.Get(manifest, "type", resCsts.SecretTypeOpaque).(string)
	spec.Immutable = mapx.GetBool(manifest, "immutable")
	switch spec.Type {
	case resCsts.SecretTypeOpaque:
		for k, v := range mapx.GetMap(manifest, "data") {
			val, _ := base64.StdEncoding.DecodeString(v.(string))
			spec.Opaque = append(spec.Opaque, model.OpaqueData{
				Key: k, Value: string(val),
			})
		}
	case resCsts.SecretTypeDocker:
		dockerconfigjson, _ := base64.StdEncoding.DecodeString(
			mapx.GetStr(manifest, []string{"data", ".dockerconfigjson"}),
		)
		dockerConf := map[string]interface{}{}
		_ = json.Unmarshal(dockerconfigjson, &dockerConf)
		for reg, conf := range mapx.GetMap(dockerConf, "auths") {
			spec.Docker = model.DockerRegistryData{
				Registry: reg,
				Username: mapx.GetStr(conf.(map[string]interface{}), "username"),
				Password: mapx.GetStr(conf.(map[string]interface{}), "password"),
			}
			// 只取首个 docker config 数据
			break
		}
	case resCsts.SecretTypeBasicAuth:
		username, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, "data.username"))
		password, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, "data.password"))
		spec.BasicAuth = model.BasicAuthData{
			Username: string(username), Password: string(password),
		}
	case resCsts.SecretTypeSSHAuth:
		publicKey, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, "data.ssh-publickey"))
		privateKey, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, "data.ssh-privatekey"))
		spec.SSHAuth = model.SSHAuthData{
			PublicKey: string(publicKey), PrivateKey: string(privateKey),
		}
	case resCsts.SecretTypeTLS:
		privateKey, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, []string{"data", "tls.key"}))
		cert, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, []string{"data", "tls.crt"}))
		spec.TLS = model.TLSData{
			PrivateKey: string(privateKey), Cert: string(cert),
		}
	case resCsts.SecretTypeSAToken:
		saName := mapx.GetStr(manifest, []string{"metadata", "annotations", "kubernetes.io/service-account.name"})
		ns, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, "data.namespace"))
		token, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, "data.token"))
		cert, _ := base64.StdEncoding.DecodeString(mapx.GetStr(manifest, []string{"data", "ca.crt"}))
		spec.SAToken = model.SATokenData{
			Namespace: string(ns), SAName: saName, Token: string(token), Cert: string(cert),
		}
	}
}
