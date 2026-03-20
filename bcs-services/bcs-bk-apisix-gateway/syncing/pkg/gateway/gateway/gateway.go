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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/pkg/config"
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
	url := g.getGatewayUrl(fmt.Sprintf(getGateway, g.syncConfig.GatewayConf.Name))
	_, err := component.HttpRequest(ctx, url, http.MethodGet, header, nil)
	if err != nil {
		blog.Errorf("Failed to get gateway: GatewayName=%s, URL=%s, Error=%v", g.syncConfig.GatewayConf.Name, url, err)
		return err
	}
	return nil
}

// CreateGateway create apisix gateway
func (g *Gateway) CreateGateway(ctx context.Context) error {
	// 获取证书内容（支持从文件路径读取或直接使用配置的内容）
	caCert, err := g.syncConfig.EtcdConf.GetEtcdCaCert()
	if err != nil {
		return err
	}
	certCert, err := g.syncConfig.EtcdConf.GetEtcdCertCert()
	if err != nil {
		return err
	}
	certKey, err := g.syncConfig.EtcdConf.GetEtcdCertKey()
	if err != nil {
		return err
	}

	data := types.CreateGatewayReq{
		ApisixType:     g.syncConfig.GatewayConf.ApisixType,
		ApisixVersion:  g.syncConfig.GatewayConf.ApisixVersion,
		Description:    g.syncConfig.GatewayConf.Description,
		EtcdCaCert:     caCert,
		EtcdCertCert:   certCert,
		EtcdCertKey:    certKey,
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
	url := g.getGatewayUrl(createGateway)
	_, err = component.HttpRequest(ctx, url, http.MethodPost, header, bytes.NewReader(jsonData))
	if err != nil {
		blog.Errorf("Failed to create gateway: GatewayName=%s, URL=%s, Error=%v", g.syncConfig.GatewayConf.Name, url, err)
		return err
	}
	return nil
}

// UpdateGateway update apisix gateway
func (g *Gateway) UpdateGateway(ctx context.Context) error {
	// 获取证书内容（支持从文件路径读取或直接使用配置的内容）
	caCert, err := g.syncConfig.EtcdConf.GetEtcdCaCert()
	if err != nil {
		return err
	}
	certCert, err := g.syncConfig.EtcdConf.GetEtcdCertCert()
	if err != nil {
		return err
	}
	certKey, err := g.syncConfig.EtcdConf.GetEtcdCertKey()
	if err != nil {
		return err
	}

	data := types.CreateGatewayReq{
		ApisixType:     g.syncConfig.GatewayConf.ApisixType,
		ApisixVersion:  g.syncConfig.GatewayConf.ApisixVersion,
		Description:    g.syncConfig.GatewayConf.Description,
		EtcdCaCert:     caCert,
		EtcdCertCert:   certCert,
		EtcdCertKey:    certKey,
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
	url := g.getGatewayUrl(fmt.Sprintf(updateGateway, g.syncConfig.GatewayConf.Name))
	_, err = component.HttpRequest(ctx, url, http.MethodPut, header, bytes.NewReader(jsonData))
	if err != nil {
		blog.Errorf("Failed to update gateway: GatewayName=%s, URL=%s, Error=%v", g.syncConfig.GatewayConf.Name, url, err)
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
	url := g.getGatewayUrl(fmt.Sprintf(publishGateway, g.syncConfig.GatewayConf.Name))
	_, err := component.HttpRequest(ctx, url, http.MethodPost, header, nil)
	if err != nil {
		blog.Errorf("Failed to publish gateway: GatewayName=%s, URL=%s, Error=%v", g.syncConfig.GatewayConf.Name, url, err)
		return err
	}
	return nil
}

func (g *Gateway) getGatewayUrl(gatewayUrl string) string {
	// 去掉 gatewayHost 的末尾的 /
	gatewayHost := strings.TrimSuffix(g.syncConfig.GatewayConf.GatewayHost, "/")
	// 去掉 gatewayUrl 的开头的 /，增加兼容性
	gatewayUrl = strings.TrimPrefix(gatewayUrl, "/")
	return fmt.Sprintf("%s/%s", gatewayHost, gatewayUrl)
}
