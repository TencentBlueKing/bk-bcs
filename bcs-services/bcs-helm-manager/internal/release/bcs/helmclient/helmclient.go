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

package helmclient

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
)

const (
	tmplKeyAPIServer = "__bcs_api__"
	tmplKeyClusterID = "__cluster_id__"
	tmplKeyToken     = "__token__"

	kubeConfigFile = "kubeconfig"
)

// Config describe the configuration for helm client group
type Config struct {
	// ConfigDir 指定了kube config暂存的目录
	// 其下会针对不同的请求, 划分临时目录
	ConfigDir          string
	BcsAPI             string
	Token              string
	Binary             string
	KubeConfigTemplate string
}

func (c *Config) renderTemplate(clusterID string) string {
	s := c.KubeConfigTemplate
	s = strings.ReplaceAll(s, tmplKeyAPIServer, c.BcsAPI)
	s = strings.ReplaceAll(s, tmplKeyClusterID, clusterID)
	s = strings.ReplaceAll(s, tmplKeyToken, c.Token)
	return s
}

// NewGroup return a new Group instance
func NewGroup(c Config) Group {
	return &helmClientGroup{
		config: &c,
		groups: make(map[string]*helmClient),
	}
}

// Group 定义了一组 Client
type Group interface {
	Cluster(clusterID string) (Client, error)
}

type helmClientGroup struct {
	config *Config

	sync.RWMutex
	groups map[string]*helmClient
}

// Cluster 生成一个新的client, 用于执行helm client指令
func (hg *helmClientGroup) Cluster(clusterID string) (Client, error) {
	r := hg.newHelmClient(clusterID)
	if err := r.init(); err != nil {
		return nil, err
	}

	return r, nil
}

func (hg *helmClientGroup) newHelmClient(clusterID string) *helmClient {
	r := &helmClient{}

	for {
		r.id = common.RandomString(5) + strconv.FormatInt(time.Now().Unix(), 10)
		if hg.insert2Group(r) {
			break
		}
	}

	r.binary = hg.config.Binary
	r.clusterID = clusterID
	r.runDir = filepath.Join(hg.config.ConfigDir, r.id)
	r.kubeConfig = hg.config.renderTemplate(clusterID)
	return r
}

func (hg *helmClientGroup) insert2Group(client *helmClient) bool {
	hg.Lock()
	defer hg.Unlock()

	if _, ok := hg.groups[client.id]; ok {
		return false
	}

	hg.groups[client.id] = client
	return true
}

// Client 定义了helm执行命令的接口
type Client interface {
	Close() error
	Install(ctx context.Context, conf release.HelmInstallConfig) (*release.HelmInstallResult, error)
	Uninstall(ctx context.Context, conf release.HelmUninstallConfig) (*release.HelmUninstallResult, error)
	Upgrade(ctx context.Context, conf release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error)
	Rollback(ctx context.Context, conf release.HelmRollbackConfig) (*release.HelmRollbackResult, error)
}

type helmClient struct {
	sync.RWMutex

	id        string
	available bool

	binary     string
	runDir     string
	clusterID  string
	kubeConfig string

	env []string
}

// Close the client connections and configurations
func (c *helmClient) Close() error {
	c.Lock()
	defer c.Unlock()

	if !c.available {
		return nil
	}

	if err := os.Remove(c.runDir); err != nil {
		return err
	}

	c.available = false
	return nil
}

func (c *helmClient) init() error {
	if err := os.MkdirAll(c.runDir, 0666); err != nil {
		blog.Errorf("init helm client id %s cluster %s mkdir failed, %s", c.id, c.clusterID, err.Error())
		return err
	}

	if err := c.saveFile("", &release.File{
		Name:    kubeConfigFile,
		Content: []byte(c.kubeConfig),
	}); err != nil {
		blog.Errorf("init helm client id %s cluster %s save kube config failed, %s",
			c.id, c.clusterID, err.Error())
		return err
	}

	c.available = true
	blog.Infof("init helm client successfully, id %s for cluster %s", c.id, c.clusterID)
	return nil
}

func (c *helmClient) isolateDir() (string, error) {
	c.Lock()
	defer c.Unlock()

	if !c.available {
		return "", fmt.Errorf("helm client not init or is closed")
	}

	for {
		dir := common.RandomString(5) + strconv.FormatInt(time.Now().Unix(), 10)
		err := os.Mkdir(filepath.Join(c.runDir, dir), 0666)
		if err == nil {
			return dir, nil
		}

		if err != os.ErrExist {
			return "", err
		}
	}
}

func (c *helmClient) exec(ctx context.Context, args []string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(
		ctx,
		"/bin/bash",
		"-c",
		strings.Join(append([]string{c.binary, "--kubeconfig=" + kubeConfigFile}, args...), " "),
	)
	cmd.Dir = c.runDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	blog.Infof("helm client going to exec command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		blog.Errorf("helm client exec command failed, %s, command: %s", err.Error(), cmd.String())
		return stdout.Bytes(), stderr.Bytes(), err
	}

	blog.Infof("helm client exec command successfully, command: %s, stdout: %s, stderr: %s",
		cmd.String(), stdout.String(), stderr.String())
	return stdout.Bytes(), stderr.Bytes(), nil
}

// Install helm chart through helm client
func (c *helmClient) Install(ctx context.Context, conf release.HelmInstallConfig) (*release.HelmInstallResult, error) {
	privateDir, err := c.isolateDir()
	if err != nil {
		blog.Errorf("helm client install id %s cluster %s, get isolate dir failed, %s",
			c.id, c.clusterID, err.Error())
		return nil, err
	}

	args := []string{
		"install",
		conf.Name,
		filepath.Join(privateDir, conf.Chart.Name),
		"-n", conf.Namespace,
	}

	if err := c.saveFile(privateDir, conf.Chart); err != nil {
		blog.Errorf("helm client install id %s cluster %s, save chart file %s failed, %s",
			c.id, c.clusterID, conf.Chart.Name, err.Error())
		return nil, err
	}

	for _, valueF := range conf.Values {
		if err := c.saveFile(privateDir, valueF); err != nil {
			blog.Errorf("helm client install id %s cluster %s, save values file %s failed, %s",
				c.id, c.clusterID, valueF.Name, err.Error())
			return nil, err
		}
		args = append(args, "-f", filepath.Join(privateDir, valueF.Name))
	}

	if conf.DryRun {
		args = append(args, "--dry-run")
	}

	stdout, stderr, err := c.exec(ctx, args)
	if err != nil {
		blog.Errorf("helm client install id %s cluster %s chart %s namespace %s name %s failed, %s, "+
			"stdout: %s, stderr: %s",
			c.id, c.clusterID, conf.Chart.Name, conf.Namespace, conf.Name, err.Error(), string(stdout), string(stderr))
		return nil, fmt.Errorf("%s, stderr: %s", err.Error(), string(stderr))
	}

	output := &releaseOutput{}
	output.parse(string(stdout))
	blog.Infof("helm client install successfully id %s cluster %s chart %s", c.id, c.clusterID, conf.Chart.Name)
	return &release.HelmInstallResult{Revision: output.revision}, nil
}

// Uninstall helm chart through helm client
func (c *helmClient) Uninstall(ctx context.Context, conf release.HelmUninstallConfig) (
	*release.HelmUninstallResult, error) {

	args := []string{
		"uninstall",
		conf.Name,
		"-n", conf.Namespace,
	}
	if conf.DryRun {
		args = append(args, "--dry-run")
	}

	stdout, stderr, err := c.exec(ctx, args)
	if err != nil {
		blog.Errorf("helm client uninstall id %s cluster %s namespace %s name %s failed, %s, "+
			"stdout: %s, stderr: %s",
			c.id, c.clusterID, conf.Namespace, conf.Name, err.Error(), string(stdout), string(stderr))
		return nil, fmt.Errorf("%s, stderr: %s", err.Error(), string(stderr))
	}

	blog.Infof("helm client uninstall successfully id %s cluster %s namespace %s name %s",
		c.id, c.clusterID, conf.Namespace, conf.Name)
	return &release.HelmUninstallResult{}, nil
}

// Upgrade helm chart through helm client
func (c *helmClient) Upgrade(ctx context.Context, conf release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error) {
	privateDir, err := c.isolateDir()
	if err != nil {
		blog.Errorf("helm client upgrade id %s cluster %s, get isolate dir failed, %s",
			c.id, c.clusterID, err.Error())
		return nil, err
	}

	args := []string{
		"upgrade",
		conf.Name,
		filepath.Join(privateDir, conf.Chart.Name),
		"-n", conf.Namespace,
	}

	if err := c.saveFile(privateDir, conf.Chart); err != nil {
		blog.Errorf("helm client upgrade id %s cluster %s, save chart file %s failed, %s",
			c.id, c.clusterID, conf.Chart.Name, err.Error())
		return nil, err
	}

	for _, valueF := range conf.Values {
		if err := c.saveFile(privateDir, valueF); err != nil {
			blog.Errorf("helm client upgrade id %s cluster %s, save values file %s failed, %s",
				c.id, c.clusterID, valueF.Name, err.Error())
			return nil, err
		}
		args = append(args, "-f", filepath.Join(privateDir, valueF.Name))
	}

	if conf.DryRun {
		args = append(args, "--dry-run")
	}

	stdout, stderr, err := c.exec(ctx, args)
	if err != nil {
		blog.Errorf("helm client upgrade id %s cluster %s chart %s namespace %s name %s failed, %s, "+
			"stdout: %s, stderr: %s",
			c.id, c.clusterID, conf.Chart.Name, conf.Namespace, conf.Name, err.Error(), string(stdout), string(stderr))
		return nil, fmt.Errorf("%s, stderr: %s", err.Error(), string(stderr))
	}

	output := &releaseOutput{}
	output.parse(string(stdout))
	blog.Infof("helm client upgrade successfully id %s cluster %s chart %s", c.id, c.clusterID, conf.Chart.Name)
	return &release.HelmUpgradeResult{Revision: output.revision}, nil
}

// Rollback helm chart through helm client
func (c *helmClient) Rollback(ctx context.Context, conf release.HelmRollbackConfig) (
	*release.HelmRollbackResult, error) {

	args := []string{
		"rollback",
		conf.Name,
		"-n", conf.Namespace,
	}
	if conf.DryRun {
		args = append(args, "--dry-run")
	}

	stdout, stderr, err := c.exec(ctx, args)
	if err != nil {
		blog.Errorf("helm client rollback id %s cluster %s namespace %s name %s failed, %s, "+
			"stdout: %s, stderr: %s",
			c.id, c.clusterID, conf.Namespace, conf.Name, err.Error(), string(stdout), string(stderr))
		return nil, fmt.Errorf("%s, stderr: %s", err.Error(), string(stderr))
	}

	blog.Infof("helm client rollback successfully id %s cluster %s namespace %s name %s to revision %d",
		c.id, c.clusterID, conf.Namespace, conf.Name, conf.Revision)
	return &release.HelmRollbackResult{}, nil
}

func (c *helmClient) saveFile(prefixDir string, f *release.File) error {
	fi, err := os.OpenFile(filepath.Join(c.runDir, prefixDir, f.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		blog.Errorf("helm client save file id %s cluster %s open file %s failed, %s",
			c.id, c.clusterID, f.Name, err.Error())
		return err
	}

	if _, err = fi.Write(f.Content); err != nil {
		blog.Errorf("helm client save file id %s cluster %s write data into file %s failed, %s",
			c.id, c.clusterID, f.Name, err.Error())
		return err
	}

	return nil
}

// install and upgrade stdout example following:
//   NAME: helm-manager
//   LAST DEPLOYED: Thu Jan  6 17:35:29 2022
//   NAMESPACE: hm
//   STATUS: deployed
//   REVISION: 1
//   TEST SUITE: None
type releaseOutput struct {
	name         string
	lastDeployed string
	namespace    string
	status       string
	revision     int
}

func (ro *releaseOutput) parse(rawContent string) {
	for _, line := range strings.Split(rawContent, "\n") {
		gap := strings.Index(line, ": ")
		if gap < 0 || gap+2 >= len(line) {
			continue
		}

		value := line[gap+2:]
		switch line[:gap] {
		case "NAME":
			ro.name = value
		case "LAST DEPLOYED":
			ro.lastDeployed = value
		case "NAMESPACE":
			ro.namespace = value
		case "STATUS":
			ro.status = value
		case "REVISION":
			ro.revision, _ = strconv.Atoi(value)
		}
	}
}
