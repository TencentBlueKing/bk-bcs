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
 *
 */

package runner

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pkg/errors"
)

const (
	// terraformCodeSnippet terraform代码段
	terraformCodeSnippet = "terraform"

	// backendCodeSnippet backend代码段
	backendCodeSnippet = "backend"

	// bcsBackendConfig bcs backend is consul
	bcsBackendConfig = "\n  backend \"consul\" {}\n"
)

// forceUpdateBackendConfig 强制更新用户的backend配置，必须为`  backend "consul" {}\n`
func (t *terraformLocalRunner) forceUpdateBackendConfig(content []byte) ([]byte, error) {
	if len(content) == 0 {
		return nil, nil
	}
	// 创建一个HCL解析器
	parser := hclparse.NewParser()
	// 解析文件内容
	file, diags := parser.ParseHCL(content, fmt.Sprintf("%s-example.tf", t.instanceID)) // 名称任意即可
	if diags.HasErrors() {
		return nil, errors.Errorf("failed to parse HCL: %s\n", diags.Error())
	}
	// 类型转换
	body, err := typeConversion(file.Body)
	if err != nil {
		return nil, err
	}
	// 查找terraform代码段
	terraform := searchTerraform(body.Blocks, terraformCodeSnippet)
	if terraform == nil {
		return nil, nil
	}
	// 结果
	res := bytes.NewBuffer(nil)
	// 查找backend代码段
	backend := searchTerraform(terraform.Body.Blocks, backendCodeSnippet)
	if backend != nil { // 有backend，需要替换
		flag := backend.Range()
		// set start
		res.Write(content[:flag.Start.Byte-1])
		// add back end
		res.WriteString(bcsBackendConfig)
		// add end
		res.Write(content[flag.End.Byte+1:])

	} else { // 没有backend,需要设置
		flag := terraform.Range()
		// set start
		res.Write(content[:flag.End.Byte-1])
		// add back end
		res.WriteString(bcsBackendConfig)
		// add end
		res.Write(content[flag.End.Byte-1:])
	}

	return res.Bytes(), nil
}

// searchTerraform 搜索terraform代码
func searchTerraform(blocks hclsyntax.Blocks, text string) *hclsyntax.Block {
	for _, block := range blocks {
		if block.Type == text {
			return block
		}
	}

	return nil
}

// typeConversion 检查类型是否为预期类型(*hclsyntax.Body)
func typeConversion(obj interface{}) (*hclsyntax.Body, error) {
	switch obj.(type) {
	case *hclsyntax.Body:

	default:
		value := reflect.TypeOf(obj)
		return nil, errors.Errorf("unable to perform type conversion, unknown type: %s", value.String())
	}

	body, ok := obj.(*hclsyntax.Body)
	if !ok || body == nil {
		return nil, errors.Errorf("body conver failed")
	}

	return body, nil
}
