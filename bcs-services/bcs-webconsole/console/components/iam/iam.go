package iam

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// IsAllowedWithCluster 校验项目, 集群是否有权限
func IsAllowedWithCluster(ctx context.Context, projectId, clusterId, username string) (bool, error) {
	var opts = &iam.Options{
		SystemID:    iam.SystemIDBKBCS,
		AppCode:     config.G.Base.AppCode,
		AppSecret:   config.G.Base.AppSecret,
		External:    false,
		GateWayHost: config.G.Auth.Host,
		Metric:      false,
		Debug:       true,
	}

	client, err := iam.NewIamClient(opts)
	if err != nil {
		return false, err
	}

	req := iam.PermissionRequest{SystemID: iam.SystemIDBKBCS, UserName: username}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: clusterId, Action: cluster.ClusterView.String()},
		{Resource: projectId, Action: project.ProjectView.String()},
	}

	relatedActionIDs := []string{project.ProjectView.String(), cluster.ClusterView.String()}

	// 项目查看权限
	projectNode := project.ProjectResourceNode{
		IsCreateProject: false,
		SystemID:        iam.SystemIDBKBCS,
		ProjectID:       projectId}.BuildResourceNodes()

	clusterNode := cluster.ClusterResourceNode{
		IsCreateCluster: false,
		SystemID:        iam.SystemIDBKBCS,
		ProjectID:       projectId,
		ClusterID:       clusterId}.BuildResourceNodes()

	// 集群查看权限
	perms, err := client.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{projectNode, clusterNode})
	if err != nil {
		return false, err
	}

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    cluster.BCSClusterModule,
		Operation: cluster.CanViewClusterOperation,
		User:      username,
	}, resources, perms)

	return allow, err
}

// MakeClusterApplyUrl 权限中心申请URL
func MakeClusterApplyUrl(ctx context.Context, projectId, clusterId, username string) (string, error) {
	var opts = &iam.Options{
		SystemID:    iam.SystemIDBKBCS,
		AppCode:     config.G.Base.AppCode,
		AppSecret:   config.G.Base.AppSecret,
		External:    false,
		GateWayHost: config.G.Auth.Host,
		Metric:      false,
		Debug:       true,
	}

	client, err := iam.NewIamClient(opts)
	if err != nil {
		return "", err
	}

	if username == "" {
		username = iam.SystemUser
	}

	req := iam.ApplicationRequest{SystemID: iam.SystemIDBKBCS}
	user := iam.BkUser{BkUserName: username}

	// 申请项目查看权限
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectId},
	})

	// 申请集群查看权限
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		IsCreateCluster: false,
		ActionID:        cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{
				Project: projectId,
				Cluster: clusterId,
			},
		},
	})

	apps := []iam.ApplicationAction{projectApp, clusterApp}

	applyUrl, err := client.GetApplyURL(req, apps, user)
	return applyUrl, err
}
