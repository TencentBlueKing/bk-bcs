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

package controller

import (
	"context"
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/secretstore"
)

// Options for project controller
type Options struct {
	// global context for graceful exit
	Context context.Context
	// work mode, tunnel or service
	Mode      string
	ClientTLS *tls.Config
	// take affect when in service mode
	Registry    string
	RegistryTLS *tls.Config
	// 专门用于设置 cluster 的地址
	APIGatewayForCluster string
	// take effect when in tunnel mode
	APIGateway string
	APIToken   string
	// interval for data sync, seconds
	Interval uint
	// gitops system storage
	Storage store.Store
	Secret  secretstore.SecretInterface
}

// Controller common definition
type Controller interface {
	Init() error
	Start() error
	Stop()
}
