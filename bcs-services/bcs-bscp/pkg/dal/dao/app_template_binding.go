/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// AppTemplateBinding supplies all the app template binding related operations.
type AppTemplateBinding interface {
	// Create one app template binding instance.
	Create(kit *kit.Kit, templateSpace *table.AppTemplateBinding) (uint32, error)
	// Update one app template binding's info.
	Update(kit *kit.Kit, templateSpace *table.AppTemplateBinding) error
	// List app template bindings with options.
	List(kit *kit.Kit, bizID, appID uint32, opt *types.BasePage) ([]*table.AppTemplateBinding, int64, error)
	// Delete one app template binding instance.
	Delete(kit *kit.Kit, templateSpace *table.AppTemplateBinding) error
}

var _ AppTemplateBinding = new(appTemplateBindingDao)

type appTemplateBindingDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one app template binding instance.
func (dao *appTemplateBindingDao) Create(kit *kit.Kit, g *table.AppTemplateBinding) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	templateReleases, err := dao.fillModel(kit, g)
	if err != nil {
		return 0, err
	}

	if err = dao.validateUpsert(kit, g, templateReleases); err != nil {
		return 0, err
	}

	// generate a app template binding id and update to app template binding.v
	var id uint32
	id, err = dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.AppTemplateBinding.WithContext(kit.Ctx)
		if err := q.Create(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err = dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Update one app template binding instance.
func (dao *appTemplateBindingDao) Update(kit *kit.Kit, g *table.AppTemplateBinding) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	templateReleases, err := dao.fillModel(kit, g)
	if err != nil {
		return err
	}

	if err = dao.validateUpsert(kit, g, templateReleases); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var oldOne *table.AppTemplateBinding
	oldOne, err = q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.AppTemplateBinding.WithContext(kit.Ctx)
		if _, err = q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Updates(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err = dao.genQ.Transaction(updateTx); err != nil {
		return err
	}

	return nil
}

// List app template bindings with options.
func (dao *appTemplateBindingDao) List(kit *kit.Kit, bizID, appID uint32,
	opt *types.BasePage) ([]*table.AppTemplateBinding, int64, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one app template binding instance.
func (dao *appTemplateBindingDao) Delete(kit *kit.Kit, g *table.AppTemplateBinding) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.AppTemplateBinding.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(deleteTx); err != nil {
		return err
	}

	return nil
}

// listTemplateReleasesByIDs list template releases details by template release ids.
func (dao *appTemplateBindingDao) listTemplateReleasesByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateRelease, error) {
	m := dao.genQ.TemplateRelease
	q := dao.genQ.TemplateRelease.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// fillModel fill model AppTemplateBinding's fields
// including TemplateSetIDs,TemplateReleaseIDs,TemplateSpaceIDs,TemplateIDs
func (dao *appTemplateBindingDao) fillModel(kit *kit.Kit, g *table.AppTemplateBinding) (
	[]*table.TemplateRelease, error) {
	templateSetIDs, templateReleaseIDs := parseBindings(g.Spec.Bindings)
	g.Spec.TemplateSetIDs = templateSetIDs
	g.Spec.TemplateReleaseIDs = templateReleaseIDs

	templateReleases, err := dao.listTemplateReleasesByIDs(kit, templateReleaseIDs)
	if err != nil {
		return nil, err
	}

	templateSpaceIDs := make(map[uint32]struct{})
	templateIDs := make(map[uint32]struct{})
	for _, tr := range templateReleases {
		templateSpaceIDs[tr.Attachment.TemplateSpaceID] = struct{}{}
		templateIDs[tr.Attachment.TemplateID] = struct{}{}
	}
	g.Spec.TemplateSpaceIDs = convertToSlice(templateSpaceIDs)
	g.Spec.TemplateIDs = convertToSlice(templateIDs)

	return templateReleases, nil
}

func convertToSlice(m map[uint32]struct{}) []uint32 {
	var keys []uint32
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// validateUpsert validate for create or update operation
func (dao *appTemplateBindingDao) validateUpsert(kit *kit.Kit, g *table.AppTemplateBinding,
	templateReleases []*table.TemplateRelease) error {
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}

	if err := dao.validateTemplateSetsExist(kit, g.Spec.TemplateSetIDs); err != nil {
		return err
	}

	if err := dao.validateTemplateReleasesExist(kit, g.Spec.TemplateReleaseIDs); err != nil {
		return err
	}

	if err := validateUniqueKeyOfInput(templateReleases); err != nil {
		return err
	}

	if err := dao.validateUniqueKey(kit, g.Attachment.BizID, g.Attachment.AppID, templateReleases); err != nil {
		return err
	}

	return nil
}

// validateUniqueKeyOfInput validates unique key which is name+path of input only
func validateUniqueKeyOfInput(templateReleases []*table.TemplateRelease) error {
	var uids []uid
	for _, tr := range templateReleases {
		uids = append(uids, uid{
			Name: tr.Spec.Name,
			Path: tr.Spec.Path,
		})
	}
	repeated := findRepeatedElements(uids)
	if len(repeated) > 0 {
		js, _ := json.Marshal(repeated)
		return fmt.Errorf("template's name and path must be unique, these are repeated: %s", js)
	}

	return nil
}

type uid struct {
	Name string
	Path string
}

func findRepeatedElements(slice []uid) []uid {
	frequencyMap := make(map[uid]int)
	var repeatedElements []uid

	// Count the frequency of each uID in the slice
	for _, key := range slice {
		frequencyMap[key]++
	}

	// Check if any uID appears more than once
	for key, count := range frequencyMap {
		if count > 1 {
			repeatedElements = append(repeatedElements, key)
		}
	}

	return repeatedElements
}

// validateUniqueKey validates unique key which is name+path
func (dao *appTemplateBindingDao) validateUniqueKey(kit *kit.Kit, bizID, appID uint32,
	templateReleases []*table.TemplateRelease) error {
	// Note: implement by comparing name+path of input with existing in table config_items and template_releases
	return nil
}

// validateAttachmentExist validate if attachment resource exists before operating template
func (dao *appTemplateBindingDao) validateAttachmentExist(kit *kit.Kit, am *table.AppTemplateBindingAttachment) error {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(am.AppID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template attached app %d is not exist", am.AppID)
		}
		return fmt.Errorf("get template attached app failed, err: %v", err)
	}

	return nil
}

// validateTemplateSetsExist validate if all app template bindings resource exists before operating app template binding
func (dao *appTemplateBindingDao) validateTemplateSetsExist(kit *kit.Kit, templateSetIDs []uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateSetIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate app template bindings exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateSetIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("app template binding id in %v is not exist", diffIDs)
	}

	return nil
}

// validateTemplateReleasesExist validate if all template releases resource exists before operating app template binding
func (dao *appTemplateBindingDao) validateTemplateReleasesExist(kit *kit.Kit, templateReleaseIDs []uint32) error {
	m := dao.genQ.TemplateRelease
	q := dao.genQ.TemplateRelease.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateReleaseIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate template releases exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateReleaseIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template release id in %v is not exist", diffIDs)
	}

	return nil
}

func parseBindings(bindings []*table.TemplateBinding) (templateSetIDs, templateReleasedIDs []uint32) {
	for _, b := range bindings {
		templateSetIDs = append(templateSetIDs, b.TemplateSetID)
		templateReleasedIDs = append(templateReleasedIDs, b.TemplateReleaseIDs...)
	}

	return templateSetIDs, templateReleasedIDs
}
