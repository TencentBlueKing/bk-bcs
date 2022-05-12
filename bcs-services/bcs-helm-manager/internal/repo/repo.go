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

package repo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// Platform 定义了repo的操作平台, 如对接bk-repo
type Platform interface {
	User(User) Handler
}

type User struct {
	Name     string
	Password string
}

// Handler 是平台下的repo基础操作对象, 根据不同租户信息得到的
type Handler interface {
	Project(projectID string) ProjectHandler
}

// ProjectHandler 定义了 Handler 下的对每一个 Project 的操作能力
type ProjectHandler interface {
	Ensure(ctx context.Context) error

	Repository(repoType RepositoryType, repository string) RepositoryHandler
}

// RepositoryHandler 定义了 ProjectHandler 下对每个 Repository 的操作能力
type RepositoryHandler interface {
	Get(ctx context.Context) (*Repository, error)
	Create(ctx context.Context, repository *Repository) (string, error)

	ListChart(ctx context.Context, option ListOption) (*ListChartData, error)

	Chart(chartName string) ChartHandler

	CreateUser(ctx context.Context) (string, string, error)
}

// ChartHandler 定义了 ProjectHandler 下对每个 Chart 的操作能力
type ChartHandler interface {
	ListVersion(ctx context.Context, option ListOption) (*ListChartVersionData, error)
	Detail(ctx context.Context, version string) (*ChartDetail, error)
	Download(ctx context.Context, version string) ([]byte, error)
}

// Config 定义了 Platform 的配置
type Config struct {
	URL      string
	OciURL   string
	AuthType string
	Username string
	Password string
}

// Auth 定义了repo下基础操作的权限信息
type Auth struct {
	Type     string
	Operator string
	Username string
	Password string
}

// Project 定义了 repo 下的项目信息
type Project struct {
	Name        string
	DisplayName string
	Description string
}

// Repository 定义了 Project 下的仓库信息
type Repository struct {
	ProjectID   string
	Name        string
	Type        RepositoryType
	Description string

	Remote         bool
	RemoteURL      string
	RemoteUsername string
	RemotePassword string
}

// RepositoryType 用来区分不同的 Repository 类型
type RepositoryType int

const (
	RepositoryTypeUnknown RepositoryType = iota
	RepositoryTypeHelm
	RepositoryTypeOCI
)

var repositoryTypes = map[RepositoryType]string{
	RepositoryTypeUnknown: "UNKNOWN",
	RepositoryTypeHelm:    "HELM",
	RepositoryTypeOCI:     "OCI",
}

// String return the string name of RepositoryType
func (rt RepositoryType) String() string {
	if s, ok := repositoryTypes[rt]; ok {
		return s
	}

	return "UNKNOWN"
}

// GetRepositoryType receive a string name and return the related RepositoryType
func GetRepositoryType(name string) RepositoryType {
	for k, v := range repositoryTypes {
		if v == name {
			return k
		}
	}

	return RepositoryTypeUnknown
}

// ListChartData 描述了分页查询 Chart 的返回信息
type ListChartData struct {
	Total  int64
	Page   int64
	Size   int64
	Charts []*Chart
}

// Chart 定义了 Repository 下的chart包信息
type Chart struct {
	Key         string
	Name        string
	Type        string
	Version     string
	AppVersion  string
	Description string
	CreateTime  string
	CreateBy    string
	UpdateTime  string
	UpdateBy    string
}

// Transfer2Proto transfer the data into protobuf struct
func (c *Chart) Transfer2Proto() *helmmanager.Chart {
	return &helmmanager.Chart{
		Name:              common.GetStringP(c.Name),
		Key:               common.GetStringP(c.Key),
		Type:              common.GetStringP(c.Type),
		LatestVersion:     common.GetStringP(c.Version),
		LatestAppVersion:  common.GetStringP(c.AppVersion),
		LatestDescription: common.GetStringP(c.Description),
		CreateBy:          common.GetStringP(c.CreateBy),
		UpdateBy:          common.GetStringP(c.UpdateBy),
		CreateTime:        common.GetStringP(c.CreateTime),
		UpdateTime:        common.GetStringP(c.UpdateTime),
	}
}

// ListChartVersionData 描述了分页查询 ChartVersion 的返回信息
type ListChartVersionData struct {
	Total    int64
	Page     int64
	Size     int64
	Versions []*ChartVersion
}

// ChartVersion 定义了 Chart 下不同版本的信息
type ChartVersion struct {
	Name        string
	Version     string
	AppVersion  string
	Description string
	CreateTime  string
	CreateBy    string
	UpdateTime  string
	UpdateBy    string
}

// Transfer2Proto transfer the data into protobuf struct
func (cv *ChartVersion) Transfer2Proto() *helmmanager.ChartVersion {
	return &helmmanager.ChartVersion{
		Name:        common.GetStringP(cv.Name),
		Version:     common.GetStringP(cv.Version),
		AppVersion:  common.GetStringP(cv.AppVersion),
		Description: common.GetStringP(cv.Description),
		CreateBy:    common.GetStringP(cv.CreateBy),
		UpdateBy:    common.GetStringP(cv.UpdateBy),
		CreateTime:  common.GetStringP(cv.CreateTime),
		UpdateTime:  common.GetStringP(cv.UpdateTime),
	}
}

// ChartDetail 定义了 Chart 某一个版本的内容信息
type ChartDetail struct {
	Name    string
	Version string

	Contents map[string]*FileContent
}

// Transfer2Proto transfer the data into protobuf struct
func (cd *ChartDetail) Transfer2Proto() *helmmanager.ChartDetail {
	r := &helmmanager.ChartDetail{
		Name:     common.GetStringP(cd.Name),
		Version:  common.GetStringP(cd.Version),
		Contents: make(map[string]*helmmanager.FileContent),
	}

	for k, v := range cd.Contents {
		r.Contents[k] = v.Transfer2Proto()
	}

	return r
}

// LoadContentFromTgz receive the byte data in gzip format(tgz), and parse the file content into ChartDetail's Content
func (cd *ChartDetail) LoadContentFromTgz(data []byte) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}

	cd.Contents = make(map[string]*FileContent)
	tr := tar.NewReader(gz)
	for {
		cur, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if cur.Typeflag != tar.TypeReg {
			continue
		}

		fd, err := io.ReadAll(tr)
		if err != nil {
			return err
		}

		cd.Contents[cur.Name] = &FileContent{
			Name:    path.Base(cur.Name),
			Path:    cur.Name,
			Content: fd,
		}
	}

	return nil
}

// FileContent 定义了 ChartDetail 下每个文件的内容信息
type FileContent struct {
	Name    string
	Path    string
	Content []byte
}

// Transfer2Proto transfer the data into protobuf struct
func (fc *FileContent) Transfer2Proto() *helmmanager.FileContent {
	return &helmmanager.FileContent{
		Name:    common.GetStringP(fc.Name),
		Path:    common.GetStringP(fc.Path),
		Content: common.GetStringP(string(fc.Content)),
	}
}

// ListOption 定义了批量查询的参数
type ListOption struct {
	Page int64
	Size int64
}
