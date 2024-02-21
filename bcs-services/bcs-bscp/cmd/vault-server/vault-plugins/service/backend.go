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

// Package service encapsulates Vault secret plugins and provides key-value storage and PKI storage functionality.
package service

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const backendHelp = `
bcs-bscp-vault-plugin provides secure kv storage, key storage;
Support multiple encryption algorithms for secure credential acquisition
`

// Factory factory for backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(ctx, conf)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend
}

// Backend new backend
// nolint
func Backend(ctx context.Context, conf *logical.BackendConfig) *backend {

	b := &backend{}
	b.Backend = &framework.Backend{
		Help: backendHelp,
		Paths: []*framework.Path{
			b.pathKvs(),
			b.pathKeys(),
			b.pathKvEncrypt(),
		},
		Secrets:     []*framework.Secret{},
		BackendType: logical.TypeLogical,
	}

	return b
}
