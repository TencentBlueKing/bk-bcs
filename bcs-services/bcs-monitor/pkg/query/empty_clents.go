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

/*
	此包中实现thanos的metadata，target，rule，exemplar 的proxy接口
	替换掉query模块中的各种各个用不到的模块
*/

// NotImplementErr
var NotImplementErr = errors.New("api not implement")

// NewEmptyMetaDataClient
func NewEmptyMetaDataClient() *emptyMetadataClient {
	return &emptyMetadataClient{}
}

// NewEmptyTargetClient
func NewEmptyTargetClient() *emptyTargetClient {
	return &emptyTargetClient{}
}

// NewEmptyRuleClient
func NewEmptyRuleClient() *emptyRuleClient {
	return &emptyRuleClient{}
}

// NewEmptyExemplarClient
func NewEmptyExemplarClient() *emptyExemplarClient {
	return &emptyExemplarClient{}
}

type emptyMetadataClient struct{}

// MetricMetadata
func (e *emptyMetadataClient) MetricMetadata(_ context.Context, _ *metadatapb.MetricMetadataRequest) (map[string][]metadatapb.Meta, storage.Warnings, error) {
	return nil, []error{NotImplementErr}, nil
}

type emptyTargetClient struct{}

// Targets
func (e *emptyTargetClient) Targets(_ context.Context, _ *targetspb.TargetsRequest) (*targetspb.TargetDiscovery, storage.Warnings, error) {
	return nil, []error{NotImplementErr}, nil
}

type emptyRuleClient struct{}

// Rules
func (e *emptyRuleClient) Rules(_ context.Context, _ *rulespb.RulesRequest) (*rulespb.RuleGroups, storage.Warnings, error) {
	return nil, []error{NotImplementErr}, nil
}

type emptyExemplarClient struct{}

// Exemplars
func (e *emptyExemplarClient) Exemplars(_ context.Context, _ *exemplarspb.ExemplarsRequest) ([]*exemplarspb.ExemplarData, storage.Warnings, error) {
	return nil, []error{NotImplementErr}, nil
}
