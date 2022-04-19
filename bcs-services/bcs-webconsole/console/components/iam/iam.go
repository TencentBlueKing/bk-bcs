package iam

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

func newIAMClient() (iam.PermClient, error) {
	var opts = &iam.Options{
		SystemID:    iam.SystemIDBKBCS,
		AppCode:     config.G.Base.AppCode,
		AppSecret:   config.G.Base.AppSecret,
		External:    false,
		GateWayHost: config.G.Auth.Host,
		Metric:      false,
		Debug:       config.G.IsDevMode(),
	}

	client, err := iam.NewIamClient(opts)
	return client, err
}

// IsAllowedWithResource 校验项目, 集群是否有权限
func IsAllowedWithResource(ctx context.Context, projectId, clusterId, username string) (bool, error) {
	iamClient, err := newIAMClient()
	if err != nil {
		return false, err
	}

	req := iam.PermissionRequest{SystemID: iam.SystemIDBKBCS, UserName: username}

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

	relatedActionIDs := []string{project.ProjectView.String()}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectId, Action: project.ProjectView.String()},
	}

	nodes := [][]iam.ResourceNode{projectNode}

	module := project.BCSProjectModule
	operator := project.CanViewProjectOperation

	if clusterId != "" {
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resources = append(resources, utils.ResourceAction{Resource: clusterId, Action: cluster.ClusterView.String()})
		nodes = append(nodes, clusterNode)
		module = cluster.BCSClusterModule
		operator = cluster.CanViewClusterOperation
	}

	// 集群查看权限
	perms, err := iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, nodes)
	if err != nil {
		return false, err
	}

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    module,
		Operation: operator,
		User:      username,
	}, resources, perms)

	return allow, err
}

// MakeResourceApplyUrl 权限中心申请URL
func MakeResourceApplyUrl(ctx context.Context, projectId, clusterId, username string) (string, error) {
	iamClient, err := newIAMClient()
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

	apps := []iam.ApplicationAction{projectApp}
	if clusterId != "" {
		apps = append(apps, clusterApp)
	}

	applyUrl, err := iamClient.GetApplyURL(req, apps, user)
	return applyUrl, err
}
