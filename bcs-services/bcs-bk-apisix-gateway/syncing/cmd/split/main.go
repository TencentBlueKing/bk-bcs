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

// Package main provides a command-line tool for splitting merged JSON resources into individual YAML files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResourceSplitter 资源拆分器
type ResourceSplitter struct {
	inputFile string
	outputDir string
	overwrite bool
}

// NewResourceSplitter 创建资源拆分器
func NewResourceSplitter(inputFile, outputDir string, overwrite bool) *ResourceSplitter {
	return &ResourceSplitter{
		inputFile: inputFile,
		outputDir: outputDir,
		overwrite: overwrite,
	}
}

// SplitResources 拆分资源文件
func (rs *ResourceSplitter) SplitResources() error {
	// 读取输入文件
	data, err := os.ReadFile(rs.inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	// 解析 JSON
	var resources map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &resources); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	fmt.Printf("Found %d resource types in input file\n", len(resources))

	// 创建输出目录
	if err := os.MkdirAll(rs.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 拆分每个资源类型
	for resourceType, resourceList := range resources {
		fmt.Printf("Processing %s: %d resources\n", resourceType, len(resourceList))

		// 创建子目录
		typeDir := filepath.Join(rs.outputDir, resourceType)
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", typeDir, err)
		}

		// 拆分每个资源
		for i, resource := range resourceList {
			if err := rs.writeResourceFile(typeDir, resource, i); err != nil {
				return fmt.Errorf("failed to write resource file: %v", err)
			}
		}
	}

	fmt.Printf("Successfully split resources to %s\n", rs.outputDir)
	return nil
}

// writeResourceFile 写入单个资源文件
func (rs *ResourceSplitter) writeResourceFile(dir string, resource map[string]interface{}, index int) error {
	// 生成文件名
	filename := rs.generateFilename(resource, index)
	filepath := filepath.Join(dir, filename)

	// 检查文件是否已存在
	if !rs.overwrite {
		if _, err := os.Stat(filepath); err == nil {
			fmt.Printf("  Skipping %s (already exists, use -overwrite to replace)\n", filename)
			return nil
		}
	}

	// 转换为 YAML
	yamlData, err := yaml.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filepath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filepath, err)
	}

	fmt.Printf("  Created %s\n", filename)
	return nil
}

// generateFilename 生成文件名
func (rs *ResourceSplitter) generateFilename(resource map[string]interface{}, index int) string {
	// 尝试从资源中获取名称信息
	var name string

	// 优先使用 name 字段
	if nameVal, ok := resource["name"].(string); ok && nameVal != "" {
		name = nameVal
	} else if idVal, ok := resource["id"].(string); ok && idVal != "" {
		// 使用 id 字段
		name = idVal
	} else if resourceIDVal, ok := resource["resource_id"].(string); ok && resourceIDVal != "" {
		// 使用 resource_id 字段
		name = resourceIDVal
	} else {
		// 使用索引
		name = fmt.Sprintf("resource-%d", index+1)
	}

	// 清理文件名（移除特殊字符）
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")

	// 确保文件名不为空
	if name == "" {
		name = fmt.Sprintf("resource-%d", index+1)
	}

	return fmt.Sprintf("%s.yaml", name)
}

func main() {
	var (
		inputFile = flag.String("input", "", "Input JSON file path (required)")
		outputDir = flag.String("output", "./split-resources", "Output directory path")
		overwrite = flag.Bool("overwrite", false, "Overwrite existing files")
		help      = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		fmt.Println("Resource Splitter - Split merged JSON resources into individual YAML files")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run main.go -input <json-file> [-output <output-dir>] [-overwrite]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -input string")
		fmt.Println("        Input JSON file path (required)")
		fmt.Println("  -output string")
		fmt.Println("        Output directory path (default: ./split-resources)")
		fmt.Println("  -overwrite")
		fmt.Println("        Overwrite existing files (default: false)")
		fmt.Println("  -help")
		fmt.Println("        Show this help information")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go -input merged-resources.json")
		fmt.Println("  go run main.go -input merged-resources.json -output ./apisix-config")
		fmt.Println("  go run main.go -input merged-resources.json -output ./apisix-config -overwrite")
		return
	}

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input parameter is required\n")
		fmt.Fprintf(os.Stderr, "Use -help for more information\n")
		os.Exit(1)
	}

	// 检查输入文件是否存在
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input file %s does not exist\n", *inputFile)
		os.Exit(1)
	}

	// 创建拆分器
	splitter := NewResourceSplitter(*inputFile, *outputDir, *overwrite)

	// 执行拆分
	if err := splitter.SplitResources(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
