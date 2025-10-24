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

// Package resource xxx
package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/config"
)

const (
	// importResource import apisix gateway resource
	importResource = "/api/v1/open/gateways/%s/resources/-/import"
)

// Resource xxxx
type Resource struct {
	syncConfig *config.SyncConfig
}

// NewResource xxx
func NewResource(syncConfig *config.SyncConfig) *Resource {
	return &Resource{
		syncConfig: syncConfig,
	}
}

// ImportResource import apisix gateway resource
func (g *Resource) ImportResource(ctx context.Context) error {
	body, contentType, err := g.mergeResource()
	if err != nil {
		return err
	}
	header := http.Header{
		"X-BK-API-TOKEN": []string{g.syncConfig.GatewayConf.XBkApiToken},
		"Content-Type":   []string{contentType},
	}
	_, err = component.HttpRequest(ctx,
		g.getResourcesUrl(fmt.Sprintf(importResource, g.syncConfig.GatewayConf.Name)), http.MethodPost, header, body)
	if err != nil {
		return err
	}
	return nil
}

// ResourcesContent 资源内容
type ResourcesContent []map[string]interface{}

// mergeResource 合并资源
func (g *Resource) mergeResource() (io.Reader, string, error) {
	result := make(map[string]ResourcesContent)
	err := filepath.WalkDir(g.syncConfig.GatewayConf.ResourcesPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 只处理子目录文件
		if d.IsDir() {
			return nil
		}
		// 获取相对路径
		rel, err := filepath.Rel(g.syncConfig.GatewayConf.ResourcesPath, path)
		if err != nil {
			return err
		}
		// 文件夹目录不符合的不处理
		filepathStr := strings.Split(rel, string(filepath.Separator))
		if len(filepathStr) != 2 {
			return nil
		}
		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		resource := make(ResourcesContent, 0)
		if filepath.Ext(d.Name()) == ".yaml" || filepath.Ext(d.Name()) == ".yml" {
			if err := yaml.Unmarshal(content, &resource); err != nil {
				blog.Errorf("YAML parsing failed: %s", err)
				return err
			}
		}

		if filepath.Ext(d.Name()) == ".json" {
			if err := json.Unmarshal(content, &resource); err != nil {
				blog.Errorf("JSON parsing failed: %s", err)
				return err
			}
		}

		result[filepathStr[0]] = append(result[filepathStr[0]], resource...)

		return nil
	})
	if err != nil {
		return nil, "", err
	}

	// 转成 JSON（二进制数据在内存中）
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", err
	}

	// 创建 multipart/form-data 请求体
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建一个“文件字段”，但数据来自内存
	fileField, err := writer.CreateFormFile("resource_file", "data.json")
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(fileField, bytes.NewReader(jsonBytes)) // 将内存数据写入“文件字段”
	if err != nil {
		return nil, "", err
	}
	contentType := writer.FormDataContentType()
	err = writer.Close()
	if err != nil {
		return nil, "", err
	}
	return body, contentType, nil
}

func (g *Resource) getResourcesUrl(gatewayUrl string) string {
	return fmt.Sprintf("%s%s", g.syncConfig.GatewayConf.GatewayHost, gatewayUrl)
}
