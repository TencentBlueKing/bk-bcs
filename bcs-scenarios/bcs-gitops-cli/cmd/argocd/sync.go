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

package argocd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/internal/secretutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

var (
	secretStore *secretutils.Handler
)

func changeApplicationSync(c *cobra.Command) {
	originalFunc := c.Run
	c.Run = func(cmd *cobra.Command, args []string) {
		local, err := c.Flags().GetString("local")
		// directly use original function if not use local for sync
		if err != nil || local == "" {
			originalFunc(cmd, args)
			return
		}
		// directly use original function if not use vault-secret
		if !checkUseVaultSecret(local) {
			originalFunc(cmd, args)
			return
		}
		// build target dir
		localSplit := strings.Split(local, "/")
		targetDir := fmt.Sprintf(".gitops_synclocal_%s_%d", localSplit[len(localSplit)-1], time.Now().UnixMilli())
		color.Yellow(">>> copying dir '%s' to '%s'", local, targetDir)
		copyDir(local, targetDir)
		color.Yellow(">>> copy dir success")
		color.Yellow(">>> rendering vault-secret")
		secretStore = secretutils.NewHandler()
		renderSecret(targetDir)
		color.Yellow(">>> render vault-secret success")
		color.Yellow(fmt.Sprintf(">>> override param 'local' to '%s'", targetDir))
		c.Flags().Set("local", targetDir)
		originalFunc(cmd, args)
	}
}

func renderSecret(dir string) {
	secrets := make(map[string]string)
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		bs, err := os.ReadFile(path)
		// not care
		if err != nil {
			return nil
		}
		if !regx.Match(bs) {
			return nil
		}
		result := string(bs)
		matches := regx.FindAllString(result, -1)
		for _, match := range matches {
			if v, ok := secrets[match]; ok {
				result = strings.ReplaceAll(result, match, v)
				continue
			}
			secretValue, err := secretStore.GetSecret(match)
			if err != nil {
				utils.ExitError(fmt.Sprintf("get secret '%s' failed: %s", match, err.Error()))
			}
			result = strings.ReplaceAll(result, match, secretValue)
			color.Green(fmt.Sprintf("    rendering '%s' with '%s'", path, match))
		}
		file, err := os.Create(path)
		if err != nil {
			utils.ExitError(fmt.Sprintf("create file '%s' failed: %s", path, err.Error()))
		}
		defer file.Close()
		if _, err = io.WriteString(file, result); err != nil {
			utils.ExitError(fmt.Sprintf("write file '%s' failed: %s", path, err.Error()))
		}
		if err = file.Sync(); err != nil {
			utils.ExitError(fmt.Sprintf("sync file '%s' failed: %s", path, err.Error()))
		}
		return nil
	})
}

func copyDir(source, target string) {
	if err := os.MkdirAll(target, 0655); err != nil {
		utils.ExitError(fmt.Sprintf("mkdir '%s' failed", target))
	}
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(target, relPath)
		if info.IsDir() {
			err = os.MkdirAll(dstPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			err = copyFile(path, dstPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		utils.ExitError(fmt.Sprintf("copy dir '%s' to '%s' failed: %s", source, target, err.Error()))
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

var (
	regx = regexp.MustCompile(`<path:[^>]+>`)
)

func checkUseVaultSecret(dir string) bool {
	var result bool
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		bs, err := os.ReadFile(path)
		// not care
		if err != nil {
			return nil
		}
		if !regx.Match(bs) {
			return nil
		}
		result = true
		return errors.Errorf("finish")
	})
	return result
}
