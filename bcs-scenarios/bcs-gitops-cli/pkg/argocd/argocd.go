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

// Package argocd defines the argocd command
package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/httputils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// Handler defines the handler of terraform
type Handler struct {
	op *options.GitOpsOptions
}

// NewHandler create the terraform handler instance
func NewHandler() *Handler {
	return &Handler{
		op: options.GlobalOption(),
	}
}

var (
	dryRunPath   = "/api/v1/applications/dry-run"
	diffPath     = "/api/v1/applications/diff"
	generatePath = "/api/v1/applicationsets/generate"
)

// DryRun dry run application, it will show the result of every resource after dry-run
func (h *Handler) DryRun(ctx context.Context, req *argocd.ApplicationDryRunRequest, showDetails bool) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   dryRunPath,
		Method: http.MethodPost,
		Body:   req,
	})
	resp := new(argocd.ApplicationDryRunResponse)
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}
	if resp.Code != 0 {
		utils.ExitError(fmt.Sprintf("response code not 0: %s", resp.Message))
	}

	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"命名空间", "资源类型", "资源名称", "LIVE是否存在", "试运行结果", "错误信息",
		}
	}())
	for _, item := range resp.Data.Result {
		exist := fmt.Sprintf("%v", item.Existed)
		succeed := fmt.Sprintf("%v", item.IsSucceed)
		if !item.IsSucceed {
			succeed = color.RedString(succeed)
		}
		errMessage := item.ErrMessage
		if errMessage != "" {
			errMessage = color.RedString(errMessage)
		}
		tw.Append(func() []string {
			return []string{
				item.Namespace,
				fmt.Sprintf("%s/%s", item.Version, item.Kind),
				item.Name, exist, succeed, errMessage,
			}
		}())
	}
	tw.Render()
	if !showDetails {
		return
	}
	fmt.Println()
	for _, item := range resp.Data.Result {
		color.Blue("---")
		color.Blue("### Type: %s/%s, Name: %s\n", item.Version, item.Kind, item.Name)
		if item.Merged != nil {
			// nolint
			item.Merged.SetManagedFields(nil)
			bs, _ := yaml.Marshal(item.Merged)
			fmt.Printf("%s\n", string(bs))
		}
		if item.ErrMessage != "" {
			color.Red("# >>> dry-run failed: %s\n", item.ErrMessage)
		}
	}
}

// Diff the applications live state with target revision, and print the diff result
func (h *Handler) Diff(ctx context.Context, req *argocd.ApplicationDiffRequest, showDetails bool,
	showTarget, showLive bool) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   diffPath,
		Method: http.MethodPost,
		Body:   req,
	})
	resp := new(argocd.ApplicationDiffResponse)
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}
	if resp.Code != 0 {
		utils.ExitError(fmt.Sprintf("response code not 0: %s", resp.Message))
	}
	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"命名空间", "资源类型", "资源名称", "目标版本资源存在", "现网资源存在", "是否有变化",
		}
	}())
	for _, item := range resp.Data.Result {
		targetExist := false
		if item.Local != nil {
			targetExist = true
		}
		liveExist := false
		if item.Live != nil {
			liveExist = true
		}
		targetExistStr := fmt.Sprintf("%v", targetExist)
		liveExistStr := fmt.Sprintf("%v", liveExist)
		if targetExist != liveExist {
			targetExistStr = color.YellowString(targetExistStr)
			liveExistStr = color.YellowString(liveExistStr)
		}
		isChanged := !reflect.DeepEqual(item.Local, item.Live)
		changedStr := fmt.Sprintf("%v", isChanged)
		if isChanged {
			changedStr = color.RedString(changedStr)
		}
		tw.Append(func() []string {
			return []string{
				item.Namespace,
				fmt.Sprintf("%s/%s", item.Version, item.Kind),
				item.Name, targetExistStr, liveExistStr, changedStr,
			}
		}())
	}
	tw.Render()
	if showDetails {
		fmt.Println()
		color.Yellow("# left is local, right is live\n")
		dmp := diffmatchpatch.New()
		dmp.DiffTimeout = 0
		for _, item := range resp.Data.Result {
			if reflect.DeepEqual(item.Local, item.Live) {
				continue
			}
			color.Blue("---")
			color.Blue("### Type: %s/%s, Name: %s\n", item.Version, item.Kind, item.Name)
			bs1, _ := json.Marshal(item.Local)
			bs2, _ := json.Marshal(item.Live)
			src := string(utils.JsonToYaml(bs1))
			dst := string(utils.JsonToYaml(bs2))
			wSrc, wDst, warray := dmp.DiffLinesToRunes(src, dst)
			diffs := dmp.DiffMainRunes(wSrc, wDst, false)
			diffs = dmp.DiffCharsToLines(diffs, warray)
			for _, diff := range diffs {
				switch diff.Type {
				case diffmatchpatch.DiffInsert:
					text := strings.TrimSuffix(diff.Text, "\n")
					t := strings.Split(text, "\n")
					for i := range t {
						color.Green("+ %s", t[i])
					}
				case diffmatchpatch.DiffDelete:
					text := strings.TrimSuffix(diff.Text, "\n")
					t := strings.Split(text, "\n")
					for i := range t {
						color.Red("- %s", t[i])
					}
				case diffmatchpatch.DiffEqual:
					lines := strings.Split(diff.Text, "\n")
					for _, line := range lines {
						if len(line) > 0 {
							fmt.Printf("  %s\n", line)
						}
					}
				}
			}
		}
		return
	}
	if showTarget {
		fmt.Println()
		for _, item := range resp.Data.Result {
			if reflect.DeepEqual(item.Local, item.Live) {
				continue
			}
			color.Blue("---")
			color.Blue("### Type: %s/%s, Name: %s\n", item.Version, item.Kind, item.Name)
			bs, _ := json.Marshal(item.Local)
			transfer := string(utils.JsonToYaml(bs))
			fmt.Println(transfer)
			return
		}
	}
	if showLive {
		fmt.Println()
		for _, item := range resp.Data.Result {
			if reflect.DeepEqual(item.Local, item.Live) {
				continue
			}
			color.Blue("---")
			color.Blue("### Type: %s/%s, Name: %s\n", item.Version, item.Kind, item.Name)
			bs, _ := json.Marshal(item.Live)
			transfer := string(utils.JsonToYaml(bs))
			fmt.Println(transfer)
			return
		}
	}
}

// AppSetGenerate generate the applications by applicationset's manifest. It will print all the
// result that generate
func (h *Handler) AppSetGenerate(ctx context.Context, appset *v1alpha1.ApplicationSet, appSetShowDetails bool) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   generatePath,
		Method: http.MethodPost,
		Body:   appset,
	})
	resp := new(argocd.ApplicationSetGenerateResponse)
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}
	if resp.Code != 0 {
		utils.ExitError(fmt.Sprintf("response code not 0: %s", resp.Message))
	}

	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"应用名称", "目标集群", "目标命名空间", "仓库地址", "目标版本", "VALUE文件",
		}
	}())
	for _, app := range resp.Data {
		cserver := strings.Split(app.Spec.Destination.Server, "/")
		cluster := cserver[len(cserver)-1]
		var repo string
		var revision string
		var valueFile string
		if app.Spec.Source != nil {
			repo = app.Spec.Source.RepoURL
			revision = app.Spec.Source.TargetRevision
			if app.Spec.Source.Helm != nil {
				valueFile = strings.Join(app.Spec.Source.Helm.ValueFiles, "\n")
			}
		} else {
			repos := make([]string, 0)
			revisions := make([]string, 0)
			values := make([]string, 0)
			for i := range app.Spec.Sources {
				src := app.Spec.Sources[i]
				repos = append(repos, src.RepoURL)
				revisions = append(revisions, src.TargetRevision)
				if src.Helm != nil {
					values = append(values, src.Helm.ValueFiles...)
				}
			}
			repo = strings.Join(repos, "\n")
			revision = strings.Join(revisions, "\n")
			valueFile = strings.Join(values, "\n")
		}
		tw.Append([]string{
			app.Name, cluster, app.Spec.Destination.Namespace,
			repo, revision, valueFile,
		})
	}
	tw.Render()
	if !appSetShowDetails {
		return
	}
	fmt.Println()
	for _, app := range resp.Data {
		color.Blue("---")
		color.Blue("### Name: %s\n", app.Name)
		bs, _ := json.Marshal(app)
		fmt.Println(string(utils.JsonToYaml(bs)))
	}
}
