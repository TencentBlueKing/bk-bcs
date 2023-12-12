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

	"gorm.io/gorm"

	"bscp.io/pkg/criteria/constant"
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

	kv2, err := s.dao.Kv().GetByKvState(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key,
		[]string{string(table.KvStateAdd), string(table.KvStateUnchange), string(table.KvStateRevise)})
	if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}
	if !errors.Is(gorm.ErrRecordNotFound, err) {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, fmt.Errorf("kv same key %s already exists", req.Spec.Key)
	}
	fmt.Println(kv2)

	app, err := s.dao.App().Get(kt, req.Attachment.BizId, req.Attachment.AppId)
	if err != nil {
		return nil, fmt.Errorf("get app fail,err : %v", req.Spec.Key)
	}
	if !checkKVTypeMatch(table.DataType(req.Spec.KvType), app.Spec.DataType) {
		return nil, fmt.Errorf("kv type does not match the data type defined in the application")
	}

	opt := &types.UpsertKvOption{
		BizID:  req.Attachment.BizId,
		AppID:  req.Attachment.AppId,
		Key:    req.Spec.Key,
		Value:  req.Spec.Value,
		KvType: table.DataType(req.Spec.KvType),
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
		},
	}
	kv.Spec.Version = uint32(version)
	kv.KvState = table.KvStateAdd
	id, err := s.dao.Kv().Create(kt, kv)
	if err != nil {
		logs.Errorf("create kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

func checkKVTypeMatch(kvType, appKvType table.DataType) bool {
	if appKvType == table.KvAny {
		return true
	}
	return kvType == appKvType
}

// UpdateKv is used to update key-value data.
func (s *Service) UpdateKv(ctx context.Context, req *pbds.UpdateKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kv, err := s.dao.Kv().GetByKvState(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key,
		[]string{string(table.KvStateAdd), string(table.KvStateUnchange), string(table.KvStateRevise)})
	if err != nil {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}

	opt := &types.UpsertKvOption{
		BizID:  req.Attachment.BizId,
		AppID:  req.Attachment.AppId,
		Key:    kv.Spec.Key,
		Value:  req.Spec.Value,
		KvType: kv.Spec.KvType,
	}
	version, err := s.vault.UpsertKv(kt, opt)
	if err != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if kv.KvState != table.KvStateRevise {
		ok, err := s.shouldUpdateKvStatus(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key)
		if err != nil {
			return nil, err
		}
		if ok {
			kv.KvState = table.KvStateRevise
		}
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

// shouldUpdateKvStatus 是否需要修改状态
func (s *Service) shouldUpdateKvStatus(kt *kit.Kit, bizID, appID uint32, key string) (bool, error) {

	_, err := s.dao.ReleasedKv().GetReleasedLatelyByKey(kt, bizID, appID, key)
	if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
		return false, err
	}

	if errors.Is(gorm.ErrRecordNotFound, err) {
		return false, nil
	}

	return true, nil

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
		KvType:    req.KvType,
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
	var kvReleased []*table.ReleasedKv

	kvReleased, err = s.dao.ReleasedKv().GetReleasedLately(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("get released rkv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	kvs, err = s.setKvTypeAndValue(kt, details, kvs, kvReleased)
	if err != nil {
		return nil, err
	}

	resp := &pbds.ListKvsResp{
		Count:   uint32(count),
		Details: kvs,
	}
	return resp, nil

}

func (s *Service) setKvTypeAndValue(kt *kit.Kit, details []*table.Kv, _ []*pbkv.Kv,
	kvRelease []*table.ReleasedKv) ([]*pbkv.Kv, error) {

	kvMap := make(map[string]*table.Kv, len(details))
	for _, detail := range details {
		kvMap[detail.Spec.Key] = detail
	}
	releaseMap := make(map[string]*table.ReleasedKv, len(kvRelease))
	for _, release := range kvRelease {
		releaseMap[release.Spec.Key] = release
	}

	kvs := make([]*pbkv.Kv, 0)

	for _, one := range details {
		if one.KvState == constant.KvStateDelete {
			rkv, ok := releaseMap[one.Spec.Key]
			if !ok {
				return nil, fmt.Errorf("no released key found for key %s, rid: %s", one.Spec.Key, kt.Rid)
			}
			_, kvValue, err := s.getReleasedKv(kt, one.Attachment.BizID, one.Attachment.AppID, rkv.Spec.Version,
				rkv.ReleaseID, rkv.Spec.Key)
			if err != nil {
				return nil, err
			}

			kvs = append(kvs, pbkv.PbKv(one, kvValue))

		} else {
			kv, ok := kvMap[one.Spec.Key]
			if !ok {
				return nil, fmt.Errorf("no key found for key %s, rid: %s", one.Spec.Key, kt.Rid)
			}
			_, kvValue, err := s.getKv(kt, one.Attachment.BizID, one.Attachment.AppID, kv.Spec.Version,
				kv.Spec.Key)
			if err != nil {
				return nil, err
			}

			kvs = append(kvs, pbkv.PbKv(one, kvValue))

		}
	}

	return kvs, nil

}

// DeleteKv is used to delete key-value data.
func (s *Service) DeleteKv(ctx context.Context, req *pbds.DeleteKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kvs, err := s.dao.Kv().ListAllKvByKey(kt, req.Attachment.AppId, req.Attachment.BizId, []string{req.Spec.Key}, nil)
	if err != nil {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
	}

	tx := s.dao.GenQuery().Begin()

	for _, kv := range kvs {

		switch kv.KvState {
		case table.KvStateAdd:
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

		case table.KvStateRevise, table.KvStateUnchange:
			kv.KvState = table.KvStateDelete
			err = s.dao.Kv().UpdateWithTx(kt, tx, kv)
			if err != nil {
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				logs.Errorf("delete vault kv (%d) data failed: err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
				return nil, err
			}

		}

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
	kvStateArr := []string{
		string(table.KvStateUnchange),
		string(table.KvStateAdd),
		string(table.KvStateRevise),
	}
	editingKv, err := s.dao.Kv().ListAllKvByKey(kt, req.AppId, req.BizId, editingKeyArr, kvStateArr)
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

	// 如果 添加、 删除

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

func (s *Service) getKv(kt *kit.Kit, bizID, appID, version uint32, key string) (table.DataType, string, error) {
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
			opt.KvType = table.DataType(kv.KvSpec.KvType)
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
			if editing.KvState == table.KvStateUnchange {
				ok, err := s.shouldUpdateKvStatus(kt, editing.Attachment.BizID, editing.Attachment.AppID, editing.Spec.Key)
				if err != nil {
					return nil, nil, err
				}
				if ok {
					editing.KvState = table.KvStateRevise
				}

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
			}

		} else {
			// 创建
			toCreate = append(toCreate, &table.Kv{
				Spec: &table.KvSpec{
					Key:     kv.KvSpec.Key,
					Version: uint32(version),
					KvType:  table.DataType(kv.KvSpec.KvType),
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

// UnDeleteKv Revert the deletion of the key-value pair by restoring it to the version before the last one.
func (s *Service) UnDeleteKv(ctx context.Context, req *pbds.UnDeleteKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	rkv, err := s.dao.ReleasedKv().GetReleasedLatelyByKey(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key)
	if err != nil {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}
	kvType, kvValue, err := s.getReleasedKv(kt, req.Attachment.BizId, req.Attachment.AppId, rkv.Spec.Version,
		rkv.ReleaseID, req.Spec.Key)
	if err != nil {
		logs.Errorf("get vault kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}

	opt := &types.UpsertKvOption{
		BizID:  req.Attachment.BizId,
		AppID:  req.Attachment.AppId,
		Key:    req.Spec.Key,
		Value:  kvValue,
		KvType: kvType,
	}
	v, err := s.vault.UpsertKv(kt, opt)
	if err != nil {
		logs.Errorf("update vault kv (%d) data failed: err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}

	kv, err := s.dao.Kv().GetByKey(kt, req.Attachment.BizId, req.Attachment.AppId, req.Spec.Key)
	if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
		logs.Errorf("get kv (%d) failed, err: %v, rid: %s", req.Spec.Key, err, kt.Rid)
		return nil, err
	}
	if errors.Is(gorm.ErrRecordNotFound, err) {
		kv = &table.Kv{
			Spec:       req.Spec.KvSpec(),
			Attachment: req.Attachment.KvAttachment(),
			Revision: &table.Revision{
				Creator: kt.User,
			},
		}
		kv.Spec.Version = uint32(v)
		if _, e := s.dao.Kv().Create(kt, kv); e != nil {
			logs.Errorf("update kv failed, err: %v, rid: %s", e, kt.Rid)
			return nil, e
		}
	} else {
		kv.Spec.Version = uint32(v)
		kv.Spec.KvType = rkv.Spec.KvType
		kv.Revision = &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: rkv.Revision.CreatedAt,
		}
		if e := s.dao.Kv().Update(kt, kv); e != nil {
			logs.Errorf("update kv failed, err: %v, rid: %s", e, kt.Rid)
			return nil, e
		}
	}

	return new(pbbase.EmptyResp), nil

}
