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

// Package sdk xxx
package sdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/spf13/pflag"
	yaml3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	projectClient "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
)

const (
	bcsAPIGWK8SBaseURI = "%s/clusters/%s/"
	defaultMaxHistory  = 10
	defaultTimeout     = "15s"
)

// bcs var in values regex, eg: {{ .BCS_SYS_CLUSTER_ID }}
var bcsVarRegex = regexp.MustCompile(`{{\s*\.BCS_(\w+)\s*}}`)

// Config 定义了使用sdk的基本参数
type Config struct {
	PatchTemplates []*release.File
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
	Config(clusterID string) *genericclioptions.ConfigFlags
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

// Config return cluster config
func (g *group) Config(clusterID string) *genericclioptions.ConfigFlags {
	g.getClient(clusterID)
	g.RLock()
	defer g.RUnlock()
	return g.groups[clusterID].cf
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
		apiserver := options.GlobalOptions.Release.APIServer
		flags.APIServer = common.GetStringP(fmt.Sprintf(bcsAPIGWK8SBaseURI, apiserver, clusterID))
		flags.BearerToken = common.GetStringP(options.GlobalOptions.Release.Token)
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
	Get(ctx context.Context, namespace, name string, revision int) (*rspb.Release, error)
	List(ctx context.Context, option release.ListOption) ([]*rspb.Release, error)
	Install(ctx context.Context, config release.HelmInstallConfig) (*release.HelmInstallResult, error)
	Upgrade(ctx context.Context, config release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error)
	Uninstall(ctx context.Context, config release.HelmUninstallConfig) (*release.HelmUninstallResult, error)
	Rollback(ctx context.Context, config release.HelmRollbackConfig) (*release.HelmRollbackResult, error)
	History(ctx context.Context, namespace, name string, max int) ([]*rspb.Release, error)
}

type client struct {
	group *group

	clusterID string
	cf        *genericclioptions.ConfigFlags
}

// Get helm release
func (c *client) Get(_ context.Context, namespace, name string, revision int) (*rspb.Release, error) {
	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(namespace), namespace, "", blog.Infof); err != nil {
		return nil, err
	}

	get := action.NewGet(conf)
	get.Version = revision
	re, err := get.Run(name)
	if err != nil {
		return nil, err
	}
	re.Config = removeValuesTemplate(re.Config)
	return re, nil
}

// List helm release
func (c *client) List(_ context.Context, option release.ListOption) ([]*rspb.Release, error) {
	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(option.Namespace), option.Namespace, "", blog.Infof); err != nil {
		return nil, err
	}

	lister := action.NewList(conf)
	lister.All = true
	lister.StateMask = action.ListDeployed | action.ListUninstalled | action.ListUninstalling |
		action.ListPendingInstall | action.ListPendingRollback | action.ListPendingUpgrade | action.ListFailed
	if len(option.Namespace) == 0 {
		lister.AllNamespaces = true
	}
	if len(option.Name) != 0 {
		lister.Filter = option.Name
	}
	releases, err := lister.Run()
	if err != nil {
		return nil, err
	}
	for i := range releases {
		releases[i].Config = removeValuesTemplate(releases[i].Config)
	}
	return releases, nil
}

// Install helm release through helm client
func (c *client) Install(ctx context.Context, config release.HelmInstallConfig) (*release.HelmInstallResult, error) {
	blog.Infof("sdk client try install release name %s, namespace %s, dryrun %v", config.Name, config.Namespace,
		config.DryRun)

	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(config.Namespace), config.Namespace, "", blog.Infof); err != nil {
		blog.Errorf("sdk client install and init configuration failed, %s, %v", err.Error(), config)
		return nil, err
	}

	installer := action.NewInstall(conf)
	installer.DryRun = config.DryRun
	installer.Replace = config.Replace
	installer.ClientOnly = config.ClientOnly
	installer.ReleaseName = config.Name
	installer.Namespace = config.Namespace
	installer.PostRenderer = newPatcher(c.group.config.PatchTemplates, config.PatchTemplateValues)
	valueOpts := &values.Options{}
	if err := parseArgs4Install(installer, config.Args, valueOpts); err != nil {
		blog.Errorf("sdk client install and parse from args failed, %s, args: %v", err.Error(), config.Args)
		return nil, err
	}

	// chart文件数据
	chartF, err := getChartFile(config.Chart)
	if err != nil {
		blog.Errorf("sdk client install and load chart files failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name) // nolint
		return nil, err
	}

	// values数据, 增加Var values在最后
	varValues, vars, err := c.getVarValue(config.ProjectCode, config.Namespace)
	if err != nil {
		blog.Errorf("sdk client get vars failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}
	values, err := getValues(append(parseVarValue(config.Values, vars), varValues))
	if err != nil {
		blog.Errorf("sdk client install and get values failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}
	values, err = mergeValues(valueOpts, values)
	if err != nil {
		blog.Errorf("sdk client install and get values failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	r, err := installer.Run(ctx, chartF, values)
	if err != nil {
		blog.Errorf("sdk client install failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return getHelmInstallResult(r), err
	}

	blog.Infof("sdk client install release successfully name %s, namespace %s, revision: %d, dryrun: %t",
		config.Name, config.Namespace, r.Version, config.DryRun)
	return getHelmInstallResult(r), nil
}

// Upgrade helm release through helm client
func (c *client) Upgrade(ctx context.Context, config release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error) {
	blog.Infof("sdk client try upgrade release name %s, namespace %s, dryrun %v", config.Name, config.Namespace,
		config.DryRun)

	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(config.Namespace), config.Namespace, "", blog.Infof); err != nil {
		blog.Errorf("sdk client upgrade and init configuration failed, %s, %v", err.Error(), config)
		return nil, err
	}

	upgrader := action.NewUpgrade(conf)
	upgrader.DryRun = config.DryRun
	upgrader.Namespace = config.Namespace
	upgrader.PostRenderer = newPatcher(c.group.config.PatchTemplates, config.PatchTemplateValues)
	valueOpts := &values.Options{}
	if err := parseArgs4Upgrade(upgrader, config.Args, valueOpts); err != nil {
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
	varValues, vars, err := c.getVarValue(config.ProjectCode, config.Namespace)
	if err != nil {
		blog.Errorf("sdk client get vars failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}
	values, err := getValues(append(parseVarValue(config.Values, vars), varValues))
	if err != nil {
		blog.Errorf("sdk client upgrade and get values failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}
	values, err = mergeValues(valueOpts, values)
	if err != nil {
		blog.Errorf("sdk client upgrade and get values failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	r, err := upgrader.Run(ctx, config.Name, chartF, values)
	if err != nil {
		// install when upgrade has --install args and release is not exist
		if e, ok := err.(*driver.StorageDriverError); ok && upgrader.Install &&
			errors.Is(e.Unwrap(), driver.ErrNoDeployedReleases) {
			blog.Infof("%s of namespace %s, installing it now.", e.Error(), config.Namespace)
			result, err := c.Install(context.Background(), config.ToInstallConfig()) // nolint
			if err != nil {
				return result.ToUpgradeResult(), err
			}
			blog.Infof("sdk client upgrade release successfully name %s, namespace %s, revision: %d",
				config.Name, config.Namespace, result.Release.Version)
			return result.ToUpgradeResult(), nil
		}
		blog.Errorf("sdk client upgrade failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return getHelmUpgradeResult(r), err
	}

	blog.Infof("sdk client upgrade release successfully name %s, namespace %s, revision: %d",
		config.Name, config.Namespace, r.Version)
	return getHelmUpgradeResult(r), nil
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
	uninstaller.Wait = true
	uninstaller.Timeout = 10 * time.Minute
	uninstaller.DisableOpenAPIValidation = true

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

// History get helm release history
func (c *client) History(_ context.Context, namespace, name string, max int) ([]*rspb.Release, error) {
	conf := new(action.Configuration)
	if err := conf.Init(c.getConfigFlag(namespace), namespace, "", blog.Infof); err != nil {
		return nil, err
	}
	conf.Releases.MaxHistory = max

	releases, err := action.NewHistory(conf).Run(name)
	if err != nil {
		return nil, err
	}
	for i := range releases {
		releases[i].Config = removeValuesTemplate(releases[i].Config)
	}
	return releases, nil
}

// getConfigFlag 获取helm-client配置
func (c *client) getConfigFlag(namespace string) *genericclioptions.ConfigFlags {
	flags := genericclioptions.NewConfigFlags(false)

	flags.APIServer = c.cf.APIServer
	flags.BearerToken = c.cf.BearerToken
	flags.Insecure = c.cf.Insecure
	flags.Namespace = common.GetStringP(namespace)
	flags.Timeout = common.GetStringP(defaultTimeout)
	return flags
}

// getVarValue 将命名空间变量转化到 values文件
// 兼容旧版本，注入 bcs 变量
func (c *client) getVarValue(projectCode, namespace string) (*release.File, map[string]interface{}, error) {
	// get project info
	project, err := projectClient.GetProjectByCode(projectCode)
	if err != nil {
		return nil, nil, err
	}
	variables, err := projectClient.GetVariable(projectCode, c.clusterID, namespace)
	if err != nil {
		blog.Infof("get vars failed: %s", err.Error())
	}

	// generate vars
	vars := make(map[string]interface{}, 0)
	kind := 0
	if project.Kind == "k8s" {
		kind = 1
	}
	bzID, _ := strconv.Atoi(project.BusinessID)
	bcsVars := map[string]interface{}{
		"SYS_STANDARD_DATA_ID":     0,
		"SYS_NON_STANDARD_DATA_ID": 0,
		"SYS_JFROG_DOMAIN":         "",
		"SYS_CLUSTER_ID":           c.clusterID,
		"SYS_NAMESPACE":            namespace,
		"SYS_CC_APP_ID":            bzID,
		"SYS_PROJECT_CODE":         project.ProjectCode,
		"SYS_PROJECT_ID":           project.ProjectID,
		"SYS_PROJECT_KIND":         kind,
	}
	for _, v := range variables {
		if v == nil {
			continue
		}
		bcsVars[v.Key] = v.Value
	}
	vars["__BCS__"] = bcsVars
	vars["default"] = map[string]interface{}{
		"__BCS__": bcsVars,
	}

	// marshal vars to yaml
	out, err := yaml3.Marshal(vars)
	if err != nil {
		return nil, nil, err
	}
	return &release.File{Name: "vars.yaml", Content: out}, bcsVars, nil
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

func mergeValues(valuesOpts *values.Options, base map[string]interface{}) (map[string]interface{}, error) {
	// User specified a value via --set
	for _, value := range valuesOpts.Values {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set data, %s", err.Error())
		}
	}

	// User specified a value via --set-string
	for _, value := range valuesOpts.StringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set-string data, %s", err.Error())
		}
	}
	return base, nil
}

// 渲染 values 中的变量，将 {{xxx}} 替换为变量值
func parseVarValue(fs []*release.File, vars map[string]interface{}) []*release.File {
	for i := range fs {
		repl := bcsVarRegex.ReplaceAllStringFunc(string(fs[i].Content), func(match string) string {
			varName := strings.TrimPrefix(match, "{{")
			varName = strings.TrimSuffix(varName, "}}")
			varName = strings.TrimSpace(varName)
			varName = strings.TrimPrefix(varName, ".BCS_")
			value, ok := vars[varName]
			if !ok {
				// 如果变量不存在，则直接返回原始值，不做替换
				return match
			}
			return fmt.Sprint(value)
		})
		fs[i].Content = []byte(repl)
	}
	return fs
}

// removeValuesTemplate 移除 values 中的 bcs 模版变量
func removeValuesTemplate(values map[string]interface{}) map[string]interface{} {
	delete(values, common.BCSPrefix)
	if _, ok := values[common.ValuesDefaultKey]; !ok {
		return values
	}
	if _, ok := values[common.ValuesDefaultKey].(map[string]interface{})[common.BCSPrefix]; ok {
		// 因为 ValuesDefaultKey 是 bcs-ui 下发添加的键，如果只包含 bcs 变量，则直接删除 ValuesDefaultKey,
		// 如果还有其他值，则表示 ValuesDefaultKey 下还有用户添加的值，则需要保留其他的值。
		if len(values[common.ValuesDefaultKey].(map[string]interface{})) == 1 {
			delete(values, common.ValuesDefaultKey)
		} else {
			delete(values[common.ValuesDefaultKey].(map[string]interface{}), common.BCSPrefix)
		}
	}
	return values
}

// parseArgs4Install 从用户给出的原生参数里, 直接parse到helm的install设置中
// 当前使用的flags设置版本来源helm.sh/helm/v3 v3.6.3
func parseArgs4Install(install *action.Install, args []string, valueOpts *values.Options) error {
	for i := range args {
		args[i] = strings.TrimRight(args[i], "=")
	}
	f := pflag.NewFlagSet("install", pflag.ContinueOnError)
	// 兼容更新时传入 history-max 参数，不作使用
	var _maxHistory int

	f.BoolVar(&install.CreateNamespace, "create-namespace", false, "create the release namespace if not present")
	f.BoolVar(&install.DisableHooks, "no-hooks", false, "prevent hooks from running during install")
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
	f.BoolVar(&install.DisableOpenAPIValidation, "disable-openapi-validation", true,
		"if set, the installation process will not validate rendered templates against "+
			"the Kubernetes OpenAPI Schema")
	f.BoolVar(&install.Atomic, "atomic", false,
		"if set, the installation process deletes the installation on failure. "+
			"The --wait flag will be set automatically if --atomic is used")
	f.IntVar(&_maxHistory, "history-max", defaultMaxHistory, "limit the maximum number of revisions saved "+
		"per release. Use 0 for no limit")
	f.BoolVar(&install.SkipCRDs, "skip-crds", false,
		"if set, no CRDs will be installed. By default, CRDs are installed if not already present")
	f.BoolVar(&install.SubNotes, "render-subchart-notes", false,
		"if set, render subchart notes along with the parent")
	f.BoolVar(&install.ChartPathOptions.InsecureSkipTLSverify, "insecure-skip-tls-verify", false,
		"skip tls certificate checks for the chart download")

	addValueOptionsFlags(f, valueOpts)
	return f.Parse(args)
}

// parseArgs4Upgrade 从用户给出的原生参数里, 直接parse到helm的upgrade设置中
// 当前使用的flags设置版本来源helm.sh/helm/v3 v3.6.3
func parseArgs4Upgrade(upgrade *action.Upgrade, args []string, valueOpts *values.Options) error {
	for i := range args {
		args[i] = strings.TrimRight(args[i], "=")
	}
	f := pflag.NewFlagSet("upgrade", pflag.ContinueOnError)

	// 兼容更新时传入 create-namespace 参数，不作使用
	var _createNamespace bool

	f.BoolVarP(&upgrade.Install, "install", "i", true,
		"if a release by this name doesn't already exist, run an install")
	f.BoolVar(&_createNamespace, "create-namespace", false, "create the release namespace if not present")
	f.BoolVar(&upgrade.Devel, "devel", false,
		"use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&upgrade.Recreate, "recreate-pods", false,
		"performs pods restart for the resource if applicable")
	f.BoolVar(&upgrade.Force, "force", false,
		"force resource updates through a replacement strategy")
	f.BoolVar(&upgrade.DisableHooks, "no-hooks", false,
		"disable pre/post upgrade hooks")
	f.BoolVar(&upgrade.DisableOpenAPIValidation, "disable-openapi-validation", true,
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
	f.IntVar(&upgrade.MaxHistory, "history-max", defaultMaxHistory, "limit the maximum number of revisions saved "+
		"per release. Use 0 for no limit")
	f.BoolVar(&upgrade.CleanupOnFail, "cleanup-on-fail", false,
		"allow deletion of new resources created in this upgrade when upgrade fails")
	f.BoolVar(&upgrade.SubNotes, "render-subchart-notes", false,
		"if set, render subchart notes along with the parent")
	f.StringVar(&upgrade.Description, "description", "", "add a custom description")
	f.BoolVar(&upgrade.ChartPathOptions.InsecureSkipTLSverify, "insecure-skip-tls-verify", false,
		"skip tls certificate checks for the chart download")
	f.BoolVar(&upgrade.Verify, "verify", false, "verify the package before using it")

	addValueOptionsFlags(f, valueOpts)
	return f.Parse(args)
}

func addValueOptionsFlags(f *pflag.FlagSet, v *values.Options) {
	f.StringSliceVarP(&v.ValueFiles, "values", "f", []string{},
		"specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&v.Values, "set", []string{},
		"set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.StringValues, "set-string", []string{},
		"set STRING values on the command line (can specify multiple or separate values with commas: "+
			"key1=val1,key2=val2)")
	f.StringArrayVar(&v.FileValues, "set-file", []string{},
		"set values from respective files specified via the command line (can specify multiple or separate "+
			"values with commas: key1=path1,key2=path2)")
}

func getHelmUpgradeResult(rl *rspb.Release) *release.HelmUpgradeResult {
	if rl == nil {
		return nil
	}
	var status, appVersion, lastDeployed string
	if rl.Info != nil {
		status = rl.Info.Status.String()
		lastDeployed = rl.Info.LastDeployed.Local().String()
	}
	if rl.Chart != nil && rl.Chart.Metadata != nil {
		appVersion = rl.Chart.Metadata.AppVersion
	}
	return &release.HelmUpgradeResult{
		Release:    rl,
		Revision:   rl.Version,
		Status:     status,
		AppVersion: appVersion,
		UpdateTime: lastDeployed,
	}
}

func getHelmInstallResult(rl *rspb.Release) *release.HelmInstallResult {
	if rl == nil {
		return nil
	}
	var status, appVersion, lastDeployed string
	if rl.Info != nil {
		status = rl.Info.Status.String()
		lastDeployed = rl.Info.LastDeployed.Local().String()
	}
	if rl.Chart != nil && rl.Chart.Metadata != nil {
		appVersion = rl.Chart.Metadata.AppVersion
	}
	return &release.HelmInstallResult{
		Release:    rl,
		Revision:   rl.Version,
		Status:     status,
		AppVersion: appVersion,
		UpdateTime: lastDeployed,
	}
}
