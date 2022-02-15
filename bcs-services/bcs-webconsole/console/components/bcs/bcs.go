package bcs

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"

	req "github.com/imroc/req/v3"
	"github.com/pkg/errors"
)

var client = req.C().SetTimeout(5 * time.Second).DevMode()

type Cluster struct {
	ClusterId   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
}

type Result struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []*Cluster `json:"data"`
}

func ListClusters(ctx context.Context, projectId string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", config.G.BCS.Host)
	resp, err := client.R().SetBearerAuthToken(config.G.BCS.Token).SetQueryParam("projectID", projectId).Get(url)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := resp.Unmarshal(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(fmt.Sprintf("query clustermanager error, %s", result.Message))
	}

	var clusters []*Cluster
	for _, cluster := range result.Data {
		// 过滤掉共享集群
		if cluster.IsShared {
			continue
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
