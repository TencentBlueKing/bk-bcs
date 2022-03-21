package iam

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	authIAM "github.com/TencentBlueKing/iam-go-sdk"
)

const (
	// SysNamespace resource namespace
	Project iam.TypeID = "project"
	Cluster iam.TypeID = "cluster"
)

func IsAllowedWithResource(ctx context.Context, projectId, username string) (bool, error) {
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

	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: username,
	}
	rn := []iam.ResourceNode{
		{
			System:    iam.SystemIDBKBCS,
			RType:     string(Project),
			RInstance: projectId,
			Rp: cluster.ClusterScopedResourcePath{
				ProjectID: projectId,
			},
		},
	}
	allow, err := client.IsAllowedWithResource("project_view", req, rn, false)
	if err != nil {
		return false, err
	}

	return allow, nil
}

// ApplyUrl 权限中心申请URL
func ApplyUrl(ctx context.Context, projectId string) (string, error) {
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
	res := []authIAM.ApplicationRelatedResourceType{
		{
			SystemID: iam.SystemIDBKBCS,
			Type:     "project",
			Instances: []authIAM.ApplicationResourceInstance{
				{
					authIAM.ApplicationResourceNode{
						Type: "project",
						ID:   projectId,
					},
				},
			},
		},
	}

	actionApplication3 := []iam.ApplicationAction{
		{
			ActionID:         "project_view",
			RelatedResources: res,
		},
	}

	applyUrl, err := client.GetApplyURL(iam.ApplicationRequest{SystemID: iam.SystemIDBKBCS}, actionApplication3, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	return applyUrl, err
}
