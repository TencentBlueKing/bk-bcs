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
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

// var genericPlaceholder, _ = regexp.Compile(`(?mU)<(.*)>`)
var specificPathPlaceholder, _ = regexp.Compile(`(?mU)<path:([^#]+)#([^#]+)(?:#([^#]+))?>`)
var indivPlaceholderSyntax, _ = regexp.Compile(`(?mU)path:(?P<project>[^/]+)/data/(?P<path>[^#]+?)#(?P<key>[^#]+?)(?:#(?P<version>.+?))??`)

// GenerateSecretForTF 通过GitOps密钥管理接口渲染provider声明的vault格式AKSK
func (t *terraformLocalRunner) GenerateSecretForTF(ctx context.Context, workdir string) error {
	files, err := os.ReadDir(workdir)
	if err != nil {
		return errors.Wrapf(err, "read dir error, dir: %s", workdir)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".tf") || file.IsDir() {
			continue
		}
		path := fmt.Sprintf("%s/%s", workdir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "read file error, file: %s", path)
		}
		blog.Info("Processed file: %s \n", path)

		var placeholderRegex = specificPathPlaceholder
		res := placeholderRegex.ReplaceAllFunc(content, func(match []byte) []byte {
			placeholder := strings.Trim(string(match), "<>")

			if indivPlaceholderSyntax.Match([]byte(placeholder)) {
				indivSecretMatches := indivPlaceholderSyntax.FindStringSubmatch(placeholder)
				vaultProejct := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("project")]
				vaultPath := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("path")]
				vaultKey := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("key")]
				version := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("version")]

				secretVal, err := t.getSecret(ctx, vaultProejct, vaultPath, vaultKey, version)
				if err != nil {
					blog.Errorf("found placeholder `%s` but get secret err: %s \n", placeholder, err.Error())
					return match
				}
				blog.Info("found placeholder `%s` and replace to `%s` \n", placeholder, string(secretVal))
				return secretVal
			} else {
				blog.Errorf("not match placeholder: `%s` \n", placeholder)
			}

			return match
		})

		if config, err := t.forceUpdateBackendConfig(res); err != nil { // res set bcs backend
			blog.Errorf("update backend config failed, err: %s", err.Error())
		} else if len(config) != 0 { // 覆盖
			res = config
		}

		if err := os.WriteFile(path, res, 0644); err != nil {
			continue
		}
	}

	return nil
}

// 通过gitops secret接口获取vault存储的密钥
func (t *terraformLocalRunner) getSecret(ctx context.Context, project, path, key, version string) ([]byte, error) {
	if t.terraform.Spec.Project != project {
		return nil, errors.New("secret.project not equal terraform.spec.project")
	}

	if err := t.secret.Init(); err != nil {
		return nil, errors.Wrapf(err, "secret init error for path `%s` and key `%s`", path, key)
	}
	req := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}

	data, err := t.secret.GetSecretWithVersion(ctx, req, utils.StringToInt(version))
	if err != nil {
		return nil, errors.Wrapf(err, "get secret error for path `%s` and key `%s`", path, key)
	}

	val, ok := data[key]
	if !ok {
		return nil, errors.Wrapf(err, "get secret key not found for path `%s` and key `%s`", path, key)
	}
	// provider中需要双引号填充
	str := fmt.Sprintf("%s%v%s", "\"", val, "\"")

	return []byte(str), nil
}
