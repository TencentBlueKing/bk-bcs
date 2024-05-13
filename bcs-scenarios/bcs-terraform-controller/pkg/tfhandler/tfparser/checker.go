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

// Package tfparser include terraform file parser
package tfparser

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

var (
	rootSchema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "terraform",
				LabelNames: nil,
			},
		},
	}

	terraformBlockSchema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "backend",
				LabelNames: []string{"name"},
			},
		},
	}
	// nolint
	specificPathPlaceholder, _ = regexp.Compile(`(?mU)<path:([^#]+)#([^#]+)(?:#([^#]+))?>`)
	// nolint
	indivPlaceholderSyntax, _ = regexp.
					Compile(`(?mU)path:(?P<project>[^/]+)/data/(?P<path>[^#]+?)#(?P<key>[^#]+?)(?:#(?P<version>.+?))??`)
)

// Interface 定义 Terraform 文件检查接口
type Interface interface {
	CheckBackendConsul() error
	RewriteSecret(ctx context.Context) error
}

type terraformParser struct {
	tfProject     string
	workerPath    string
	secretManager secret.SecretManagerWithVersion
}

// NewTerraformParser 创建 TerraformParser 实例
func NewTerraformParser(tfProject, workerPath string) Interface {
	return &terraformParser{
		tfProject:     tfProject,
		workerPath:    workerPath,
		secretManager: option.GetSecretManager(),
	}
}

// CheckBackendConsul 检查 tf 文件中 backend 是否是 consul. 用户必须显式设置 backend consul
// 才能满足规范
func (p *terraformParser) CheckBackendConsul() error {
	entries, err := os.ReadDir(p.workerPath)
	if err != nil {
		return errors.Wrapf(err, "read worker dir '%s' failed", p.workerPath)
	}
	hasRootEntry := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		var hasRoot bool
		var exist bool
		hasRoot, exist, err = p.existBackendConsul(path.Join(p.workerPath, entry.Name()))
		if err != nil {
			return errors.Wrapf(err, "check backend consul failed")
		}
		if hasRoot {
			hasRootEntry = true
		}
		if exist {
			return nil
		}
	}
	if hasRootEntry {
		return errors.Errorf(`must set terraform.backend with consul(just an empty consul backend),
		such as: backend "consul" {}`)
	}
	return errors.Errorf("current directory '%s' not found entrance .tf file, such as: terraform {}", p.workerPath)
}

// existBackendConsul 采用 hcl 解析 tf 文件，确认 consul backend 是否存在
func (p *terraformParser) existBackendConsul(filePath string) (bool, bool, error) {
	bs, err := os.ReadFile(filePath)
	if err != nil {
		return false, false, errors.Wrapf(err, `read file "%s" failed`, filePath)
	}
	parser := hclparse.NewParser()
	hclFile, diagnostics := parser.ParseHCL(bs, "")
	if len(diagnostics) != 0 {
		return false, false, errors.Errorf("parse file \"%s\" with hcl failed: %v", filePath, diagnostics.Error())
	}
	var hasRoot = false
	rootBC, _, diagnostics := hclFile.Body.PartialContent(rootSchema)
	if len(diagnostics) != 0 {
		return hasRoot, false, errors.Errorf("parse file \"%s\" root content failed: %v", filePath, diagnostics.Error())
	}
	hasRoot = true
	for _, rootBlock := range rootBC.Blocks {
		tfBC, _, diagnostics := rootBlock.Body.PartialContent(terraformBlockSchema)
		if len(diagnostics) != 0 {
			return hasRoot, false, errors.Errorf("parse file \"%s\" terraform block failed: %v", filePath, diagnostics.Error())
		}
		for _, backendBlock := range tfBC.Blocks {
			if backendBlock.Type != "backend" {
				continue
			}
			if len(backendBlock.Labels) != 1 {
				return hasRoot, false, errors.Errorf("parse file \"%s\" terraform.backend's labels \"%v\" not normal, "+
					"must be consul", filePath, backendBlock.Labels)
			}
			if backendBlock.Labels[0] != "consul" {
				return hasRoot, false, errors.Errorf(`terraform.backend must be consul, such as: backend "consul" {}`)
			}
			return hasRoot, true, nil
		}
	}
	return hasRoot, false, nil
}

// RewriteSecret 重写 tf 文件中的 secret 内容，逐个文件进行遍历，发现有 secret 格式则从 SecretManager 中
// 获取对应密钥信息，并进行重写
func (p *terraformParser) RewriteSecret(ctx context.Context) error {
	entries, err := os.ReadDir(p.workerPath)
	if err != nil {
		return errors.Wrapf(err, "read worker dir '%s' failed", p.workerPath)
	}
	var errs = make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		if err = p.rewriteSecretForFile(ctx, path.Join(p.workerPath, entry.Name())); err != nil {
			logctx.Warnf(ctx, "rewrite secret failed: %s", err.Error())
			errs = append(errs, err.Error())
		}
	}
	if len(errs) != 0 {
		return errors.Errorf("rewrite secret failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// rewriteSecretForFile 检测每个 tf 文件中是否满足对应表达式的 Secret，若存在则进行替换
func (p *terraformParser) rewriteSecretForFile(ctx context.Context, filePath string) error {
	logctx.Infof(ctx, "rewriting secret for file: %s", filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrapf(err, `read file "%s" failed`, filePath)
	}
	var errs = make([]string, 0)
	var placeholderRegex = specificPathPlaceholder
	result := placeholderRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		placeholder := strings.Trim(string(match), "<>")

		if !indivPlaceholderSyntax.Match([]byte(placeholder)) {
			errs = append(errs, fmt.Sprintf("secret '%s' incorrect format", string(match)))
			return match
		}
		secretMatches := indivPlaceholderSyntax.FindStringSubmatch(placeholder)
		vaultProject := secretMatches[indivPlaceholderSyntax.SubexpIndex("project")]
		vaultPath := secretMatches[indivPlaceholderSyntax.SubexpIndex("path")]
		vaultKey := secretMatches[indivPlaceholderSyntax.SubexpIndex("key")]
		version := secretMatches[indivPlaceholderSyntax.SubexpIndex("version")]
		logctx.Infof(ctx, "secret found for project '%s' with path '%s' key '%s' version '%s'",
			vaultProject, vaultPath, vaultKey, version)
		if vaultProject != p.tfProject {
			errs = append(errs, fmt.Sprintf("secret '%s' project not same as terraform project '%s'",
				string(match), p.tfProject))
			return match
		}

		var data map[string]interface{}
		data, err = p.secretManager.GetSecretWithVersion(ctx, &secret.SecretRequest{
			Project: vaultProject,
			Path:    vaultPath,
		}, utils.StringToInt(version))
		if err != nil {
			errs = append(errs, fmt.Sprintf("secret %s get from server failed: %s", string(match), err.Error()))
			return match
		}
		val, ok := data[vaultKey]
		if !ok {
			errs = append(errs, fmt.Sprintf("secret key '%s' not found", string(match)))
			return match
		}
		// provider 中需要双引号填充
		str := fmt.Sprintf("%s%v%s", "\"", val, "\"")
		return []byte(str)
	})
	if len(errs) != 0 {
		return errors.Errorf("rewrite secret for %s failed: %s", filePath, strings.Join(errs, ", "))
	}
	if err = os.WriteFile(filePath, result, 0644); err != nil {
		return errors.Wrapf(err, "rewrite secret for '%s' override failed", filePath)
	}
	logctx.Infof(ctx, "rewrite secret for file '%s' success", filePath)
	return nil
}
