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

package repository

import (
	"context"
	"os"

	"github.com/argoproj/argo-cd/v2/util/db"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

// Handler 定义用来进行仓库相关操作的接口
type Handler interface {
	GetLastCommitId(ctx context.Context, repo *tfv1.GitRepository) (string, error)
	CheckoutCommit(ctx context.Context, repo *tfv1.GitRepository, commitID, path string) (string, error)
}

type handler struct {
	argoDB db.ArgoDB
}

// NewRepositoryHandler 创建 Repository Handler 的实例
func NewRepositoryHandler(argoDB db.ArgoDB) Handler {
	return &handler{
		argoDB: argoDB,
	}
}

func (h *handler) buildRepoAuth(ctx context.Context, repoUrl string) (transport.AuthMethod, error) {
	argoRepo, err := h.argoDB.GetRepository(ctx, repoUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "get repository '%s' from argo db failed", repoUrl)
	}
	if argoRepo == nil {
		return nil, errors.Errorf("repository '%s' not found", repoUrl)
	}
	if argoRepo.Username != "" && argoRepo.Password != "" {
		return &http.BasicAuth{
			Username: argoRepo.Username,
			Password: argoRepo.Password,
		}, nil
	}
	if argoRepo.SSHPrivateKey != "" {
		publicKeys, err := ssh.NewPublicKeys("git", []byte(argoRepo.SSHPrivateKey), "")
		if err != nil {
			return nil, errors.Wrapf(err, "create public keys failed")
		}
		return publicKeys, nil
	}
	return nil, errors.Errorf("not https/ssh authentication")
}

// GetLastCommitId 获取仓库的最后一次提交 CommitID
func (h *handler) GetLastCommitId(ctx context.Context, repo *tfv1.GitRepository) (string, error) {
	memStore := memory.NewStorage()
	remoteRepo := git.NewRemote(memStore, &config.RemoteConfig{
		Name: repo.Repo,
		URLs: []string{repo.Repo},
	})
	auth, err := h.buildRepoAuth(ctx, repo.Repo)
	if err != nil {
		return "", errors.Wrapf(err, "repository build authencation failed")
	}

	refs, err := remoteRepo.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to list references")
	}
	refRevisions := make(map[string]string)
	var targetRefName string
	for _, ref := range refs {
		if ref.Type() == plumbing.HashReference {
			refRevisions[ref.Name().String()] = ref.Hash().String()
		}
		if ref.Name().Short() == repo.TargetRevision {
			if ref.Type() == plumbing.HashReference {
				return ref.Hash().String(), nil
			}
			if ref.Type() == plumbing.SymbolicReference {
				targetRefName = ref.Target().String()
			}
		}
	}
	hash, ok := refRevisions[targetRefName]
	if ok {
		return hash, nil
	}
	return "", errors.Errorf("not found branch '%s'", repo.TargetRevision)
}

// CheckoutCommit 将目录中的代码仓库 checkout 到对应的 commit
func (h *handler) CheckoutCommit(ctx context.Context, repo *tfv1.GitRepository, commitID, path string) (string, error) {
	repoName := utils.ParseGitRepoName(repo.Repo)
	repoPath := path + "/" + repoName
	auth, err := h.buildRepoAuth(ctx, repo.Repo)
	if err != nil {
		return "", errors.Wrapf(err, "repository build authencation failed")
	}
	if err := os.RemoveAll(repoPath); err != nil {
		return "", errors.Wrapf(err, "remove repo '%s' failed", repoPath)
	}
	gitRepo, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:  repo.Repo,
		Auth: auth,
	})
	if err != nil {
		return repoPath, errors.Wrapf(err, "clone repository failed")
	}
	logctx.Infof(ctx, "clone repository success")
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return repoPath, errors.Wrapf(err, "repository get worktree failed")
	}
	if err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitID),
	}); err != nil {
		return repoPath, errors.Wrapf(err, "repository checkout commit '%s' failed", commitID)
	}
	return repoPath, nil
}
