package iam

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

const (
	// SysNamespace resource namespace
	Project iam.TypeID = "project"
	Cluster iam.TypeID = "cluster"
)

func IsAllowedWithResource(ctx context.Context, projectId, clusterId, username string) (bool, error) {
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

	rn := []iam.ResourceNode{
		{
			System:    iam.SystemIDBKBCS,
			RType:     string(Cluster),
			RInstance: clusterId,
			Rp: cluster.ClusterScopedResourcePath{
				ProjectID: projectId,
			},
		},
	}
	allow, err := client.IsAllowedWithResource("cluster_view", req, rn, false)
	if err != nil {
		return false, err
	}

	return allow, nil
}

// MakeApplyUrl 权限中心申请URL
func MakeApplyUrl(ctx context.Context, projectId, clusterId, username string) (string, error) {
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
