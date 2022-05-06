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

var NotImplementErr = errors.New("api not implement")

func NewEmptyMetaDataClient() *emptyMetadataClient {
	return &emptyMetadataClient{}
}

func NewEmptyTargetClient() *emptyTargetClient {
	return &emptyTargetClient{}
}

func NewEmptyRuleClient() *emptyRuleClient {
	return &emptyRuleClient{}
}

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

func (e *emptyRuleClient) Rules(_ context.Context, _ *rulespb.RulesRequest) (*rulespb.RuleGroups, storage.Warnings, error) {
	return nil, []error{NotImplementErr}, nil
}

type emptyExemplarClient struct{}

func (e *emptyExemplarClient) Exemplars(_ context.Context, _ *exemplarspb.ExemplarsRequest) ([]*exemplarspb.ExemplarData, storage.Warnings, error) {
	return nil, []error{NotImplementErr}, nil
}
