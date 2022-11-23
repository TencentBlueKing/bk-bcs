package handler

import (
	"context"
	"fmt"

	v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

// StartupProject implementation
func (e *BcsGitopsHandler) StartupProject(ctx context.Context,
	req *pb.ProjectSyncRequest, rsp *pb.GitOpsResponse) error {
	if err := req.Validate(); err != nil {
		blog.Errorf("request is not validate, %s", err.Error())
		rsp.Code = -1
		rsp.Error = err.Error()
		rsp.Message = "request is not validate"
		return err
	}
	blog.Infof("prepared to sync project info: %s", req.GetProjectCode())
	project, err := e.option.ProjectControl.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		blog.Errorf("bcs-project-manager get project %s faileure", req.GetProjectCode())
		rsp.Code = -1
		rsp.Error = err.Error()
		rsp.Message = "bcs-project request failure"
		return err
	}
	// check local storage
	destPro, err := e.option.Storage.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		blog.Errorf("get local gitops Project %s information failure, %s", req.GetProjectCode(), err.Error())
		rsp.Code = -1
		rsp.Error = err.Error()
		rsp.Message = "gitops storage failure"
		return err
	}
	if destPro != nil {
		blog.Errorf("Project %s information already startup", req.GetProjectCode())
		rsp.Code = 0
		rsp.Message = fmt.Sprintf("project %s already startup", req.GetProjectCode())
		return nil
	}
	// ready to setting bcs project to AppProject
	destPro = defaultAppProject(e.option.AdminNamespace, project)

	// save to AppProject
	if err := e.option.Storage.CreateProject(ctx, destPro); err != nil {
		blog.Errorf("startup project %s failed, %s, project details: %+v",
			req.ProjectCode, err.Error(), destPro)
		rsp.Code = -1
		rsp.Error = err.Error()
		rsp.Message = "create project to storage failure"
		return err
	}
	rsp.Code = 0
	rsp.Message = "ok"
	return nil
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
				common.ProjectIDKey:         project.ProjectID,
				common.ProjectBusinessIDKey: project.BusinessID,
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
