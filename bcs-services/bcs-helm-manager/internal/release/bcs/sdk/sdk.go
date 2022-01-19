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
		blog.Errorf("sdk client install and get values file %s failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	r, err := installer.Run(chartF, values)
	if err != nil {
		blog.Errorf("sdk client install failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	blog.Infof("sdk client install release successfully name %s, namespace %s, revision: %d",
		config.Name, config.Namespace, r.Version)
	return &release.HelmInstallResult{Revision: r.Version}, nil
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
		blog.Errorf("sdk client upgrade and get values file %s failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	r, err := upgrader.Run(config.Name, chartF, values)
	if err != nil {
		blog.Errorf("sdk client upgrade failed, %s, "+
			"namespace %s, name %s", err.Error(), config.Namespace, config.Name)
		return nil, err
	}

	blog.Infof("sdk client upgrade release successfully name %s, namespace %s, revision: %d",
		config.Name, config.Namespace, r.Version)
	return &release.HelmUpgradeResult{Revision: r.Version}, nil
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

func (c *client) getConfigFlag(namespace string) *genericclioptions.ConfigFlags {
	flags := genericclioptions.NewConfigFlags(false)

	flags.APIServer = c.cf.APIServer
	flags.BearerToken = c.cf.BearerToken
	flags.Insecure = c.cf.Insecure
	flags.Namespace = common.GetStringP(namespace)
	return flags
}

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

func getChartFile(f *release.File) (*chart.Chart, error) {
	bufferedFile, err := loader.LoadArchiveFiles(bytes.NewReader(f.Content))
	if err != nil {
		return nil, err
	}

	return loader.LoadFiles(bufferedFile)
}

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

func replaceVarTplKey(keys map[string]string, data []byte) []byte {
	for k, v := range keys {
		data = []byte(strings.ReplaceAll(string(data), common.Vtk(k), v))
	}

	return common.EmptyAllVarTemplateKey(data)
}
