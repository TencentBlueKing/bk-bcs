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
 *
 */

// Package bcs xxx
package bcs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
)

var tracer = otel.Tracer("bcs_client")

// Cluster 集群信息
type Cluster struct {
	ProjectId   string `json:"projectID"`
	ClusterId   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	BKBizID     string `json:"businessID"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
}

// String :
func (c *Cluster) String() string {
	return fmt.Sprintf("cluster<%s, %s>", c.ClusterName, c.ClusterId)
}

// ListClusters 获取项目集群列表
func ListClusters(ctx context.Context, bcsConf *config.BCSConf, projectId string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", bcsConf.Host)

	commonAttrs := []attribute.KeyValue{
		attribute.String("url", url),
		attribute.String("projectId", projectId),
	}
	ctx, span := tracer.Start(ctx, "ListClusters", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(bcsConf.Token).
		SetQueryParam("projectID", projectId).
		Get(url)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var result []*Cluster
	if err := component.UnmarshalBKResult(resp, &result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	clusters := make([]*Cluster, 0, len(result))
	for _, cluster := range result {
		// 过滤掉共享集群
		if cluster.IsShared {
			continue
		}
		clusters = append(clusters, cluster)
	}
	clustersStr, _ := json.Marshal(clusters)
	// 设置额外标签
	span.SetAttributes(attribute.String("clusters", string(clustersStr)))
	return clusters, nil
}

// GetClusterMap 获取全部集群数据, map格式
func GetClusterMap(ctx context.Context, bcsConf *config.BCSConf) (map[string]*Cluster, error) {
	cacheKey := fmt.Sprintf("bcs.GetClusterMap:%s", bcsConf.ClusterEnv)
	commonAttrs := []attribute.KeyValue{
		attribute.String("cacheKey", cacheKey),
	}
	ctx, span := tracer.Start(ctx, "GetClusterMap", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()

	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		resultStr, _ := json.Marshal(cacheResult)
		span.SetAttributes(attribute.Key("cacheResult").String(string(resultStr)))
		return cacheResult.(map[string]*Cluster), nil
	}

	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", bcsConf.Host)
	span.SetAttributes(attribute.Key("url").String(url))
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(bcsConf.Token).
		Get(url)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var result []*Cluster
	if err := component.UnmarshalBKResult(resp, &result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	clusterMap := map[string]*Cluster{}
	for _, cluster := range result {
		// 过滤掉共享集群
		if cluster.IsShared {
			continue
		}
		// 集群状态 https://github.com/Tencent/bk-bcs/blob/master/bcs-services/bcs-cluster-manager/api/clustermanager/clustermanager.proto#L1003
		if cluster.Status != "RUNNING" {
			continue
		}
		clusterMap[cluster.ClusterId] = cluster
	}

	storage.LocalCache.Slot.Set(cacheKey, clusterMap, time.Minute*10)

	clusterMapStr, _ := json.Marshal(clusterMap)
	span.SetAttributes(attribute.Key("clusterMap").String(string(clusterMapStr)))
	return clusterMap, nil
}

// GetCluster 获取集群详情
func GetCluster(ctx context.Context, bcsConf *config.BCSConf, clusterId string) (*Cluster, error) {
	cacheKey := fmt.Sprintf("bcs.GetCluster:%s", clusterId)
	commonAttrs := []attribute.KeyValue{
		attribute.String("cacheKey", cacheKey),
	}
	ctx, span := tracer.Start(ctx, "GetCluster", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		resultStr, _ := json.Marshal(cacheResult)
		span.SetAttributes(attribute.Key("cacheResult").String(string(resultStr)))
		return cacheResult.(*Cluster), nil
	}

	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster/%s", bcsConf.Host, clusterId)
	span.SetAttributes(attribute.Key("url").String(url))
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(bcsConf.Token).
		Get(url)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var cluster *Cluster
	if err := component.UnmarshalBKResult(resp, &cluster); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	storage.LocalCache.Slot.Set(cacheKey, cluster, storage.LocalCache.DefaultExpiration)
	clusterStr, _ := json.Marshal(cluster)
	span.SetAttributes(attribute.Key("cluster").String(string(clusterStr)))
	return cluster, nil
}
