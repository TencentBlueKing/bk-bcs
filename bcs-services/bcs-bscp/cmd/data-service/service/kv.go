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
	"errors"
	"fmt"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbkv "bscp.io/pkg/protocol/core/kv"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateKv is used to create key-value data.
func (s *Service) CreateKv(ctx context.Context, req *pbds.CreateKvReq) (*pbds.CreateResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Kv().GetByKey(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key); err == nil {
		return nil, fmt.Errorf("kv same key %s already exists", req.Spec.Key)
	}

	opt := &types.UpsertKvOption{
		BizID:  req.Attachment.BizId,
		AppID:  req.Attachment.AppId,
		Key:    req.Spec.Key,
		Value:  req.Spec.Value,
		KvType: types.KvType(req.Spec.KvType),
	}
	version, err := s.vault.UpsertKv(kt, opt)
	if err != nil {
		logs.Errorf("create kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	kv := &table.Kv{
		Spec:       req.Spec.KvSpec(),
		Attachment: req.Attachment.KvAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	kv.Spec.Version = uint32(version)
	id, err := s.dao.Kv().Create(kt, kv)
	if err != nil {
		logs.Errorf("create kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

// UpdateKv is used to update key-value data.
func (s *Service) UpdateKv(ctx context.Context, req *pbds.UpdateKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kv, err := s.dao.Kv().GetByKey(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key)
	if err != nil {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}

	kvType, _, err := s.getKv(kt, req.Attachment.BizId, req.Attachment.AppId, kv.Spec.Version, kv.Spec.Key)
	if err != nil {
		logs.Errorf("get vault kv (%d) data failed: err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}

	opt := &types.UpsertKvOption{
		BizID:  req.Attachment.BizId,
		AppID:  req.Attachment.AppId,
		Key:    kv.Spec.Key,
		Value:  req.Spec.Value,
		KvType: kvType,
	}
	version, err := s.vault.UpsertKv(kt, opt)
	if err != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	kv.Revision = &table.Revision{
		Reviser:   kt.User,
		UpdatedAt: time.Now().UTC(),
	}

	kv.Spec.Version = uint32(version)
	if e := s.dao.Kv().Update(kt, kv); e != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", e, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil

}

// ListKvs is used to list key-value data.
func (s *Service) ListKvs(ctx context.Context, req *pbds.ListKvsReq) (*pbds.ListKvsResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListKvOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		Key:       req.Key,
		SearchKey: req.SearchKey,
		All:       req.All,
		Page:      page,
	}
	po := &types.PageOption{
		EnableUnlimitedLimit: true,
	}
	if err := opt.Validate(po); err != nil {
		return nil, err
	}
	details, count, err := s.dao.Kv().List(kt, opt)
	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var kvs []*pbkv.Kv

	for _, detail := range details {
		kvType, value, e := s.getKv(kt, req.BizId, req.AppId, detail.Spec.Version, detail.Spec.Key)
		if e != nil {
			logs.Errorf("list kv failed, err: %v, rid: %s", e, kt.Rid)
			return nil, e
		}
		kvs = append(kvs, pbkv.PbKv(detail, kvType, value))
	}

	resp := &pbds.ListKvsResp{
		Count:   uint32(count),
		Details: kvs,
	}
	return resp, nil

}

// DeleteKv is used to delete key-value data.
func (s *Service) DeleteKv(ctx context.Context, req *pbds.DeleteKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kv, err := s.dao.Kv().GetByKey(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key)
	if err != nil {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()
	if e := s.dao.Kv().DeleteWithTx(kt, tx, kv); e != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	opt := &types.DeleteKvOpt{
		BizID: kv.Attachment.BizID,
		AppID: kv.Attachment.AppID,
		Key:   kv.Spec.Key,
	}
	if e := s.vault.DeleteKv(kt, opt); e != nil {
		logs.Errorf("delete vault kv (%d) data failed: err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// BatchUpsertKvs is used to insert or update key-value data in bulk.
func (s *Service) BatchUpsertKvs(ctx context.Context, req *pbds.BatchUpsertKvsReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	var editingKeyArr []string
	for _, kv := range req.Kvs {
		editingKeyArr = append(editingKeyArr, kv.KvSpec.Key)
	}
	editingKv, err := s.dao.Kv().ListAllKvByKey(kt, req.AppId, req.BizId, editingKeyArr)
	if err != nil {
		logs.Errorf("list editing kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	newKvMap := make(map[string]*pbds.BatchUpsertKvsReq_Kv)

	for _, kv := range req.Kvs {
		newKvMap[kv.KvSpec.Key] = kv
	}

	editingKvMap := make(map[string]*table.Kv)
	for _, kv := range editingKv {
		editingKvMap[kv.Spec.Key] = kv
	}

	// 在vault中执行更新
	versionMap, err := s.doBatchUpsertVault(kt, req, editingKvMap)
	if err != nil {
		return nil, err
	}

	toUpdate, toCreate, err := s.checkKvs(kt, req, editingKvMap, versionMap)
	if err != nil {
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	if len(toCreate) > 0 {
		if err = s.dao.Kv().BatchCreateWithTx(kt, tx, toCreate); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	if len(toUpdate) > 0 && req.ReplaceAll {
		if err = s.dao.Kv().BatchUpdateWithTx(kt, tx, toUpdate); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

func (s *Service) getKv(kt *kit.Kit, bizID, appID, version uint32, key string) (types.KvType, string, error) {
	opt := &types.GetKvByVersion{
		BizID:   bizID,
		AppID:   appID,
		Key:     key,
		Version: int(version),
	}

	return s.vault.GetKvByVersion(kt, opt)
}

// doBatchUpsertVault is used to perform bulk insertion or update of key-value data in Vault.
func (s *Service) doBatchUpsertVault(kt *kit.Kit, req *pbds.BatchUpsertKvsReq,
	editingKvMap map[string]*table.Kv) (map[string]int, error) {

	versionMap := make(map[string]int)

	for _, kv := range req.Kvs {

		opt := &types.UpsertKvOption{
			BizID: req.BizId,
			AppID: req.AppId,
			Key:   kv.KvSpec.Key,
			Value: kv.KvSpec.Value,
		}

		if editing, exists := editingKvMap[kv.KvSpec.Key]; exists {
			kvType, _, err := s.getKv(kt, req.BizId, req.AppId, editing.Spec.Version, kv.KvSpec.Key)
			if err != nil {
				return nil, err
			}
			opt.KvType = kvType
		} else {
			opt.KvType = types.KvType(kv.KvSpec.KvType)
		}

		version, err := s.vault.UpsertKv(kt, opt)
		if err != nil {
			return nil, err
		}
		versionMap[kv.KvSpec.Key] = version
	}

	return versionMap, nil

}

func (s *Service) checkKvs(kt *kit.Kit, req *pbds.BatchUpsertKvsReq, editingKvMap map[string]*table.Kv,
	versionMap map[string]int) (toUpdate, toCreate []*table.Kv, err error) {

	for _, kv := range req.Kvs {

		var version int
		var exists bool
		var editing *table.Kv

		if version, exists = versionMap[kv.KvSpec.Key]; !exists {
			return nil, nil, errors.New("save kv fail")
		}

		now := time.Now().UTC()

		if editing, exists = editingKvMap[kv.KvSpec.Key]; exists {
			// 更新
			toUpdate = append(toUpdate, &table.Kv{
				ID: editing.ID,
				Spec: &table.KvSpec{
					Key:     kv.KvSpec.Key,
					Version: uint32(version),
				},
				Attachment: &table.KvAttachment{
					BizID: req.BizId,
					AppID: req.AppId,
				},
				Revision: editing.Revision,
			})
		} else {
			// 创建
			toCreate = append(toCreate, &table.Kv{
				Spec: &table.KvSpec{
					Key:     kv.KvSpec.Key,
					Version: uint32(version),
				},
				Attachment: &table.KvAttachment{
					BizID: req.BizId,
					AppID: req.AppId,
				},
				Revision: &table.Revision{
					Creator:   kt.User,
					Reviser:   kt.User,
					CreatedAt: now,
					UpdatedAt: now,
				},
			})

		}

	}

	return toUpdate, toCreate, nil
}
