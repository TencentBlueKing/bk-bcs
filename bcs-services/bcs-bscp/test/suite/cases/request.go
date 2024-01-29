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
 */

package cases

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/filter"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

// GenListAppByIdsReq generate a list_app request by biz_id and app_ids.
func GenListAppByIdsReq(bizId uint32, appIds []uint32) (*pbcs.ListAppsReq, error) {
	pbStruct, err := GenQueryFilterByIds(appIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListAppsReq{
		BizId:  bizId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListHookByIdsReq generate a list_hook request by app_ids.
func GenListHookByIdsReq(appId uint32, hookIds []uint32) (*pbcs.ListHooksReq, error) {
	pbStruct, err := GenQueryFilterByIds(hookIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListHooksReq{
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListConfigItemByIdsReq generate a list_config_item request by biz_id and app_ids.
//func GenListConfigItemByIdsReq(bizId, appId uint32, configItemIds []uint32) (*pbcs.ListConfigItemsReq, error) {
//	pbStruct, err := GenQueryFilterByIds(configItemIds)
//	if err != nil {
//		return nil, err
//	}
//
//	return &pbcs.ListConfigItemsReq{
//		BizId:  bizId,
//		AppId:  appId,
//		Filter: pbStruct,
//		Page:   ListPage(),
//	}, nil
//}

// GenListContentByIdsReq generate a list_content request by biz_id, app_id and content_ids.
func GenListContentByIdsReq(bizId, appId uint32, contentIds []uint32) (*pbcs.ListContentsReq, error) {
	pbStruct, err := GenQueryFilterByIds(contentIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListContentsReq{
		BizId:  bizId,
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListCommitByIdsReq generate a list_commit request by biz_id, app_id and commit_ids.
func GenListCommitByIdsReq(bizId, appId uint32, commitIds []uint32) (*pbcs.ListCommitsReq, error) {
	pbStruct, err := GenQueryFilterByIds(commitIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListCommitsReq{
		BizId:  bizId,
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListReleaseByIdsReq generate a list_release request by biz_id, app_id and release_ids.
func GenListReleaseByIdsReq(bizId, appId uint32, releaseIds []uint32) (*pbcs.ListReleasesReq, error) {
	pbStruct, err := GenQueryFilterByIds(releaseIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListReleasesReq{
		BizId:  bizId,
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListStrategySetByIdsReq generate a list_strategy_set request by biz_id, app_id and strategy_set ids.
func GenListStrategySetByIdsReq(bizId, appId uint32, stgSetIds []uint32) (*pbcs.ListStrategySetsReq, error) {
	pbStruct, err := GenQueryFilterByIds(stgSetIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListStrategySetsReq{
		BizId:  bizId,
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListStrategyByIdsReq generate a list_strategy request by biz_id, app_id and strategy ids.
func GenListStrategyByIdsReq(bizId, appId uint32, strategyIds []uint32) (*pbcs.ListStrategiesReq, error) {
	pbStruct, err := GenQueryFilterByIds(strategyIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListStrategiesReq{
		BizId:  bizId,
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListStrategyPublishByIdsReq generate a list_strategy_publish request by biz_id, app_id and publish ids.
func GenListStrategyPublishByIdsReq(bizId, appId uint32, publishIds []uint32) (
	*pbcs.ListPubStrategyHistoriesReq, error) {
	pbStruct, err := GenQueryFilterByIds(publishIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListPubStrategyHistoriesReq{
		BizId:  bizId,
		AppId:  appId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenListInstancePublishByIdsReq generate a list_app_instance_publish request by biz_id and publish ids.
func GenListInstancePublishByIdsReq(bizId uint32, publishIds []uint32) (
	*pbcs.ListPublishedInstanceReq, error) {
	pbStruct, err := GenQueryFilterByIds(publishIds)
	if err != nil {
		return nil, err
	}

	return &pbcs.ListPublishedInstanceReq{
		BizId:  bizId,
		Filter: pbStruct,
		Page:   ListPage(),
	}, nil
}

// GenQueryFilterByIds query app filter by id
func GenQueryFilterByIds(ids []uint32) (*pbstruct.Struct, error) {
	ft := filter.Expression{
		Op: filter.Or,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: ids,
			},
		},
	}
	marshal, err := json.Marshal(ft)
	if err != nil {
		return nil, err
	}

	pbStruct := new(pbstruct.Struct)
	if err := pbStruct.UnmarshalJSON(marshal); err != nil {
		return nil, err
	}
	return pbStruct, nil
}

// ListPage list data page.
func ListPage() *pbbase.BasePage {
	return &pbbase.BasePage{
		Count: false,
		Start: 0,
		Limit: 200,
	}
}

// CountPage count data page.
func CountPage() *pbbase.BasePage {
	return &pbbase.BasePage{
		Count: true,
	}
}

// GenSubSelector generate a sub selector for test
func GenSubSelector() *selector.Selector {
	return &selector.Selector{
		MatchAll: false,
		LabelsOr: selector.Label{
			{
				Key:   "city",
				Op:    new(selector.InType),
				Value: []string{"guangzhou", "shenzhen"},
			},
			{
				Key:   "version",
				Op:    new(selector.GreaterThanEqualType),
				Value: 12,
			},
		},
	}
}

// GenMainSelector generate a main selector for test
func GenMainSelector() *selector.Selector {
	return &selector.Selector{
		MatchAll: false,
		LabelsOr: selector.Label{
			{
				Key:   "area",
				Op:    new(selector.InType),
				Value: []string{"south of china", "eastern of china"},
			},
			{
				Key:   "os",
				Op:    new(selector.EqualType),
				Value: "windows",
			},
		},
	}
}

// GenApiCtxHeader generate request context for api client
func GenApiCtxHeader() (context.Context, http.Header) {
	header := http.Header{}
	header.Set(constant.UserKey, constant.BKUserForTestPrefix+"suite")
	header.Set(constant.RidKey, uuid.UUID())
	header.Set(constant.AppCodeKey, "test")
	header.Add("Cookie", "bk_token="+constant.BKTokenForTest)
	return context.Background(), header
}

// GenCacheContext generate request context for cache client
func GenCacheContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, constant.UserKey, constant.BKUserForTestPrefix+"suite")
	ctx = context.WithValue(ctx, constant.RidKey, uuid.UUID())
	ctx = context.WithValue(ctx, constant.AppCodeKey, "test")
	return ctx
}
