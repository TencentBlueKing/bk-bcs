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

// Package gateway xxx
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/types"
)

const (
	// createGateway create apisix gateway
	createGateway = "/api/v1/open/gateways"
	// updateGateway update apisix gateway
	updateGateway = "/api/v1/open/gateways/%s"
	// getGateway get apisix gateway by name
	getGateway = "/api/v1/open/gateways/%s"
	// publishGateway publish apisix gateway by name
	publishGateway = "/api/v1/open/gateways/%s/publish"
)

// Gateway xxxx
type Gateway struct {
	syncConfig *config.SyncConfig
}

// NewGateway xxx
func NewGateway(syncConfig *config.SyncConfig) *Gateway {
	return &Gateway{
		syncConfig: syncConfig,
	}
}

// GetGateway get apisix gateway by name
func (g *Gateway) GetGateway(ctx context.Context) error {
	header := http.Header{
		"X-BK-API-TOKEN": []string{g.syncConfig.GatewayConf.XBkApiToken},
	}
	_, err := component.HttpRequest(ctx,
		g.getGatewayUrl(fmt.Sprintf(getGateway, g.syncConfig.GatewayConf.Name)), http.MethodGet, header, nil)
	if err != nil {
		return err
	}
	return nil
}

// CreateGateway create apisix gateway
func (g *Gateway) CreateGateway(ctx context.Context) error {
	data := types.CreateGatewayReq{
		ApisixType:     g.syncConfig.GatewayConf.ApisixType,
		ApisixVersion:  g.syncConfig.GatewayConf.ApisixVersion,
		Description:    g.syncConfig.GatewayConf.Description,
		EtcdCaCert:     g.syncConfig.EtcdConf.EtcdCaCert,
		EtcdCertCert:   g.syncConfig.EtcdConf.EtcdCertCert,
		EtcdCertKey:    g.syncConfig.EtcdConf.EtcdCertKey,
		EtcdEndpoints:  g.syncConfig.EtcdConf.EtcdEndpoints,
		EtcdPassword:   g.syncConfig.EtcdConf.EtcdPassword,
		EtcdPrefix:     g.syncConfig.EtcdConf.EtcdPrefix,
		EtcdSchemaType: g.syncConfig.EtcdConf.EtcdSchemaType,
		EtcdUsername:   g.syncConfig.EtcdConf.EtcdUsername,
		Maintainers:    g.syncConfig.GatewayConf.Maintainers,
		Mode:           g.syncConfig.GatewayConf.Mode,
		Name:           g.syncConfig.GatewayConf.Name,
		ReadOnly:       g.syncConfig.GatewayConf.ReadOnly,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	header := http.Header{
		"X-BK-API-TOKEN": []string{g.syncConfig.GatewayConf.XBkApiToken},
		"Content-Type":   []string{"application/json"},
	}
	_, err = component.HttpRequest(ctx,
		g.getGatewayUrl(createGateway), http.MethodPost, header, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	return nil
}

// UpdateGateway update apisix gateway
func (g *Gateway) UpdateGateway(ctx context.Context) error {
	data := types.CreateGatewayReq{
		ApisixType:     g.syncConfig.GatewayConf.ApisixType,
		ApisixVersion:  g.syncConfig.GatewayConf.ApisixVersion,
		Description:    g.syncConfig.GatewayConf.Description,
		EtcdCaCert:     g.syncConfig.EtcdConf.EtcdCaCert,
		EtcdCertCert:   g.syncConfig.EtcdConf.EtcdCertCert,
		EtcdCertKey:    g.syncConfig.EtcdConf.EtcdCertKey,
		EtcdEndpoints:  g.syncConfig.EtcdConf.EtcdEndpoints,
		EtcdPassword:   g.syncConfig.EtcdConf.EtcdPassword,
		EtcdPrefix:     g.syncConfig.EtcdConf.EtcdPrefix,
		EtcdSchemaType: g.syncConfig.EtcdConf.EtcdSchemaType,
		EtcdUsername:   g.syncConfig.EtcdConf.EtcdUsername,
		Maintainers:    g.syncConfig.GatewayConf.Maintainers,
		Mode:           g.syncConfig.GatewayConf.Mode,
		Name:           g.syncConfig.GatewayConf.Name,
		ReadOnly:       g.syncConfig.GatewayConf.ReadOnly,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	header := http.Header{
		"X-BK-API-TOKEN": []string{g.syncConfig.GatewayConf.XBkApiToken},
		"Content-Type":   []string{"application/json"},
	}
	_, err = component.HttpRequest(ctx, g.getGatewayUrl(fmt.Sprintf(updateGateway, g.syncConfig.GatewayConf.Name)),
		http.MethodPut, header, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	return nil
}

// PublishGateway publish apisix gateway
func (g *Gateway) PublishGateway(ctx context.Context) error {
	header := http.Header{
		"X-BK-API-TOKEN": []string{g.syncConfig.GatewayConf.XBkApiToken},
		"Content-Type":   []string{"application/json"},
	}
	_, err := component.HttpRequest(ctx,
		g.getGatewayUrl(fmt.Sprintf(publishGateway, g.syncConfig.GatewayConf.Name)), http.MethodPost, header, nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *Gateway) getGatewayUrl(gatewayUrl string) string {
	return fmt.Sprintf("%s%s", g.syncConfig.GatewayConf.GatewayHost, gatewayUrl)
}
