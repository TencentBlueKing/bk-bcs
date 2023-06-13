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

package query

import (
	"context"
	"errors"

	"github.com/prometheus/prometheus/storage"
	"github.com/thanos-io/thanos/pkg/exemplars/exemplarspb"
	"github.com/thanos-io/thanos/pkg/metadata/metadatapb"
	"github.com/thanos-io/thanos/pkg/rules/rulespb"
	"github.com/thanos-io/thanos/pkg/targets/targetspb"
)

// 此包中实现thanos的metadata，target，rule，exemplar 的proxy接口 替换掉query模块中的各种各个用不到的模块

// ErrNotImplement xxx
var ErrNotImplement = errors.New("api not implement")

// NewEmptyMetaDataClient xxx
func NewEmptyMetaDataClient() *EmptyMetadataClient {
	return &EmptyMetadataClient{}
}

// NewEmptyTargetClient :
func NewEmptyTargetClient() *EmptyTargetClient {
	return &EmptyTargetClient{}
}

// NewEmptyRuleClient :
func NewEmptyRuleClient() *EmptyRuleClient {
	return &EmptyRuleClient{}
}

// NewEmptyExemplarClient :
func NewEmptyExemplarClient() *EmptyExemplarClient {
	return &EmptyExemplarClient{}
}

// EmptyMetadataClient empty metadata client
type EmptyMetadataClient struct{}

// MetricMetadata :
func (e *EmptyMetadataClient) MetricMetadata(_ context.Context, _ *metadatapb.MetricMetadataRequest) (
	map[string][]metadatapb.Meta, storage.Warnings, error) {
	return nil, []error{ErrNotImplement}, nil
}

// EmptyTargetClient empty target client
type EmptyTargetClient struct{}

// Targets :
func (e *EmptyTargetClient) Targets(_ context.Context, _ *targetspb.TargetsRequest) (*targetspb.TargetDiscovery,
	storage.Warnings, error) {
	return nil, []error{ErrNotImplement}, nil
}

// EmptyRuleClient empty rule client
type EmptyRuleClient struct{}

// Rules :
func (e *EmptyRuleClient) Rules(_ context.Context, _ *rulespb.RulesRequest) (*rulespb.RuleGroups, storage.Warnings,
	error) {
	return nil, []error{ErrNotImplement}, nil
}

// EmptyExemplarClient empty exemplar client
type EmptyExemplarClient struct{}

// Exemplars :
func (e *EmptyExemplarClient) Exemplars(_ context.Context, _ *exemplarspb.ExemplarsRequest) (
	[]*exemplarspb.ExemplarData, storage.Warnings, error) {
	return nil, []error{ErrNotImplement}, nil
}
