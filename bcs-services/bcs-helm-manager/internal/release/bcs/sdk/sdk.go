/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sdk

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	rspb "helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	bcsAPIGWK8SBaseURI = "%s/clusters/%s/"
)

// Config 定义了使用sdk的基本参数
type Config struct {
	BcsAPI string
	Token  string

	PatchTemplates []*release.File
	VarTemplates   []*release.File
}

// NewGroup return a new Group instance
func NewGroup(c Config) Group {
	return &group{
		config: &c,
		groups: make(map[string]*client),
	}
}

// Group 定义了一组 Client
type Group interface {
	Cluster(clusterID string) Client
}

type group struct {
	config *Config

	sync.RWMutex
	groups map[string]*client
}

// Cluster 根据给定的Cluster, 返回一个能操作对应集群的Client
func (g *group) Cluster(clusterID string) Client {
	return g.getClient(clusterID)
}

func (g *group) getClient(clusterID string) Client {
	g.RLock()
	c, ok := g.groups[clusterID]
	g.RUnlock()
	if ok {
		return c
	}

	g.Lock()
	defer g.Unlock()
	c, ok = g.groups[clusterID]
	if !ok {
		flags := genericclioptions.NewConfigFlags(false)
		flags.APIServer = common.GetStringP(fmt.Sprintf(bcsAPIGWK8SBaseURI, g.config.BcsAPI, clusterID))
		flags.BearerToken = common.GetStringP(g.config.Token)
		flags.Insecure = common.GetBoolP(true)

		c = &client{
			group:     g,
			clusterID: clusterID,
			cf:        flags,
		}
		g.groups[clusterID] = c
	}

	return c
}

// Client 定义了支持的helm operation接口
type Client interface {
	List(ctx context.Context, namespace string) ([]*rspb.Release, error)
	Install(ctx context.Context, config release.HelmInstallConfig) (*release.HelmInstallResult, error)
	Upgrade(ctx context.Context, config release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error)
	Uninstall(ctx context.Context, config release.HelmUninstallConfig) (*release.HelmUninstallResult, error)
	Rollback(ctx context.Context, config release.HelmRollbackConfig) (*release.HelmRollbackResult, error)
}

type client struct {
	group *group

	clusterID string
	cf        *genericclioptions.ConfigFlags
}

// List helm release
func (c *client) List(_ context.Context, namespace string) ([]*rspb.Release, error) {
	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(namespace), namespace, "", blog.Infof); err != nil {
		return nil, err
	}

	return action.NewList(conf).Run()
}

// Install helm release through helm client
func (c *client) Install(_ context.Context, config release.HelmInstallConfig) (*release.HelmInstallResult, error) {
	blog.Infof("sdk client try install release name %s, namespace %s", config.Name, config.Namespace)

	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(config.Namespace), config.Namespace, "", blog.Infof); err != nil {
		blog.Errorf("sdk client install and init configuration failed, %s, %v", err.Error(), config)
		return nil, err
	}

	installer := action.NewInstall(conf)
	installer.DryRun = config.DryRun
	installer.ReleaseName = config.Name
	installer.Namespace = config.Namespace
	installer.PostRenderer = newPatcher(c.group.config.PatchTemplates, config.PatchTemplateValues)
	if err := parseArgs4Install(installer, config.Args); err != nil {
		blog.Errorf("sdk client install and parse from args failed, %s, args: %v", err.Error(), config.Args)
		return nil, err
	}

	// chart文件数据
	chartF, err := getChartFile(config.Chart)
	if err != nil {
		blog.Errorf("sdk client install and load chart files failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	// values数据, 增加Var values在最后
	values, err := getValues(append(config.Values, c.getVarValue(config.VarTemplateValues)...))
	if err != nil {
		blog.Errorf("sdk client install and get values failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	r, err := installer.Run(chartF, values)
	if err != nil {
		blog.Errorf("sdk client install failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	var status, appVersion, lastDeployed string
	if r.Info != nil {
		status = r.Info.Status.String()
		lastDeployed = r.Info.LastDeployed.Local().String()
	}
	if r.Chart != nil && r.Chart.Metadata != nil {
		appVersion = r.Chart.Metadata.AppVersion
	}
	blog.Infof("sdk client install release successfully name %s, namespace %s, revision: %d",
		config.Name, config.Namespace, r.Version)
	return &release.HelmInstallResult{
		Revision:   r.Version,
		Status:     status,
		AppVersion: appVersion,
		UpdateTime: lastDeployed,
	}, nil
}

// Upgrade helm release through helm client
func (c *client) Upgrade(_ context.Context, config release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error) {
	blog.Infof("sdk client try upgrade release name %s, namespace %s", config.Name, config.Namespace)

	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(config.Namespace), config.Namespace, "", blog.Infof); err != nil {
		blog.Errorf("sdk client upgrade and init configuration failed, %s, %v", err.Error(), config)
		return nil, err
	}

	upgrader := action.NewUpgrade(conf)
	upgrader.DryRun = config.DryRun
	upgrader.Namespace = config.Namespace
	upgrader.PostRenderer = newPatcher(c.group.config.PatchTemplates, config.PatchTemplateValues)
	if err := parseArgs4Upgrade(upgrader, config.Args); err != nil {
		blog.Errorf("sdk client upgrade and parse from args failed, %s, args: %v", err.Error(), config.Args)
		return nil, err
	}

	// chart文件数据
	chartF, err := getChartFile(config.Chart)
	if err != nil {
		blog.Errorf("sdk client upgrade and load chart files failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	// values数据, 增加Var values在最后
	values, err := getValues(append(config.Values, c.getVarValue(config.VarTemplateValues)...))
	if err != nil {
		blog.Errorf("sdk client upgrade and get values failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	r, err := upgrader.Run(config.Name, chartF, values)
	if err != nil {
		blog.Errorf("sdk client upgrade failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	var status, appVersion, lastDeployed string
	if r.Info != nil {
		status = r.Info.Status.String()
		lastDeployed = r.Info.LastDeployed.Local().String()
	}
	if r.Chart != nil && r.Chart.Metadata != nil {
		appVersion = r.Chart.Metadata.AppVersion
	}
	blog.Infof("sdk client upgrade release successfully name %s, namespace %s, revision: %d",
		config.Name, config.Namespace, r.Version)
	return &release.HelmUpgradeResult{
		Revision:   r.Version,
		Status:     status,
		AppVersion: appVersion,
		UpdateTime: lastDeployed,
	}, nil
}

// Uninstall helm release through helm client
func (c *client) Uninstall(_ context.Context, config release.HelmUninstallConfig) (
	*release.HelmUninstallResult, error) {

	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(config.Namespace), config.Namespace, "", blog.Infof); err != nil {
		blog.Errorf("sdk client uninstall and init configuration failed, %s, %v", err.Error(), config)
		return nil, err
	}

	uninstaller := action.NewUninstall(conf)
	uninstaller.DryRun = config.DryRun

	_, err := uninstaller.Run(config.Name)
	if err != nil {
		blog.Errorf("sdk client uninstall failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	return &release.HelmUninstallResult{}, nil
}

// Rollback helm release through helm client
func (c *client) Rollback(_ context.Context, config release.HelmRollbackConfig) (*release.HelmRollbackResult, error) {
	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(config.Namespace), config.Namespace, "", blog.Infof); err != nil {
		blog.Errorf("sdk client rollback and init configuration failed, %s, %v", err.Error(), config)
		return nil, err
	}

	rollbacker := action.NewRollback(conf)
	rollbacker.DryRun = config.DryRun
	rollbacker.Version = config.Revision

	if err := rollbacker.Run(config.Name); err != nil {
		blog.Errorf("sdk client rollback failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	return &release.HelmRollbackResult{}, nil
}

// getConfigFlag 获取helm-client配置
func (c *client) getConfigFlag(namespace string) *genericclioptions.ConfigFlags {
	flags := genericclioptions.NewConfigFlags(false)

	flags.APIServer = c.cf.APIServer
	flags.BearerToken = c.cf.BearerToken
	flags.Insecure = c.cf.Insecure
	flags.Namespace = common.GetStringP(namespace)
	return flags
}

// getVarValue 将给定的var数据渲染到模版中, 并转化为values文件
func (c *client) getVarValue(vars map[string]string) []*release.File {
	r := make([]*release.File, 0, 5)
	for index, f := range c.group.config.VarTemplates {
		r = append(r, &release.File{
			Name:    "vars-" + strconv.Itoa(index) + ".yaml",
			Content: replaceVarTplKey(vars, f.Content),
		})
	}

	return r
}

// getChartFile 从下载的tar包中获取chart数据
func getChartFile(f *release.File) (*chart.Chart, error) {
	bufferedFile, err := loader.LoadArchiveFiles(bytes.NewReader(f.Content))
	if err != nil {
		return nil, err
	}

	return loader.LoadFiles(bufferedFile)
}

// getValues 从给定的values文件中, 合并成最后给到helm的value数据
// values文件优先级按照命令行的原则, 后指定的优先级高
func getValues(fs []*release.File) (map[string]interface{}, error) {
	base := map[string]interface{}{}
	for _, value := range fs {
		blog.Infof("get values from %s: \n%s", value.Name, string(value.Content))
		currentMap := map[string]interface{}{}
		if err := yaml.Unmarshal(value.Content, &currentMap); err != nil {
			return nil, err
		}
		base = mergeMaps(base, currentMap)
	}

	return base, nil
}

// mergeMaps 合并两个values配置, b比a优先级高, 在b中指定的配置将会覆盖a中的相同配置
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// replaceVarTplKey 替换掉varTemplate中的模版变量, 用于作为最后的values文件, 让用户能够在chart中直接引用.Values.__BCS__相关配置
// 默认清除所有未渲染的模版变量, 置为空
func replaceVarTplKey(keys map[string]string, data []byte) []byte {
	for k, v := range keys {
		data = []byte(strings.ReplaceAll(string(data), common.Vtk(k), v))
	}

	return common.EmptyAllVarTemplateKey(data)
}

// parseArgs4Install 从用户给出的原生参数里, 直接parse到helm的install设置中
// 当前使用的flags设置版本来源helm.sh/helm/v3 v3.6.3
func parseArgs4Install(install *action.Install, args []string) error {
	f := pflag.NewFlagSet("install", pflag.ContinueOnError)

	f.BoolVar(&install.DisableHooks, "no-hooks", false,
		"prevent hooks from running during install")
	f.BoolVar(&install.Replace, "replace", false,
		"re-use the given name, only if that name is a deleted release which remains in the history. "+
			"This is unsafe in production")
	f.DurationVar(&install.Timeout, "timeout", 300*time.Second,
		"time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&install.Wait, "wait", false,
		"if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, "+
			"StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. "+
			"It will wait for as long as --timeout")
	f.BoolVar(&install.WaitForJobs, "wait-for-jobs", false,
		"if set and --wait enabled, will wait until all Jobs have been completed before marking the "+
			"release as successful. It will wait for as long as --timeout")
	f.StringVar(&install.NameTemplate, "name-template", "",
		"specify template used to name the release")
	f.StringVar(&install.Description, "description", "",
		"add a custom description")
	f.BoolVar(&install.Devel, "devel", false,
		"use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&install.DependencyUpdate, "dependency-update", false,
		"update dependencies if they are missing before installing the chart")
	f.BoolVar(&install.DisableOpenAPIValidation, "disable-openapi-validation", false,
		"if set, the installation process will not validate rendered templates against "+
			"the Kubernetes OpenAPI Schema")
	f.BoolVar(&install.Atomic, "atomic", false,
		"if set, the installation process deletes the installation on failure. "+
			"The --wait flag will be set automatically if --atomic is used")
	f.BoolVar(&install.SkipCRDs, "skip-crds", false,
		"if set, no CRDs will be installed. By default, CRDs are installed if not already present")
	f.BoolVar(&install.SubNotes, "render-subchart-notes", false,
		"if set, render subchart notes along with the parent")

	return f.Parse(args)
}

// parseArgs4Upgrade 从用户给出的原生参数里, 直接parse到helm的upgrade设置中
// 当前使用的flags设置版本来源helm.sh/helm/v3 v3.6.3
func parseArgs4Upgrade(upgrade *action.Upgrade, args []string) error {
	f := pflag.NewFlagSet("upgrade", pflag.ContinueOnError)

	f.BoolVarP(&upgrade.Install, "install", "i", false,
		"if a release by this name doesn't already exist, run an install")
	f.BoolVar(&upgrade.Devel, "devel", false,
		"use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&upgrade.Recreate, "recreate-pods", false,
		"performs pods restart for the resource if applicable")
	f.BoolVar(&upgrade.Force, "force", false,
		"force resource updates through a replacement strategy")
	f.BoolVar(&upgrade.DisableHooks, "no-hooks", false,
		"disable pre/post upgrade hooks")
	f.BoolVar(&upgrade.DisableOpenAPIValidation, "disable-openapi-validation", false,
		"if set, the upgrade process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	f.BoolVar(&upgrade.SkipCRDs, "skip-crds", false,
		"if set, no CRDs will be installed when an upgrade is performed with install flag enabled. "+
			"By default, CRDs are installed if not already present, "+
			"when an upgrade is performed with install flag enabled")
	f.DurationVar(&upgrade.Timeout, "timeout", 300*time.Second,
		"time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&upgrade.ResetValues, "reset-values", false,
		"when upgrading, reset the values to the ones built into the chart")
	f.BoolVar(&upgrade.ReuseValues, "reuse-values", false,
		"when upgrading, reuse the last release's values and merge in any overrides from the command line via "+
			"--set and -f. If '--reset-values' is specified, this is ignored")
	f.BoolVar(&upgrade.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, "+
		"and minimum number of Pods of a Deployment, StatefulSet, "+
		"or ReplicaSet are in a ready state before marking the release as successful. "+
		"It will wait for as long as --timeout")
	f.BoolVar(&upgrade.WaitForJobs, "wait-for-jobs", false,
		"if set and --wait enabled, will wait until all Jobs have been completed before marking "+
			"the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&upgrade.Atomic, "atomic", false,
		"if set, upgrade process rolls back changes made in case of failed upgrade. "+
			"The --wait flag will be set automatically if --atomic is used")
	f.BoolVar(&upgrade.CleanupOnFail, "cleanup-on-fail", false,
		"allow deletion of new resources created in this upgrade when upgrade fails")
	f.BoolVar(&upgrade.SubNotes, "render-subchart-notes", false,
		"if set, render subchart notes along with the parent")
	f.StringVar(&upgrade.Description, "description", "", "add a custom description")

	return f.Parse(args)
}
