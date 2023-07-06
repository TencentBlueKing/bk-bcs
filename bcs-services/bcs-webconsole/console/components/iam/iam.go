// Package iam xxx
package iam

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
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
func IsAllowedWithResource(ctx context.Context, projectId, clusterId, namespaceName, username string) (bool, error) {
	logger.Infof("auth with iam, projectId=%s, clusterId=%s, namespace=%s, username=%s", projectId, clusterId,
		namespaceName, username)

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

	relatedActionIDs := []string{project.ProjectView.String()}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectId, Action: project.ProjectView.String()},
	}

	nodes := [][]iam.ResourceNode{projectNode}

	module := project.BCSProjectModule
	operator := project.CanViewProjectOperation

	// 集群查看权限
	if clusterId != "" {
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resources = append(resources, utils.ResourceAction{Resource: clusterId, Action: cluster.ClusterView.String()})
		clusterNode := cluster.ClusterResourceNode{
			IsCreateCluster: false,
			SystemID:        iam.SystemIDBKBCS,
			ProjectID:       projectId,
			ClusterID:       clusterId}.BuildResourceNodes()
		nodes = append(nodes, clusterNode)
		module = cluster.BCSClusterModule
		operator = cluster.CanViewClusterOperation
	}

	if namespaceName != "" {
		nameSpaceID, err := calcNamespaceID(clusterId, namespaceName)
		if err != nil {
			return false, err
		}
		relatedActionIDs = append(relatedActionIDs, namespace.NameSpaceScopedCreate.String())
		resources = append(resources, utils.ResourceAction{Resource: nameSpaceID,
			Action: namespace.NameSpaceScopedCreate.String()})
		namespaceNode := iam.ResourceNode{
			System:    iam.SystemIDBKBCS,
			RType:     string(namespace.SysNamespace),
			RInstance: nameSpaceID,
			Rp: namespace.NamespaceScopedResourcePath{
				ProjectID: projectId,
				ClusterID: clusterId,
			},
		}
		nodes = append(nodes, []iam.ResourceNode{namespaceNode})

		// 只做日志使用
		module = string(namespace.SysNamespace)
		operator = namespace.NameSpaceScopedCreate.String()
	}

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
func MakeResourceApplyUrl(ctx context.Context, projectId, clusterId, namespaceName, username string) (string, error) {
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

	apps := []iam.ApplicationAction{projectApp}

	// 申请集群查看权限
	if clusterId != "" {
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
		apps = append(apps, clusterApp)
	}

	// 命名空间域创建权限
	// 和bcs-api校验权限一致(POST请求, 命名空间下的资源)
	if namespaceName != "" {
		nameSpaceID, err := calcNamespaceID(clusterId, namespaceName)
		if err != nil {
			return "", err
		}
		namespaceApp := iam.ApplicationAction{
			ActionID: namespace.NameSpaceScopedCreate.String(),
			RelatedResources: []bkiam.ApplicationRelatedResourceType{
				{
					SystemID: iam.SystemIDBKBCS,
					Type:     string(namespace.SysNamespace),
					Instances: []bkiam.ApplicationResourceInstance{
						[]bkiam.ApplicationResourceNode{
							{
								Type: string(project.SysProject),
								ID:   projectId,
							},
							{
								Type: string(cluster.SysCluster),
								ID:   clusterId,
							},
							{
								Type: string(namespace.SysNamespace),
								ID:   nameSpaceID,
							},
						},
					},
				},
			},
		}
		apps = append(apps, namespaceApp)
	}

	applyUrl, err := iamClient.GetApplyURL(req, apps, user)
	return applyUrl, err
}

// md5Digest 字符串转 MD5
func md5Digest(key string) string {
	hash := md5.New()
	hash.Write([]byte(key))
	return hex.EncodeToString(hash.Sum(nil))
}

// calcNamespaceID 计算(压缩)出注册到权限中心的命名空间 ID，具备唯一性. 当前的算法并不能完全避免冲突，但概率较低。
// note: 权限中心对资源 ID 有长度限制，不超过32位。长度越长，处理性能越低
// NamespaceID 是命名空间注册到权限中心的资源 ID，它是对结构`集群ID:命名空间name`的一个压缩，
// 如 `BCS-K8S-10000:default` 会被处理成 `10000:5f03d33dde`。
func calcNamespaceID(clusterID string, name string) (string, error) {
	clusterStrs := strings.Split(clusterID, "-")
	if len(clusterStrs) != 3 {
		return "", fmt.Errorf("calcNamespaceID err: %v", "length not equal 3")
	}
	clusterIDx := clusterStrs[len(clusterStrs)-1]

	iamNsID := clusterIDx + ":" + md5Digest(name)[8:16] + name[:2]
	if len(iamNsID) > 32 {
		return "", fmt.Errorf("calcNamespaceID iamNamespaceID more than 32characters")
	}

	return iamNsID, nil
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

	cacheKey := "iam.GetAccessToken"
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
		SetBody(jsonData).
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
