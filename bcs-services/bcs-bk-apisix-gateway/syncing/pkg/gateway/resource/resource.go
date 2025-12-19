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

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/pkg/config"
)

const (
	// importResource import apisix gateway resource
	importResource = "/api/v1/open/gateways/%s/resources/-/import"
)

// Merger 资源合并器
type Merger struct {
	resourcesPath string
}

// NewMerger 创建资源合并器
func NewMerger(resourcesPath string) *Merger {
	return &Merger{
		resourcesPath: resourcesPath,
	}
}

// MergeResources 合并所有资源文件
func (rm *Merger) MergeResources() (map[string]ResourcesContent, error) {
	result := make(map[string]ResourcesContent)

	err := filepath.WalkDir(rm.resourcesPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 过滤掉 Kubernetes ConfigMap 产生的特殊文件
		if strings.HasPrefix(d.Name(), "..") {
			return nil
		}

		// 获取相对路径
		rel, err := filepath.Rel(rm.resourcesPath, path)
		if err != nil {
			return err
		}

		// 文件夹目录不符合的不处理（必须是子目录/文件的结构）
		filepathStr := strings.Split(rel, string(filepath.Separator))
		if len(filepathStr) != 2 {
			return nil
		}

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// 解析文件内容
		resource, err := rm.parseFileContent(content, d.Name())
		if err != nil {
			blog.Errorf("Failed to parse file %s: %v", path, err)
			return err
		}

		// 将资源添加到对应目录的数组中
		dirName := filepathStr[0]
		result[dirName] = append(result[dirName], resource...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	blog.Infof("MergeResources completed: %d directories processed", len(result))
	for dirName, resources := range result {
		blog.Infof("  %s: %d resources", dirName, len(resources))
	}

	return result, nil
}

// parseFileContent 解析文件内容（支持 YAML 和 JSON）
func (rm *Merger) parseFileContent(content []byte, filename string) (ResourcesContent, error) {
	var resource ResourcesContent

	ext := filepath.Ext(filename)
	switch ext {
	case ".yaml", ".yml":
		return rm.parseYAML(content)
	case ".json":
		return rm.parseJSON(content)
	default:
		blog.Warnf("Unsupported file extension: %s, skipping file: %s", ext, filename)
		return resource, nil
	}
}

// parseYAML 解析 YAML 文件（专门处理单个对象）
func (rm *Merger) parseYAML(content []byte) (ResourcesContent, error) {
	// 直接解析为单个对象
	var singleResource map[string]interface{}
	if err := yaml.Unmarshal(content, &singleResource); err != nil {
		blog.Errorf("YAML parsing failed: %s", err)
		blog.Errorf("YAML content preview: %s", string(content[:minInt(len(content), 200)]))
		return nil, err
	}

	// 验证对象合法性
	if err := rm.validateResourceObject(singleResource); err != nil {
		blog.Errorf("YAML validation failed: %s", err)
		blog.Errorf("YAML content: %s", string(content))
		return nil, err
	}

	return ResourcesContent{singleResource}, nil
}

// parseJSON 解析 JSON 文件（专门处理单个对象）
func (rm *Merger) parseJSON(content []byte) (ResourcesContent, error) {
	// 直接解析为单个对象
	var singleResource map[string]interface{}
	if err := json.Unmarshal(content, &singleResource); err != nil {
		blog.Errorf("JSON parsing failed: %s", err)
		blog.Errorf("JSON content preview: %s", string(content[:minInt(len(content), 200)]))
		return nil, err
	}

	// 验证对象合法性
	if err := rm.validateResourceObject(singleResource); err != nil {
		blog.Errorf("JSON validation failed: %s", err)
		blog.Errorf("JSON content: %s", string(content))
		return nil, err
	}

	return ResourcesContent{singleResource}, nil
}

// validateResourceObject 验证资源对象的合法性（只验证是正常的 YAML 对象）
func (rm *Merger) validateResourceObject(obj map[string]interface{}) error {
	// 只验证是一个非空的 map
	if len(obj) == 0 {
		return fmt.Errorf("YAML object is empty")
	}
	return nil
}

// minInt returns the smaller of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
	url := g.getResourcesUrl(fmt.Sprintf(importResource, g.syncConfig.GatewayConf.Name))
	_, err = component.HttpRequest(ctx, url, http.MethodPost, header, body)
	if err != nil {
		blog.Errorf("Failed to import resources: GatewayName=%s, URL=%s, Error=%v", g.syncConfig.GatewayConf.Name, url, err)
		return err
	}
	return nil
}

// ResourcesContent 资源内容
type ResourcesContent []map[string]interface{}

// mergeResource 合并资源
func (g *Resource) mergeResource() (io.Reader, string, error) {
	// 使用资源合并器
	merger := NewMerger(g.syncConfig.GatewayConf.ResourcesPath)
	result, err := merger.MergeResources()
	if err != nil {
		return nil, "", err
	}

	blog.Infof("Successfully merged resources from %d directories: %v", len(result), getKeys(result))

	// 打印合并后的内容用于调试
	g.printMergedContent(result)

	// 转成 JSON（二进制数据在内存中）
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", err
	}

	// 创建 multipart/form-data 请求体
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建一个"文件字段"，但数据来自内存
	fileField, err := writer.CreateFormFile("resource_file", "data.json")
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(fileField, bytes.NewReader(jsonBytes)) // 将内存数据写入"文件字段"
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

// getKeys 获取 map 的所有键（用于日志）
func getKeys(m map[string]ResourcesContent) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// printMergedContent 打印合并后的最终结果
func (g *Resource) printMergedContent(result map[string]ResourcesContent) {
	blog.Infof("=== Final Merged Resources Result ===")

	totalResources := 0
	for dirName, resources := range result {
		totalResources += len(resources)
		blog.Infof("Directory: %s (%d resources)", dirName, len(resources))
	}

	blog.Infof("Total resources merged: %d", totalResources)

	// 转换为 JSON 并打印最终结果
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		blog.Errorf("Failed to marshal final result: %v", err)
		return
	}

	blog.Infof("Final JSON result (length: %d bytes):", len(jsonBytes))
	blog.Infof("Final JSON result:\n%s", string(jsonBytes))
	blog.Infof("=== End Final Merged Resources Result ===")
}

func (g *Resource) getResourcesUrl(gatewayUrl string) string {
	return fmt.Sprintf("%s%s", g.syncConfig.GatewayConf.GatewayHost, gatewayUrl)
}
