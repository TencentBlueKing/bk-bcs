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

package service

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
	pbrkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-kv"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateKv is used to create key-value data.
func (s *Service) CreateKv(ctx context.Context, req *pbcs.CreateKvReq) (*pbcs.CreateKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if err := verifySecretVaule(grpcKit, req.SecretType, req.Value); err != nil {
		return nil, err
	}

	r := &pbds.CreateKvReq{
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
		Spec: &pbkv.KvSpec{
			Key:          req.Key,
			Memo:         req.Memo,
			KvType:       req.KvType,
			Value:        req.Value,
			SecretType:   req.SecretType,
			SecretHidden: req.SecretHidden,
		},
	}
	rp, err := s.client.DS.CreateKv(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateKvResp{
		Id: rp.Id,
	}
	return resp, nil
}

// UpdateKv is used to update key-value data.
func (s *Service) UpdateKv(ctx context.Context, req *pbcs.UpdateKvReq) (*pbcs.UpdateKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if err := verifySecretVaule(grpcKit, req.SecretType, req.Value); err != nil {
		return nil, err
	}

	r := &pbds.UpdateKvReq{
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
		Spec: &pbkv.KvSpec{
			Key:          req.Key,
			Value:        req.Value,
			Memo:         req.Memo,
			SecretHidden: req.SecretHidden,
		},
	}
	if _, err := s.client.DS.UpdateKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateKvResp{}, nil

}

// ListKvs is used to list key-value data.
func (s *Service) ListKvs(ctx context.Context, req *pbcs.ListKvsReq) (*pbcs.ListKvsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListKvsReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		Key:        req.Key,
		Start:      req.Start,
		Limit:      req.Limit,
		All:        req.All,
		SearchKey:  req.SearchKey,
		WithStatus: req.WithStatus,
		KvType:     req.KvType,
		Sort:       req.Sort,
		Order:      req.Order,
		TopIds:     req.TopIds,
		Status:     req.Status,
	}
	if !req.All {
		if req.Limit == 0 {
			return nil, errors.New("limit has to be greater than 0")
		}
		r.Start = req.Start
		r.Limit = req.Limit
	}

	rp, err := s.client.DS.ListKvs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 敏感信息类型需要判断是否隐藏密码
	for _, v := range rp.GetDetails() {
		if v.Spec.KvType != string(table.KvSecret) {
			continue
		}
		if v.Spec.SecretHidden {
			v.Spec.Value = i18n.T(grpcKit, "sensitive data is not visible, unable to view actual content")
		}
	}

	resp := &pbcs.ListKvsResp{
		Count:          rp.Count,
		Details:        rp.Details,
		ExclusionCount: rp.GetExclusionCount(),
	}
	return resp, nil

}

// DeleteKv is used to delete key-value data.
func (s *Service) DeleteKv(ctx context.Context, req *pbcs.DeleteKvReq) (*pbcs.DeleteKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteKvReq{
		Id: req.Id,
		Attachment: &pbkv.KvAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	if _, err := s.client.DS.DeleteKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteKvResp{}, nil

}

// BatchDeleteKv is used to batch delete key-value data.
func (s *Service) BatchDeleteKv(ctx context.Context, req *pbcs.BatchDeleteAppResourcesReq) (
	*pbcs.BatchDeleteResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	var ids []uint32
	ids = req.GetIds()
	if req.ExclusionOperation {
		result, err := s.client.DS.KvFetchIDsExcluding(grpcKit.RpcCtx(), &pbds.KvFetchIDsExcludingReq{
			BizId: req.BizId,
			AppId: req.AppId,
			Ids:   req.GetIds(),
		})
		if err != nil {
			return nil, err
		}
		ids = result.GetIds()
	}

	if len(ids) == 0 {
		return nil, errf.Errorf(errf.InvalidArgument, i18n.T(grpcKit, "id is required"))
	}

	eg, egCtx := errgroup.WithContext(grpcKit.RpcCtx())
	eg.SetLimit(10)

	var successfulIDs, failedIDs []uint32
	var mux sync.Mutex

	// 使用 data-service 原子接口
	for _, v := range ids {
		v := v
		eg.Go(func() error {
			r := &pbds.DeleteKvReq{
				Id: v,
				Attachment: &pbkv.KvAttachment{
					BizId: req.BizId,
					AppId: req.AppId,
				},
			}
			if _, err := s.client.DS.DeleteKv(egCtx, r); err != nil {
				logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)

				// 错误不返回异常，记录错误ID
				mux.Lock()
				failedIDs = append(failedIDs, v)
				mux.Unlock()
				return nil
			}

			mux.Lock()
			successfulIDs = append(successfulIDs, v)
			mux.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		logs.Errorf("batch delete failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete failed"))
	}

	// 全部失败, 当前API视为失败
	if len(failedIDs) == len(ids) {
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete failed"))
	}

	return &pbcs.BatchDeleteResp{SuccessfulIds: successfulIDs, FailedIds: failedIDs}, nil
}

// BatchUpsertKvs is used to insert or update key-value data in bulk.
func (s *Service) BatchUpsertKvs(ctx context.Context, req *pbcs.BatchUpsertKvsReq) (*pbcs.BatchUpsertKvsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	kvs := make([]*pbds.BatchUpsertKvsReq_Kv, 0, len(req.Kvs))
	for _, kv := range req.Kvs {
		kvs = append(kvs, &pbds.BatchUpsertKvsReq_Kv{
			KvAttachment: &pbkv.KvAttachment{
				BizId: req.BizId,
				AppId: req.AppId,
			},
			KvSpec: &pbkv.KvSpec{
				Key:          kv.Key,
				KvType:       kv.KvType,
				Value:        kv.Value,
				Memo:         kv.Memo,
				SecretType:   kv.SecretType,
				SecretHidden: kv.SecretHidden,
			},
		})
	}

	r := &pbds.BatchUpsertKvsReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		Kvs:        kvs,
		ReplaceAll: req.GetReplaceAll(),
	}
	data, err := s.client.DS.BatchUpsertKvs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("batch upsert kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.BatchUpsertKvsResp{Ids: data.Ids}, nil
}

// UnDeleteKv reverses the deletion of a key-value pair by reverting the current kvType and value to the previous
// version.
func (s *Service) UnDeleteKv(ctx context.Context, req *pbcs.UnDeleteKvReq) (*pbcs.UnDeleteKvResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UnDeleteKvReq{
		Key:   req.GetKey(),
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
	}
	if _, err := s.client.DS.UnDeleteKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UnDeleteKvResp{}, nil
}

// UndoKv Undo edited data and return to the latest published version
func (s *Service) UndoKv(ctx context.Context, req *pbcs.UndoKvReq) (*pbcs.UndoKvResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UndoKvReq{
		Key:   req.GetKey(),
		BizId: req.GetBizId(),
		AppId: req.GetAppId(),
	}
	if _, err := s.client.DS.UndoKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("undo kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UndoKvResp{}, nil
}

// CompareKvConflicts compare kv version conflicts
func (s *Service) CompareKvConflicts(ctx context.Context, req *pbcs.CompareKvConflictsReq) (
	*pbcs.CompareKvConflictsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// 两者服务不等，判断服务类型是否支持
	if req.OtherAppId != req.AppId {
		app1, err := s.client.DS.GetAppByID(grpcKit.RpcCtx(), &pbds.GetAppByIDReq{
			AppId: req.AppId,
		})
		if err != nil {
			return nil, err
		}
		app2, err := s.client.DS.GetAppByID(grpcKit.RpcCtx(), &pbds.GetAppByIDReq{
			AppId: req.OtherAppId,
		})
		if err != nil {
			return nil, err
		}

		if app1.Spec.ConfigType != string(table.KV) || app2.Spec.ConfigType != string(table.KV) {
			return nil, errors.New("not a key-value type service")
		}

		if app1.Spec.DataType != string(table.KvAny) && app1.Spec.DataType != app2.Spec.DataType {
			return nil, errors.New("the two service types do not match")
		}
	}

	// 获取该服务未发布的kv
	kv, err := s.client.DS.ListKvs(grpcKit.RpcCtx(), &pbds.ListKvsReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		All:        true,
		WithStatus: true,
		Status:     []string{constant.FileStateAdd, constant.FileStateRevise, constant.FileStateUnchange},
	})
	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 获取已发布的版本
	rkv, err := s.client.DS.ListReleasedKvs(grpcKit.RpcCtx(), &pbds.ListReleasedKvReq{
		BizId:     req.BizId,
		AppId:     req.OtherAppId,
		ReleaseId: req.ReleaseId,
		All:       true,
	})
	if err != nil {
		logs.Errorf("list released kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	conflicts := make(map[string]bool)
	for _, v := range kv.GetDetails() {
		conflicts[v.Spec.Key] = true
	}

	newKv := func(v *pbrkv.ReleasedKv) *pbcs.CompareKvConflictsResp_Kv {
		return &pbcs.CompareKvConflictsResp_Kv{
			Key:          v.Spec.Key,
			KvType:       v.Spec.KvType,
			SecretType:   v.Spec.SecretType,
			SecretHidden: v.Spec.SecretHidden,
			Value:        v.Spec.Value,
			Memo:         v.Spec.Memo,
		}
	}

	exist := make([]*pbcs.CompareKvConflictsResp_Kv, 0)
	nonExist := make([]*pbcs.CompareKvConflictsResp_Kv, 0)
	for _, v := range rkv.GetDetails() {
		if conflicts[v.Spec.Key] {
			exist = append(exist, newKv(v))
		} else {
			nonExist = append(nonExist, newKv(v))
		}
	}

	return &pbcs.CompareKvConflictsResp{Exist: exist, NonExist: nonExist}, nil
}

// ImportKvs 批量导入kv
func (s *Service) ImportKvs(ctx context.Context, req *pbcs.ImportKvsReq) (*pbcs.ImportKvsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// format：text、json、yaml
	var kvMap map[string]interface{}
	switch req.Format {
	case "json":
		if !json.Valid([]byte(req.GetData())) {
			return nil, errors.New(i18n.T(grpcKit, "not legal JSON data"))
		}
		if err := json.Unmarshal([]byte(req.GetData()), &kvMap); err != nil {
			return nil, errors.New(i18n.T(grpcKit, "json format error, err: %v", err))
		}
	case "yaml":
		if err := yaml.Unmarshal([]byte(req.GetData()), &kvMap); err != nil {
			return nil, errors.New(i18n.T(grpcKit, "yaml format error, err: %v", err))
		}
	default:
		return nil, errors.New(i18n.T(grpcKit, "%s type not supported", req.Format))
	}

	kvs, err := handleKv(grpcKit, kvMap)
	if err != nil {
		return nil, err
	}

	resp, err := s.BatchUpsertKvs(grpcKit.RpcCtx(), &pbcs.BatchUpsertKvsReq{
		BizId:      req.GetBizId(),
		AppId:      req.GetAppId(),
		Kvs:        kvs,
		ReplaceAll: false,
	})
	if err != nil {
		return nil, err
	}
	return &pbcs.ImportKvsResp{Ids: resp.Ids}, nil
}

func handleKv(kit *kit.Kit, result map[string]interface{}) ([]*pbcs.BatchUpsertKvsReq_Kv, error) {
	kvMap := []*pbcs.BatchUpsertKvsReq_Kv{}
	for key, value := range result {
		var kVType string
		entry, ok := value.(map[string]interface{})
		if !ok {
			// 判断是不是数值类型
			if isNumber(value) {
				kvMap = append(kvMap, &pbcs.BatchUpsertKvsReq_Kv{
					Key:    key,
					Value:  fmt.Sprintf("%v", value),
					KvType: string(table.KvNumber),
				})
			} else {
				kVType = determineType(value.(string))
				kvMap = append(kvMap, &pbcs.BatchUpsertKvsReq_Kv{
					Key:    key,
					Value:  fmt.Sprintf("%v", value),
					KvType: kVType,
				})
			}
		} else {

			kvType, err := checkKvType(kit, key, entry)
			if err != nil {
				return nil, err
			}

			kvValue, err := checkKv(kit, kvType, key, entry)
			if err != nil {
				return nil, err
			}

			secretType, secretHidden, err := checkSecret(kit, kvType, key, entry)
			if err != nil {
				return nil, err
			}

			kvMemo, _ := entry["memo"].(string)

			kvMap = append(kvMap, &pbcs.BatchUpsertKvsReq_Kv{
				Key:          key,
				Value:        kvValue,
				KvType:       kvType,
				Memo:         kvMemo,
				SecretType:   secretType,
				SecretHidden: secretHidden,
			})
		}
	}
	return kvMap, nil
}

func checkKvType(kit *kit.Kit, key string, entry map[string]interface{}) (string, error) {

	kvType, ok := entry["kv_type"].(string)
	if !ok {
		return "", errors.New(i18n.T(kit, "config item %s kv type error", key))
	}

	if err := validateKvType(kvType); err != nil {
		return kvType, fmt.Errorf("config item %s %v", key, err)
	}

	return kvType, nil
}

func checkKv(kit *kit.Kit, kvType, key string, entry map[string]interface{}) (string, error) {

	kvValue, okVal := entry["value"]
	if !okVal {
		return "", errors.New(i18n.T(kit, "format error, please check the key: %s", key))
	}
	val := fmt.Sprintf("%v", kvValue)
	// json 和 yaml 都需要格式化
	if kvType == string(table.KvJson) {

		v, ok := kvValue.(string)
		if !ok {
			return "", errors.New(i18n.T(kit, "config item %s format error", key))
		}

		mv, err := json.Marshal(v)
		if err != nil {
			return "", errors.New(i18n.T(kit, "config item %s json format error", key))
		}

		// 需要处理转义符
		var data interface{}
		err = json.Unmarshal(mv, &data)
		if err != nil {
			return "", errors.New(i18n.T(kit, "config item %s json format error", key))
		}
		val, ok = data.(string)
		if !ok {
			return "", errors.New(i18n.T(kit, "config item %s format error", key))
		}
	} else if kvType == string(table.KvYAML) {
		_, ok := kvValue.(string)
		if !ok {
			ys, err := yaml.Marshal(kvValue)
			if err != nil {
				return "", errors.New(i18n.T(kit, "config item %s yaml format error", key))
			}
			val = string(ys)
		}
	}

	return val, nil
}

func checkSecret(kit *kit.Kit, kvType, key string, entry map[string]interface{}) (string, bool, error) {

	var secretHidden bool

	// 不是密钥类型
	if kvType != string(table.KvSecret) {
		return "", secretHidden, nil
	}

	// 判断是否隐藏
	secretHidden, okSecretHidden := entry["secret_hidden"].(bool)
	if !okSecretHidden {
		return "", secretHidden, errors.New(i18n.T(kit, "config item %s secret hidden error", key))
	}

	secretType, ok := entry["secret_type"].(string)
	if !ok || secretType == "" {
		return secretType, secretHidden, errors.New(i18n.T(kit, "the key type for config item %s cannot be empty", key))
	}

	// 验证密钥类型
	if err := validateSecretType(secretType); err != nil {
		return secretType, secretHidden, errors.New(i18n.T(kit, "config item %s secret type error, err: %v", key, err))
	}

	// 验证密钥值
	kvValue, okVal := entry["value"].(string)
	if !okVal {
		return secretType, secretHidden, errors.New(i18n.T(kit, "config item %s value error", key))
	}
	if err := verifySecretVaule(kit, secretType, kvValue); err != nil {
		return secretType, secretHidden, fmt.Errorf("config item %s %v", key, err)
	}

	return secretType, secretHidden, nil
}

// 验证kv类型
func validateKvType(kvType string) error {
	switch kvType {
	case string(table.KvStr):
	case string(table.KvNumber):
	case string(table.KvText):
	case string(table.KvJson):
	case string(table.KvYAML):
	case string(table.KvXml):
	case string(table.KvSecret):
	default:
		return errors.New("invalid data-type")
	}
	return nil
}

// 验证密钥类型
func validateSecretType(secretType string) error {
	switch secretType {
	case string(table.SecretTypePassword):
	case string(table.SecretTypeCertificate):
	case string(table.SecretTypeSecretKey):
	case string(table.SecretTypeToken):
	case string(table.SecretTypeCustom):
	default:
		return errors.New("invalid secret-type")
	}
	return nil
}

// 根据值判断类型
func determineType(value string) string {
	var result string
	switch {
	case isJSON(value):
		result = string(table.KvJson)
	case isYAML(value):
		result = string(table.KvYAML)
	case isXML(value):
		result = string(table.KvXml)
	case isTEXT(value):
		result = string(table.KvText)
	case isNumber(value):
		result = string(table.KvNumber)
	default:
		result = string(table.KvStr)
	}
	return result
}

// 判断是否为 JSON
func isJSON(value string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(value), &js) == nil
}

// 判断是否为 YAML
func isYAML(value string) bool {
	var yml map[string]interface{}
	return yaml.Unmarshal([]byte(value), &yml) == nil
}

// 判断是否为 XML
func isXML(value string) bool {
	var xmlData interface{}
	return xml.Unmarshal([]byte(value), &xmlData) == nil
}

// 判断是否为 TEXT
func isTEXT(value string) bool {
	return strings.Contains(value, "\n")
}

// 判断是不是 Number
func isNumber(value interface{}) bool {
	// 获取值的类型
	valType := reflect.TypeOf(value)

	// 检查类型是否为数字
	switch valType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// 验证密钥的值
func verifySecretVaule(kit *kit.Kit, secretType, value string) error {
	if value == "敏感信息无法导出" {
		return errors.New(i18n.T(kit, `please set a password`))
	}

	if secretType == string(table.SecretTypeCertificate) && !validateCertificate(value) {
		return errors.New(i18n.T(kit, `the certificate format is incorrect, only X.509 format is supported`))
	}

	return nil
}

// 验证证书
func validateCertificate(certPEM string) bool {
	// 解析PEM编码的证书
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return false
	}

	// 尝试解析X.509证书
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		return false
	}

	return true
}
