package iam

import (
	"context"
	"fmt"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

func newIAMClient() (iam.PermClient, error) {
	var opts = &iam.Options{
		SystemID:  iam.SystemIDBKBCS,
		AppCode:   config.G.Base.AppCode,
		AppSecret: config.G.Base.AppSecret,
		Metric:    false,
		Debug:     config.G.IsDevMode(),
	}

	// 使用网关地址
	if config.G.Auth.IsGatewWay {
		opts.GateWayHost = config.G.Auth.Host
		opts.External = false
	} else {
		// 使用"外部" ingress 地址
		opts.IAMHost = config.G.Auth.Host
		opts.BkiIAMHost = config.G.Base.BKPaaSHost
		opts.External = true
	}

	client, err := iam.NewIamClient(opts)
	return client, err
}

// IsAllowedWithResource 校验项目, 集群是否有权限
func IsAllowedWithResource(ctx context.Context, projectId, clusterId, username string) (bool, error) {
	logger.Infof("auth with iam, projectId=%s, clusterId=%s, username=%s", projectId, clusterId, username)

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

// accessToken 返回
type accessToken struct {
	AccessToken string `json:"access_token"`
}

// GetAccessToken 获取 accessToken
func GetAccessToken(ctx context.Context) (string, error) {
	// 兼容逻辑, 如果不配置SSMHost, 不使用 access_token 鉴权
	if config.G.Auth.SSMHost == "" {
		return "", nil
	}

	cacheKey := fmt.Sprintf("iam.GetAccessToken:%s", config.G.BCSCC.Stage)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.(*accessToken).AccessToken, nil
	}

	url := fmt.Sprintf("%s/api/v1/auth/access-tokens", config.G.Auth.SSMHost)

	jsonData := map[string]string{
		"grant_type":  "client_credentials",
		"id_provider": "client",
	}

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-BK-APP-CODE", config.G.Base.AppCode).
		SetHeader("X-BK-APP-SECRET", config.G.Base.AppSecret).
		SetBodyJsonMarshal(jsonData).
		Post(url)

	if err != nil {
		return "", err
	}

	var token *accessToken
	if err := components.UnmarshalBKResult(resp, &token); err != nil {
		return "", err
	}

	storage.LocalCache.Slot.Set(cacheKey, token, time.Minute*5)

	return token.AccessToken, nil
}
