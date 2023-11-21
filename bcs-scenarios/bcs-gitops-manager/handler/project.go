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

package handler

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

func (e *BcsGitopsHandler) startProjectResult(resp *pb.GitOpsResponse, code errorCode,
	message string, err error) error {
	resp.Code = int32(code)
	if code == successCode {
		resp.Message = message
		return nil
	}
	blog.Errorf("startupProject failed: %s", err.Error())
	resp.Error = err.Error()
	return err
}

func (e *BcsGitopsHandler) checkStartupProjectPermission(ctx context.Context, projectID string) error {
	raw := metautils.ExtractIncoming(ctx).Get("Authorization")
	user, err := proxy.GetJWTInfoWithAuthorization(raw, e.option.JwtClient)
	if err != nil {
		return errors.Wrapf(err, "get userinfo failed")
	}
	permit, _, _, err := e.projectPermission.CanEditProject(user.GetUser(), projectID)
	if err != nil {
		return errors.Wrapf(err, "check user '%s' can edit project '%s' failed", user.GetUser(), projectID)
	}
	if !permit {
		return errors.Errorf("user '%s' not allowed edit project '%s'", user.GetUser(), projectID)
	}
	return nil
}

// StartupProject will start the project and sync all the clusters
func (e *BcsGitopsHandler) StartupProject(ctx context.Context, req *pb.ProjectSyncRequest,
	resp *pb.GitOpsResponse) error {
	if err := req.Validate(); err != nil {
		return e.startProjectResult(resp, failedCode, "",
			errors.Wrapf(err, "request is not validate"))
	}
	blog.Infof("prepared to sync project info: %s", req.GetProjectCode())
	project, err := e.option.ProjectControl.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		return e.startProjectResult(resp, failedCode, "",
			errors.Wrapf(err, "request project '%s' from bcs-project failed", req.GetProjectCode()))
	}
	// check the user whether have the project edit permission
	if err := e.checkStartupProjectPermission(ctx, project.ProjectID); err != nil {
		return e.startProjectResult(resp, failedCode, "",
			errors.Wrapf(err, "check startup project '%s' permission failed", req.GetProjectCode()))
	}

	// check argocd storage
	destPro, err := e.option.Storage.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		return e.startProjectResult(resp, failedCode, "",
			errors.Wrapf(err, "request project '%s' from gitops storage failed", req.GetProjectCode()))
	}
	if destPro != nil {
		return e.startProjectResult(resp, successCode,
			fmt.Sprintf("project '%s' already startup", req.GetProjectCode()), nil)
	}
	// save to AppProject
	destPro = defaultAppProject(e.option.AdminNamespace, project)
	// add secret info annotation
	secretAnnotation, err := e.option.SecretClient.GetProjectSecret(ctx, project.Name)
	if err != nil {
		blog.Errorf("[getErr]sync secret info to pro annotations [%s] failed: %s", project.ProjectCode, err.Error())
	} else {
		destPro.ObjectMeta.Annotations[common.SecretKey] = secretAnnotation
	}

	if err := e.option.Storage.CreateProject(ctx, destPro); err != nil {
		return e.startProjectResult(resp, failedCode, "",
			errors.Wrapf(err, "create project '%s' to storage failed", project.ProjectCode))
	}
	if err := e.option.ClusterControl.SyncProject(ctx, project.ProjectCode); err != nil {
		return e.startProjectResult(resp, failedCode, "",
			errors.Wrapf(err, "sync project '%s' clusters failed", project.ProjectCode))
	}

	// 完成project同步之后，需要初始化secret vault相关信息
	if err = e.option.SecretClient.InitProjectSecret(ctx, project.ProjectCode); err != nil {
		if !utils.IsSecretAlreadyExist(err) {
			return e.startProjectResult(resp, failedCode, "",
				errors.Wrapf(err, "init secret project '%s' failed", project.ProjectCode))
		}
	}
	return e.startProjectResult(resp, successCode, "ok", nil)
}

func defaultAppProject(ns string, project *bcsproject.Project) *v1alpha1.AppProject {
	return &v1alpha1.AppProject{
		TypeMeta: v1.TypeMeta{
			Kind:       "AppProject",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      project.ProjectCode,
			Namespace: ns,
			Annotations: map[string]string{
				common.ProjectAliaName:      project.Name,
				common.ProjectIDKey:         project.ProjectID,
				common.ProjectBusinessIDKey: project.BusinessID,
				common.ProjectBusinessName:  project.BusinessName,
			},
		},
		Spec: v1alpha1.AppProjectSpec{
			ClusterResourceWhitelist: []v1.GroupKind{{Group: "*", Kind: "*"}},
			Destinations:             []v1alpha1.ApplicationDestination{{Server: "*", Namespace: "*"}},
			SourceRepos:              []string{"*"},
			Description:              project.Description,
		},
	}
}
