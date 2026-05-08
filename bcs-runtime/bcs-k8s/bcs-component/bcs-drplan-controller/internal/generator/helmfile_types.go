/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generator

// HelmfileGenerateConfig holds CLI parameters for helmfile-based plan generation.
type HelmfileGenerateConfig struct {
	File           string
	Selectors      []string
	Namespace      string
	ChartRepo      string
	HookImage      string
	PlainHTTP      bool
	KeepFullValues bool
	OutputDir      string
}

// HelmfileLoadInput describes the inputs required to load and resolve one helmfile release.
type HelmfileLoadInput struct {
	File           string
	Selectors      []string
	Namespace      string
	ChartRepo      string
	HookImage      string
	PlainHTTP      bool
	KeepFullValues bool
}

// HelmfileResolvedHook is the normalized subset of release hook data used by the generator.
type HelmfileResolvedHook struct {
	Event   string
	Command string
	Args    []string
	Order   int
}

// HelmfileResolvedRelease is the normalized release model extracted from helmfile.
type HelmfileResolvedRelease struct {
	ReleaseName     string
	Namespace       string
	Chart           string
	ChartVersion    string
	ChartRepo       string
	HookImage       string
	TargetNamespace string
	ValuesYAML      string
	Hooks           []HelmfileResolvedHook
	Wait            *bool
	WaitForJob      *bool
	Atomic          *bool
	CreateNamespace *bool
	TimeoutSeconds  int32
	PlainHTTP       *bool
}
