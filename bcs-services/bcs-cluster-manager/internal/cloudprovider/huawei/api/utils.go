package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	iamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	Zones = map[string][]string{
		//俄罗斯-莫斯科二
		"ru-northwest-2": {"ru-northwest-2a", "ru-northwest-2b", "ru-northwest-2c"},
		//非洲-约翰内斯堡
		"af-south-1": {"af-south-1a", "af-south-1b"},
		//华北-北京四
		"cn-north-4": {"cn-north-4a", "cn-north-4b", "cn-north-4c", "cn-north-4g"},
		//华北-北京一
		"cn-north-1": {"cn-north-1a", "cn-north-1b", "cn-north-1c"},
		//华北-乌兰察布一
		"cn-north-9": {"cn-north-9a", "cn-north-9b"},
		//华东-上海二
		"cn-east-2": {"cn-east-2a", "cn-east-2b", "cn-east-2c", "cn-east-2d"},
		//华东-上海一
		"cn-east-3": {"cn-east-3a", "cn-east-3b", "cn-east-3c"},
		//华南-广州
		"cn-south-1": {"cn-south-1a", "cn-south-2b", "cn-south-1c", "cn-south-1e", "cn-south-1f"},
		//拉美-布宜诺斯艾利斯一
		"sa-argentina-1": {"sa-argentina-1a"},
		//拉美-利马一
		"sa-peru-1": {"sa-peru-1a", "sa-peru-1b"},
		//华南-广州-友好用户环境
		"cn-south-4": {"cn-south-4a", "cn-south-4b", "cn-south-4c"},
		//华南-深圳
		"cn-south-2": {"cn-south-2a"},
		//拉美-墨西哥城二
		"la-north-2": {"la-north-2a", "la-north-2c"},
		//拉美-墨西哥城一
		"na-mexico-1": {"na-mexico-1a", "na-mexico-1c"},
		//拉美-圣地亚哥
		"la-south-2": {"la-south-2a"},
		//欧洲-巴黎
		"eu-west-0": {"eu-west-0a", "eu-west-0b", "eu-west-0c"},
		//欧洲-都柏林
		"eu-west-101": {"eu-west-101a", "eu-west-101b"},
		//土耳其-伊斯坦布尔
		"tr-west-1": {"tr-west-1a", "tr-west-1b", "tr-west-1c"},
		//西南-贵阳一
		"cn-southwest-2": {"cn-southwest-2a", "cn-southwest-2b", "cn-southwest-2c", "cn-southwest-2d",
			"cn-southwest-2e", "cn-southwest-2f"},
		//亚太-曼谷
		"ap-southeast-2": {"ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"},
		//亚太-新加坡
		"ap-southeast-3": {"ap-southeast-3a", "ap-southeast-3b", "ap-southeast-3c"},
		//亚太-雅加达
		"ap-southeast-4": {"ap-southeast-4a", "ap-southeast-4b", "ap-southeast-4c"},
		//中东-利雅得
		"me-east-1": {"me-east-1a"},
		//中国-香港
		"ap-southeast-1": {"ap-southeast-1a", "ap-southeast-1b"},
	}
)

// GetInternalClusterKubeConfig get cce cluster kebeconfig
func GetInternalClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			Duration: int32(-1), // 集群证书有效时间，单位为天，最小值为1，最大值为10950(30*365，1年固定计365天，忽略闰年影响)；若填-1则为最大值30年。
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	currentContext := "internal"
	clusters := make([]model.Clusters, 0)
	contexts := make([]model.Contexts, 0)

	for _, v := range *rsp.Clusters {
		if *v.Name == "internalCluster" {
			clusters = append(clusters, v)
		}
	}

	if len(clusters) == 0 {
		return "", fmt.Errorf("internal cluster not found")
	}

	for _, v := range *rsp.Contexts {
		if *v.Name == "internal" {
			contexts = append(contexts, v)
		}
	}

	if len(contexts) == 0 {
		return "", fmt.Errorf("internal context not found")
	}

	kubeCfg := &model.CreateKubernetesClusterCertResponse{
		Kind:           rsp.Kind,
		ApiVersion:     rsp.ApiVersion,
		Preferences:    rsp.Preferences,
		Clusters:       &clusters,
		Users:          rsp.Users,
		Contexts:       &contexts,
		CurrentContext: &currentContext,
		PortID:         rsp.PortID,
	}

	bt, err := json.Marshal(kubeCfg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetClusterKubeConfig get cce cluster kebeconfig
func GetClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			Duration: int32(-1), // 集群证书有效时间，单位为天，最小值为1，最大值为10950(30*365，1年固定计365天，忽略闰年影响)；若填-1则为最大值30年。
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	kubeCfg := &model.CreateKubernetesClusterCertResponse{}
	if len(*rsp.Clusters) == 1 {
		kubeCfg = rsp
	} else if len(*rsp.Clusters) > 1 && *rsp.CurrentContext == "external" {
		curContext := "externalTLSVerify"
		clusters := make([]model.Clusters, 0)
		contexts := make([]model.Contexts, 0)

		for _, v := range *rsp.Clusters {
			if *v.Name == "externalClusterTLSVerify" {
				clusters = append(clusters, v)
			}
		}

		for _, v := range *rsp.Contexts {
			if *v.Name == "externalTLSVerify" {
				contexts = append(contexts, v)
			}
		}
		kubeCfg = &model.CreateKubernetesClusterCertResponse{
			Kind:           rsp.Kind,
			ApiVersion:     rsp.ApiVersion,
			Preferences:    rsp.Preferences,
			Clusters:       &clusters,
			Users:          rsp.Users,
			Contexts:       &contexts,
			CurrentContext: &curContext,
			PortID:         rsp.PortID,
		}
	}

	bt, err := json.Marshal(kubeCfg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetProjectIDByRegion get project ID by region
func GetProjectIDByRegion(opt *cloudprovider.CommonOption) (string, error) {
	client, err := GetIamClient(opt)
	if err != nil {
		return "", err
	}

	req := iamModel.KeystoneListProjectsRequest{Name: &opt.Region}
	rsp, err := client.KeystoneListProjects(&req)
	if err != nil {
		return "", err
	}

	if len(*rsp.Projects) == 0 {
		return "", fmt.Errorf("project not found")
	} else if len(*rsp.Projects) > 1 {
		return "", fmt.Errorf("the number of project is greater than one")
	}

	return (*rsp.Projects)[0].Id, nil
}
