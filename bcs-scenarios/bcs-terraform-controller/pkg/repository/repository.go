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

// Package repository xxx
package repository

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	apiv1 "k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

// Handler handle
type Handler interface {
	// Init 初始化
	Init() error
	// GetLastCommitId 拉取
	GetLastCommitId() (string, error)
	// Pull 拉取代码
	Pull() error
	// Clean 清理本地代码
	Clean() error
	// GetExecutePath tf执行路径路径
	GetExecutePath() string
	// GetCommitId 获取commit-id, 必须pull成功后才能拿到commit-id
	GetCommitId() string
}

// handler git工具, 可以拉远程仓库的代码到本地
type handler struct {
	// traceId trace id
	traceId string
	// ctx context
	ctx context.Context
	// config 仓库配置信息
	config tfv1.GitRepository
	// nn namespaced and name
	nn apiv1.NamespacedName
	// token 密钥
	token string

	// gitClient git store
	gitClient store.Store
	// rootPath git项目根目录
	rootPath string
	// executePath tf执行路径路径
	executePath string
	// repoInfo 从bcs-gitops-manager拉取到仓库源信息
	repoInfo *v1alpha1.Repository
	// commitId 提交id
	commitId string
	// lastCommitId 该分支最后的commit-id
	lastCommitId string
}

// NewHandler new handler obj
func NewHandler(ctx context.Context, config tfv1.GitRepository, token, traceId string, nn apiv1.NamespacedName) Handler {
	return &handler{
		nn:      nn,
		ctx:     ctx,
		token:   token,
		config:  config,
		traceId: traceId,
	}
}

// Init 前置检查 + 获取密钥
func (g *handler) Init() error {
	if err := g.checkGitRepository(); err != nil { // 为空检查
		return errors.Wrapf(err, "check terraform '%s' failed, traceId: %s", g.nn.String(), g.traceId)
	}
	g.gitClient = store.NewStore(option.GlobalGitopsOpt)
	if err := g.gitClient.Init(); err != nil {
		return errors.Wrapf(err, "init git client failed, terraform: %s, traceId: %s", g.nn.String(), g.traceId)
	}
	ctx, cancel := context.WithTimeout(g.ctx, 30*time.Second) // 设置超时时间为30秒
	defer cancel()

	repository, err := g.gitClient.GetRepository(ctx, g.config.Repo) // 从gitopts获取git repository
	if err != nil {
		return errors.Wrapf(err, "get repository failed, terraform: %s, traceId: %s", g.nn.String(), g.traceId)
	}
	if repository == nil {
		return errors.Errorf("repository '%s' is nil, traceId: %s", g.config.Repo, g.traceId)
	}
	blog.Infof("query '%s' repository: %s", g.config.Repo, utils.ToJsonString(repository))

	g.repoInfo = repository
	g.rootPath = path.Join(option.RepositoryStorePath, g.traceId)
	blog.Infof("terraform: %s, save path: %s, repo: %s", g.nn.String(), g.rootPath, g.config.Repo)

	return nil
}

// GetLastCommitId 拉取
// 仅用作快速验证
func (g *handler) GetLastCommitId() (string, error) {
	if len(g.lastCommitId) != 0 {
		return g.lastCommitId, nil
	}
	storer := memory.NewStorage()
	remoteRepo := git.NewRemote(storer, &config.RemoteConfig{
		Name: g.config.TargetRevision,
		URLs: []string{g.config.Repo},
	})

	// 拉取远程仓库的引用信息
	refs, err := remoteRepo.List(&git.ListOptions{
		Auth: &http.BasicAuth{
			Username: g.repoInfo.Username,
			Password: g.token,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to list references: %s, tf: %s", g.config.TargetRevision, g.nn)
	}

	// 遍历引用信息，打印分支名和commit ID
	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == g.config.TargetRevision {
			g.lastCommitId = ref.Hash().String()
			return g.lastCommitId, nil
		}
	}

	return "", errors.Errorf("not found branch, target: %s, tf: %s", g.config.TargetRevision, g.nn)
}

// Pull repository to local
func (g *handler) Pull() error {
	opt := &git.CloneOptions{
		URL: g.repoInfo.Repo,
		Auth: &http.BasicAuth{
			Username: g.repoInfo.Username,
			Password: g.token,
		},
	}

	repo, err := git.PlainClone(g.rootPath, false, opt)
	if err != nil {
		return errors.Wrapf(err, "pull repository failed, repo: %s, terraform: %s, traceId: %s",
			g.repoInfo.Proxy, g.nn.String(), g.traceId)
	}
	blog.Infof("git clone success, repo: %s", g.repoInfo.Repo)

	head, err := repo.Head()
	if err != nil {
		return errors.Wrapf(err, "get git repositroy head info failed, repo: %s, terraform: %s, traceId: %s",
			g.repoInfo.Proxy, g.nn.String(), g.traceId)
	}

	g.executePath = path.Join(g.rootPath, g.config.Path)
	if len(g.config.TargetRevision) == 0 {
		blog.Infof("current branch info: %s, execute path: %s", head.String(), g.executePath)
		return nil
	}

	if err = g.checkoutRevision(repo); err != nil {
		return err
	}

	return nil
}

// Clean 及时清理本地仓库
func (g *handler) Clean() error {
	var err error
	if len(g.rootPath) == 0 {
		return nil
	}
	if err = os.RemoveAll(g.rootPath); err == nil {
		blog.Infof("remove '%s' path success, terraform: %s", g.rootPath, g.nn.String())
		return nil
	}

	return errors.Wrapf(err, "remove '%s' path failed, terraform: %s, traceId: %s", g.rootPath,
		g.nn.String(), g.traceId)
}

// GetExecutePath tf执行路径路径
func (g *handler) GetExecutePath() string {
	return g.executePath
}

// GetCommitId 获取commit-id, 必须pull成功后才能拿到commit-id
func (g *handler) GetCommitId() string {
	return g.commitId
}

// checkGitRepository 检查repository是否为空
func (g *handler) checkGitRepository() error {
	if len(g.config.Repo) == 0 {
		return errors.New("repo address is nil")
	}
	return nil
}

// checkoutRevision git checkout revision
func (g *handler) checkoutRevision(repo *git.Repository) error {
	if repo == nil {
		return nil
	}

	// resolve revision
	hash, err := repo.ResolveRevision(plumbing.Revision(g.config.TargetRevision))
	if err != nil {
		return errors.Wrapf(err, "revisoin not foud, repo: %s, revision: %s",
			g.config.Repo, g.config.TargetRevision)
	}

	// get work tree
	worktree, err := repo.Worktree()
	if err != nil {
		return errors.Wrapf(err, "getting worktree failed, repo: %s, revision: %s",
			g.config.Repo, g.config.TargetRevision)
	}

	// checkout out revision
	if err = worktree.Checkout(&git.CheckoutOptions{Hash: *hash}); err != nil {
		return errors.Wrapf(err, "checkout out revision failed, repo: %s, revision: %s",
			g.config.Repo, g.config.TargetRevision)
	}

	blog.Infof("checkout revision success, repo: %s, current revision: %s, commit-id: %s, execute path: %s",
		g.config.Repo, g.config.TargetRevision, hash.String(), g.executePath)

	g.commitId = hash.String()

	return nil
}
