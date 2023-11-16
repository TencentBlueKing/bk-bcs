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
package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"

	"bscp.io/cmd/vault-server/vault-plugins/service"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	logger := hclog.New(&hclog.LoggerOptions{Name: "bcs-bscp-vault-plugin"})
	flags := apiClientMeta.FlagSet()
	if err := flags.Parse(os.Args[1:]); err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: service.Factory,
		TLSProviderFunc:    tlsProviderFunc,
		Logger:             logger,
	})
	if err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
