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

package bkrepo

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	bkRepoAuth "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo/bkrepo/auth"
)

// New 返回一个bkRepo, 标准的repo.Platform对象, 其背后是基于蓝鲸制品库的实现
func New(c repo.Config) repo.Platform {
	return &bkRepo{
		config: &c,
		client: newClient(),
	}
}

// bkRepo 基于蓝鲸制品库, 使用蓝鲸制品库的对外接口来操作仓库
type bkRepo struct {
	config *repo.Config

	client *client
}

// User 针对给定用户权限实例化一个handler, 共享bkRepo的client
func (br *bkRepo) User(user repo.User) repo.Handler {
	return &handler{
		bkRepo: br,
		user:   user,
		auth:   bkRepoAuth.New(br.config.AuthType, user.Name, br.config.Username, br.config.Password),
	}
}

type handler struct {
	*bkRepo

	auth *bkRepoAuth.Auth
	user repo.User
}

// Project 针对给定的projectID, 返回一个 repo.ProjectHandler 实例, 用于项目层级的所有操作
func (h *handler) Project(projectID string) repo.ProjectHandler {
	return &projectHandler{
		handler:   h,
		projectID: projectID,
	}
}

type projectHandler struct {
	*handler

	projectID string
}

// Ensure 针对给定的一个project name, 确保它存在于bk-repo中, 若不存在则创建
func (ph *projectHandler) Ensure(ctx context.Context) error {
	return ph.ensureProject(ctx, &repo.Project{
		Name:        ph.projectID,
		DisplayName: ph.projectID,
		Description: "created by bcs-helm-manager",
	})
}

// Repository 针对给定的repository type和repository name, 返回一个 repo.RepositoryHandler 实例, 用于仓库层级的所有操作
func (ph *projectHandler) Repository(repoType repo.RepositoryType, repository string) repo.RepositoryHandler {
	return &repositoryHandler{
		projectHandler: ph,
		projectID:      ph.projectID,
		repository:     repository,
		repoType:       repoType,
	}
}

type repositoryHandler struct {
	*projectHandler

	projectID  string
	repository string
	repoType   repo.RepositoryType
}

// Get 获取指定的repository信息
func (rh *repositoryHandler) Get(ctx context.Context) (*repo.Repository, error) {
	return rh.getRepository(ctx)
}

// Create 创建一个repository
func (rh *repositoryHandler) Create(ctx context.Context, repository *repo.Repository) (string, error) {
	if repository == nil {
		return "", fmt.Errorf("repository can not be empty")
	}

	repository.ProjectID = rh.projectID
	repository.Name = rh.repository
	repository.Type = rh.repoType
	return rh.createRepository(ctx, repository)
}

// ListChart 针对给定的分页信息, 返回chart维度的list数据, 同一个chart的多个版本会被合并, 只展示最新的版本信息
func (rh *repositoryHandler) ListChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	return rh.listChart(ctx, option)
}

// Chart 针对给定的chart名称, 返回一个 repo.ChartHandler 实例, 用于chart层级的所有操作
func (rh *repositoryHandler) Chart(chartName string) repo.ChartHandler {
	return &chartHandler{
		repositoryHandler: rh,
		projectID:         rh.projectID,
		repository:        rh.repository,
		repoType:          rh.repoType,
		chartName:         chartName,
	}
}

// CreateUser 针对当前的repository, 创建一个管理员账号, 并返回账号的username和password
func (rh *repositoryHandler) CreateUser(ctx context.Context) (string, string, error) {
	return rh.createUser(ctx)
}

type chartHandler struct {
	*repositoryHandler

	projectID  string
	repository string
	repoType   repo.RepositoryType
	chartName  string
}

// ListVersion 返回该chart的版本信息列表
func (ch *chartHandler) ListVersion(ctx context.Context, option repo.ListOption) (*repo.ListChartVersionData, error) {
	return ch.listChartVersion(ctx, option)
}

// Detail 返回该chart指定version的详细信息
func (ch *chartHandler) Detail(ctx context.Context, version string) (*repo.ChartDetail, error) {
	return ch.getChartVersionDetail(ctx, version)
}

// Download 返回该chart指定version的源文件信息
func (ch *chartHandler) Download(ctx context.Context, version string) ([]byte, error) {
	return ch.downloadChartVersion(ctx, version)
}
