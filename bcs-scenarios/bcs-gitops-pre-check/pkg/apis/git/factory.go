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

// Package git xxx
package git

import (
	"context"
	"fmt"
	"strings"

	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

const (
	// TGitType tGit type
	TGitType = "tGit"
	// TGitOutType tGit type
	TGitOutType = "tGitOut"
	// TGitCommunityType community type
	TGitCommunityType = "tGitCommunity"
	// GitlabType gitlab
	GitlabType = "gitlab"
)

// Client git
type Client interface {
	GetMrInfo(ctx context.Context, repoAddr, token, mrIID string) (*precheck.MRInfoData, error)
	CommentMR(ctx context.Context, repoAddr, token, mrIID, comment string) error
	SubmitCheckState(ctx context.Context, checkSystem, state, targetUrl, description,
		token string, block bool, req *precheck.PreCheckTask) error
	GetBranchDetail(ctx context.Context, repoAddr, branch, token string) (*TGitBranch, error)
	GetTagDetail(ctx context.Context, repoAddr, tag, token string) (*TGitTag, error)
}

// Factory interface
type Factory interface {
	GetClient(repo string) (Client, error)
	Init()
}

type factory struct {
	clients map[string]Client
	opts    *FactoryOpts
}

// FactoryOpts options
type FactoryOpts struct {
	TGitEndpoint          string
	TGitSubStr            string
	TGitOutEndpoint       string
	TGitOutSubStr         string
	TGitCommunityEndpoint string
	TGitCommunitySubStr   string
	GitlabEndpoint        string
	GitlabSubStr          string
}

type clientOpts struct {
	endpoint string
	subStr   string
}

// NewFactory new factory
func NewFactory(opts *FactoryOpts) Factory {
	return &factory{opts: opts, clients: make(map[string]Client)}
}

// GetClient get client
func (f *factory) GetClient(repo string) (Client, error) {
	name, err := f.judgeRepoType(repo)
	if err != nil {
		return nil, err
	}
	if exe, ok := f.clients[name]; ok {
		return exe, nil
	}
	return nil, fmt.Errorf("wrong client type %s", name)
}

// Init init
func (f *factory) Init() {
	tGitOpts := &clientOpts{endpoint: f.opts.TGitEndpoint, subStr: f.opts.TGitSubStr}
	f.clients[TGitType] = NewTGit(tGitOpts)
	tGitOutOpts := &clientOpts{endpoint: f.opts.TGitOutEndpoint, subStr: f.opts.TGitOutSubStr}
	f.clients[TGitOutType] = NewTGit(tGitOutOpts)
	tGitCommunityOpts := &clientOpts{endpoint: f.opts.TGitCommunityEndpoint, subStr: f.opts.TGitCommunitySubStr}
	f.clients[TGitCommunityType] = NewTGit(tGitCommunityOpts)
	gitlabOpts := &clientOpts{endpoint: f.opts.GitlabEndpoint, subStr: f.opts.GitlabSubStr}
	f.clients[GitlabType] = NewGitLab(gitlabOpts)
}

func (f *factory) judgeRepoType(repo string) (string, error) {
	if strings.Contains(repo, f.opts.TGitSubStr) {
		return TGitType, nil
	}
	if strings.Contains(repo, f.opts.TGitOutSubStr) {
		return TGitOutType, nil
	}
	if strings.Contains(repo, f.opts.TGitCommunitySubStr) {
		return TGitCommunityType, nil
	}
	if strings.Contains(repo, f.opts.GitlabSubStr) {
		return GitlabType, nil
	}
	return "", fmt.Errorf("not supported repo:%s", repo)
}
