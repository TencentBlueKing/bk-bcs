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

// Package clusterManger cluster-service
package clusterManger

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/header"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

/*
	云凭证导入
*/

// CreateCloudAccount 创建云凭证
func (h *handler) CreateCloudAccount(ctx context.Context, req *CreateCloudAccountRequest,
) (*pb.CreateCloudAccountResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.AccountName) == 0 {
		return nil, errors.Errorf("accountName connot be empty.")
	}
	if req.Account == nil {
		return nil, errors.Errorf("account connot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("projectID connot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换(提示：1、切换方法)
	body, err := h.createCloudAccountReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Post(ctx, req.RequestID, h.createCloudAccountApi(req), body)
	if err != nil {
		return nil, errors.Wrapf(err, "create cloud account failed, traceId: %s, body: %s", req.RequestID,
			cast.ToString(body))
	}

	result := new(pb.CreateCloudAccountResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// createCloudAccountReqToBytes to pb struct data
func (h *handler) createCloudAccountReqToBytes(req *CreateCloudAccountRequest) ([]byte, error) {
	obj := &pb.CreateCloudAccountRequest{
		CloudID:     req.CloudID,
		AccountName: req.AccountName,
		Desc:        req.Desc,
		Account: &pb.Account{
			SecretID:             req.Account.SecretID,
			SecretKey:            req.Account.SecretKey,
			ServiceAccountSecret: req.Account.ServiceAccountSecret,
			SubscriptionID:       req.Account.SubscriptionID,
			ClientID:             req.Account.ClientID,
			TenantID:             req.Account.TenantID,
			ClientSecret:         req.Account.ClientSecret,
		},
		ProjectID: req.ProjectID,
		Creator:   h.config.Username,
		Enable:    wrapperspb.Bool(true),
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.CreateCloudAccountRequest marhsal failed.")
	}

	return body.Bytes(), nil
}

// createCloudAccountApi post
func (h *handler) createCloudAccountApi(req *CreateCloudAccountRequest) string {
	// ( cloudID )
	return fmt.Sprintf(h.backendApi[createCloudAccountApi], req.CloudID)
}

// DeleteCloudAccount 删除云凭证
func (h *handler) DeleteCloudAccount(ctx context.Context, req *DeleteCloudAccountRequest,
) (*pb.DeleteCloudAccountResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.AccountID) == 0 {
		return nil, errors.Errorf("accountID connot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Delete(ctx, req.RequestID, h.deleteCloudAccountApi(req), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "delete cloud account failed, traceId: %s, req: %s", req.RequestID,
			utils.ObjToJson(req))
	}

	result := new(pb.DeleteCloudAccountResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// deleteCloudAccountApi delete
func (h *handler) deleteCloudAccountApi(req *DeleteCloudAccountRequest) string {
	//  ( cloudID + accountID )
	return fmt.Sprintf(h.backendApi[deleteCloudAccountApi], req.CloudID, req.AccountID)
}

// UpdateCloudAccount 更新云凭证
func (h *handler) UpdateCloudAccount(ctx context.Context, req *UpdateCloudAccountRequest,
) (*pb.UpdateCloudAccountResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.AccountID) == 0 {
		return nil, errors.Errorf("accountID connot be empty.")
	}
	if len(req.AccountName) == 0 {
		return nil, errors.Errorf("accountName connot be empty.")
	}
	if req.Account == nil {
		return nil, errors.Errorf("account connot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("projectID connot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换(提示：1、切换方法)
	body, err := h.updateCloudAccountReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Put(ctx, req.RequestID, h.updateCloudAccountApi(req), body)
	if err != nil {
		return nil, errors.Wrapf(err, "update cloud account failed, traceId: %s, body: %s", req.RequestID,
			cast.ToString(body))
	}

	result := new(pb.UpdateCloudAccountResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateCloudAccountReqToBytes to pb struct data
func (h *handler) updateCloudAccountReqToBytes(req *UpdateCloudAccountRequest) ([]byte, error) {
	obj := &pb.UpdateCloudAccountRequest{
		CloudID:     req.CloudID,
		AccountID:   req.AccountID,
		AccountName: req.AccountName,
		Desc:        req.Desc,
		Account: &pb.Account{
			SecretID:             req.Account.SecretID,
			SecretKey:            req.Account.SecretKey,
			ServiceAccountSecret: req.Account.ServiceAccountSecret,
			SubscriptionID:       req.Account.SubscriptionID,
			ClientID:             req.Account.ClientID,
			TenantID:             req.Account.TenantID,
			ClientSecret:         req.Account.ClientSecret,
		},
		ProjectID: req.ProjectID,
		Updater:   h.config.Username,
		Enable:    wrapperspb.Bool(req.Enable),
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.UpdateCloudAccountRequest marhsal failed.")
	}

	return body.Bytes(), nil
}

// updateCloudAccountApi put
func (h *handler) updateCloudAccountApi(req *UpdateCloudAccountRequest) string {
	//put ( cloudID + accountID )
	return fmt.Sprintf(h.backendApi[updateCloudAccountApi], req.CloudID, req.AccountID)
}

// ListCloudAccount 查询云凭证
func (h *handler) ListCloudAccount(ctx context.Context, req *ListCloudAccountRequest,
) (*pb.ListCloudAccountResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换(提示：1、切换方法)
	body, err := h.listCloudAccountReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, req.RequestID, h.listCloudAccountApi(req), body)
	if err != nil {
		return nil, errors.Wrapf(err, "list cloud account failed, traceId: %s, body: %s", req.RequestID,
			utils.ObjToJson(body))
	}

	result := new(pb.ListCloudAccountResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudAccountReqToBytes to pb struct data
func (h *handler) listCloudAccountReqToBytes(req *ListCloudAccountRequest) ([]byte, error) {
	obj := &pb.ListCloudAccountRequest{
		Operator:  req.Operator,
		ProjectID: req.ProjectID,
		AccountID: req.AccountID,
		CloudID:   req.CloudID,
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudAccountRequest marhsal failed.")
	}

	return body.Bytes(), nil
}

// listCloudAccountApi  get
func (h *handler) listCloudAccountApi(req *ListCloudAccountRequest) string {
	// ( cloudID )
	base := fmt.Sprintf(h.backendApi[listCloudAccountApi], req.CloudID)
	return fmt.Sprintf("%s?projectID=%s&operator=%s&accountID=%s", base, req.ProjectID, req.Operator, req.AccountID)
}
