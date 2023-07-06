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

package argocd

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	argorepo "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

type argoGrpcHandler func(ctx context.Context, req *http.Request) *httpResponse

var (
	// grpcAccessUrl 定义 grpc 模式下准入的 API 列表，及处理方法
	grpcAccessUrlHandlers map[string]argoGrpcHandler
)

// GrpcPlugin for internal project authorization
type GrpcPlugin struct {
	*mux.Router
	middleware MiddlewareInterface
}

// Init the grpc plugin
// 参见: github.com/argoproj/argocd/v2/cmd/argocd/commands
func (plugin *GrpcPlugin) Init() error {
	grpcAccessUrlHandlers = map[string]argoGrpcHandler{
		"/project.ProjectService/List":                   plugin.handleProjectList,
		"/project.ProjectService/GetDetailedProject":     plugin.handleProjectGet,
		"/repository.RepositoryService/ListRepositories": plugin.handleRepoList,
		"/repository.RepositoryService/Get":              plugin.handleRepoGet,
		"/repository.RepositoryService/ValidateAccess":   plugin.handleRepoAccess,
		"/repository.RepositoryService/CreateRepository": plugin.handleRepoCreate,
		"/repository.RepositoryService/DeleteRepository": plugin.handleRepoDelete,
		"/cluster.ClusterService/List":                   plugin.handleClusterList,
		"/application.ApplicationService/List":           plugin.handleAppList,
		"/application.ApplicationService/Get":            plugin.handleAppGet,
		"/application.ApplicationService/Create":         plugin.handleAppCreate,
		"/application.ApplicationService/Sync":           plugin.handleAppSync,
		"/application.ApplicationService/Delete":         plugin.handleAppDelete,
		// "y/application.ApplicationService/Watch":         nil,
	}
	plugin.Path("").Handler(plugin.middleware.HttpWrapper(plugin.serve))
	return nil
}

// ServeHTTP http handler implementation
func (plugin *GrpcPlugin) serve(ctx context.Context, req *http.Request) *httpResponse {
	if !proxy.IsAdmin(req) {
		return &httpResponse{
			statusCode: http.StatusForbidden,
			err:        fmt.Errorf("request not come from admin"),
		}
	}
	handler, ok := grpcAccessUrlHandlers[strings.TrimPrefix(req.URL.Path, common.GitOpsProxyURL)]
	if !ok {
		return &httpResponse{
			statusCode: http.StatusForbidden,
			err:        fmt.Errorf("request url '%s' unahtourized", req.URL.Path),
		}
	}
	return handler(ctx, req)
}

// parseRequestBytes GRPC 的前 5 位为 header，第 1 位标注是否压缩, 第 2-5 位标注 body 长度。
func (plugin *GrpcPlugin) parseRequestBytes(request []byte) ([]byte, error) {
	if len(request) < 5 {
		return nil, fmt.Errorf("request body %v bytes not over 5", request)
	}
	// NOTE: 默认未压缩，此处不做处理
	isCompressed := request[0]
	blog.V(5).Infof("request compressed: %v", isCompressed)
	bodyBytes := request[1:5]
	bodyLen := binary.BigEndian.Uint32(bodyBytes)
	if len(request) < int(bodyLen+5) {
		return nil, fmt.Errorf("request body %v not normal", request)
	}
	return request[5 : bodyLen+5], nil
}

func (plugin *GrpcPlugin) readRequestBody(req *http.Request, query interface{}) error {
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrapf(err, "read request body failed")
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	body, err := plugin.parseRequestBytes(bs)
	if err != nil {
		return errors.Wrapf(err, "parse request body failed")
	}
	if err = encoding.GetCodec(proto.Name).Unmarshal(body, query); err != nil {
		return errors.Wrapf(err, "unmarshal request body failed")
	}
	return nil
}

// rewriteRequestBody 对于 Application Create 需要重写 request body
func (plugin *GrpcPlugin) rewriteRequestBody(req *http.Request, body interface{}) (int, error) {
	bodyBs, err := encoding.GetCodec(proto.Name).Marshal(body)
	if err != nil {
		return 0, errors.Wrapf(err, "encoding request body failed")
	}
	contentLen := make([]byte, 4)
	binary.BigEndian.PutUint32(contentLen, uint32(len(bodyBs)))
	result := make([]byte, 0, 5+len(bodyBs))
	result = append(result, 0)
	result = append(result, contentLen...)
	result = append(result, bodyBs...)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(result))
	return len(result), nil
}

func (plugin *GrpcPlugin) handleProjectList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        projectList,
	}
}

func (plugin *GrpcPlugin) handleProjectGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &project.ProjectQuery{}
	if err := plugin.readRequestBody(req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, query.Name, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' view permission failed", query.Name),
		}
	}
	return nil
}

func (plugin *GrpcPlugin) handleRepoList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	names := make([]string, 0, len(projectList.Items))
	for _, item := range projectList.Items {
		names = append(names, item.Name)
	}
	repoList, statusCode, err := plugin.middleware.ListRepositories(ctx, names, false)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list repositories failed"),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        repoList,
	}
}

func (plugin *GrpcPlugin) handleRepoGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	repo, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if err != nil {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        repo,
	}
}

func (plugin *GrpcPlugin) handleRepoAccess(ctx context.Context, req *http.Request) *httpResponse {
	repoAccess := &argorepo.RepoAccessQuery{}
	if err := plugin.readRequestBody(req, repoAccess); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if repoAccess.Project == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("create repo request project cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, repoAccess.Project, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' edit permission failed", repoAccess.Project),
		}
	}
	return nil
}

func (plugin *GrpcPlugin) handleRepoCreate(ctx context.Context, req *http.Request) *httpResponse {
	repoCreate := &argorepo.RepoCreateRequest{}
	if err := plugin.readRequestBody(req, repoCreate); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if repoCreate.Repo == nil || repoCreate.Repo.Project == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("create repo request project cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, repoCreate.Repo.Project, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' edit permission failed", repoCreate.Repo.Project),
		}
	}
	return nil
}

func (plugin *GrpcPlugin) handleRepoDelete(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if query.Repo == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("delete repo request repo cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return nil
}

func (plugin *GrpcPlugin) handleClusterList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	names := make([]string, 0, len(projectList.Items))
	for _, item := range projectList.Items {
		names = append(names, item.Name)
	}
	blog.Infof("[requestID=%s] list cluster with projects: %v", ctx.Value(ctxKeyRequestID).(string), names)
	clusters, statusCode, err := plugin.middleware.ListClusters(ctx, names)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list clusters failed"),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        clusters,
	}
}

func (plugin *GrpcPlugin) handleAppList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	names := make([]string, 0, len(projectList.Items))
	for _, item := range projectList.Items {
		names = append(names, item.Name)
	}

	appList, err := plugin.middleware.ListApplications(ctx, names)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusInternalServerError,
			err:        errors.Wrapf(err, "list applications by project '%s' from storage failed", names),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: http.StatusOK,
		obj:        appList,
	}
}

func (plugin *GrpcPlugin) handleAppGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &application.ApplicationQuery{}
	if err := plugin.readRequestBody(req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	app, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, *query.Name, iam.ProjectView)
	if err != nil {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", *query.Name),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: http.StatusOK,
		obj:        app,
	}
}

func (plugin *GrpcPlugin) handleAppCreate(ctx context.Context, req *http.Request) *httpResponse {
	appCreate := &application.ApplicationCreateRequest{}
	if err := plugin.readRequestBody(req, appCreate); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}

	projectName := appCreate.Application.Spec.Project
	if projectName == "" || projectName == "default" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Errorf("project information lost"),
		}
	}
	argoProject, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, projectName,
		iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", projectName),
		}
	}
	// setting control annotations
	if appCreate.Application.Annotations == nil {
		appCreate.Application.Annotations = make(map[string]string)
	}
	appCreate.Application.Annotations[common.ProjectIDKey] =
		common.GetBCSProjectID(argoProject.Annotations)
	appCreate.Application.Annotations[common.ProjectBusinessIDKey] =
		argoProject.Annotations[common.ProjectBusinessIDKey]
	contentLen, err := plugin.rewriteRequestBody(req, appCreate)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Wrapf(err, "rewrite request body failed"),
		}
	}
	req.Header.Set("Content-Length", strconv.Itoa(contentLen))
	req.ContentLength = int64(contentLen)
	return nil
}

func (plugin *GrpcPlugin) handleAppSync(ctx context.Context, req *http.Request) *httpResponse {
	query := &application.ApplicationSyncRequest{}
	if err := plugin.readRequestBody(req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, *query.Name, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", *query.Name),
		}
	}
	return nil
}

func (plugin *GrpcPlugin) handleAppDelete(ctx context.Context, req *http.Request) *httpResponse {
	appDelete := &application.ApplicationDeleteRequest{}
	if err := plugin.readRequestBody(req, appDelete); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, *appDelete.Name, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", *appDelete.Name),
		}
	}
	return nil
}

// handleAppWatch 暂时不处理 watch 动作
func (plugin *GrpcPlugin) handleAppWatch(ctx context.Context, req *http.Request) *httpResponse {
	return nil
}
