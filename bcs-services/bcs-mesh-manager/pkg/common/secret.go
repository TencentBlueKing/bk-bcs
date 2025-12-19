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

// Package common 提供Remote Cluster Secret相关的YAML模板
package common

import (
	"fmt"
	"strings"
)

// SecretKind Secret资源类型
const SecretKind = "Secret"

// RemoteClusterSecretNamePrefix 远程集群Secret名称前缀
// #nosec G101 -- This is not a hardcoded credential, just a resource name prefix
const RemoteClusterSecretNamePrefix = "istio-remote-secret-"

// RemoteClusterSecretTemplate 远程集群Secret的YAML模板，用于多集群通信
// #nosec G101 -- This is not a hardcoded credential, just a resource name prefix
const RemoteClusterSecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  annotations:
    networking.istio.io/cluster: %s
  labels:
    created-by: bcs-mesh-manager
    istio/multiCluster: "true"
  name: %s
  namespace: istio-system
type: Opaque
stringData:
  %s: |
    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        server: '%s/clusters/%s/'
      name: '%s'
    contexts:
    - context:
        cluster: '%s'
        user: 'istio-remote'
      name: BCS
    current-context: BCS
    users:
    - name: 'istio-remote'
      user:
        token: '%s'`

// GetRemoteClusterSecretYAML 获取远程集群Secret的YAML
// clusterID: 远程集群ID
// bcsEndpoint: BCS API端点
// bcsToken: BCS认证token
func GetRemoteClusterSecretYAML(clusterID, bcsEndpoint, bcsToken string) string {
	lowerClusterID := strings.ToLower(clusterID)
	secretName := fmt.Sprintf("%s%s", RemoteClusterSecretNamePrefix, lowerClusterID)
	return fmt.Sprintf(RemoteClusterSecretTemplate,
		lowerClusterID, // networking.istio.io/cluster annotation
		secretName,     // secret name
		lowerClusterID, // stringData key
		bcsEndpoint,    // server URL
		clusterID,      // cluster ID in server URL
		clusterID,      // cluster name
		clusterID,      // cluster name in context
		bcsToken,       // auth token
	)
}
