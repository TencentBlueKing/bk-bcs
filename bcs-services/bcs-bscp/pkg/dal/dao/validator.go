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
	"errors"
	"fmt"

	"bscp.io/pkg/dal/types"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/tools"
)

// Validator supplies all the validate operations.
type Validator interface {
	// ValidateTemplateSpacesExist validate if templates spaces exists
	ValidateTemplateSpacesExist(kit *kit.Kit, templateSpaceIDs []uint32) error
	// ValidateTemplatesExist validate if templates exists
	ValidateTemplatesExist(kit *kit.Kit, templateIDs []uint32) error
	// ValidateTemplateRevisionsExist validate if template releases exists
	ValidateTemplateRevisionsExist(kit *kit.Kit, templateRevisionIDs []uint32) error
	// ValidateTemplateRevisionsExistWithTx validate if template releases exists with transaction
	ValidateTemplateRevisionsExistWithTx(kit *kit.Kit, tx *gen.QueryTx, templateRevisionIDs []uint32) error
	// ValidateTemplateSetsExist validate if template sets exists
	ValidateTemplateSetsExist(kit *kit.Kit, templateSetIDs []uint32) error
	// ValidateTemplateExist validate if one template exists
	ValidateTemplateExist(kit *kit.Kit, id uint32) error
	// ValidateTemplateRevisionExist validate if one template release exists
	ValidateTemplateRevisionExist(kit *kit.Kit, id uint32) error
	// ValidateTemplateSetExist validate if one template set exists
	ValidateTemplateSetExist(kit *kit.Kit, id uint32) error
	// ValidateTemplateSpaceNoSubResource validate if one template space has not subresource
	ValidateTemplateSpaceNoSubResource(kit *kit.Kit, id uint32) error
	// ValidateTemplatesBelongToTemplateSet validate if templates belong to a template set
	ValidateTemplatesBelongToTemplateSet(kit *kit.Kit, templateIDs []uint32, templateSetID uint32) error
}

var _ Validator = new(validatorDao)

type validatorDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ValidateTemplateSpacesExist validate if templates spaces exists
func (dao *validatorDao) ValidateTemplateSpacesExist(kit *kit.Kit, templateSpaceIDs []uint32) error {
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateSpaceIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate templates exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateSpaceIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template space id in %v is not exist", diffIDs)
	}

	return nil
}

// ValidateTemplatesExist validate if templates exists
func (dao *validatorDao) ValidateTemplatesExist(kit *kit.Kit, templateIDs []uint32) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate templates exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template id in %v is not exist", diffIDs)
	}

	return nil
}

// ValidateTemplateRevisionsExist validate if template releases exists
func (dao *validatorDao) ValidateTemplateRevisionsExist(kit *kit.Kit, templateRevisionIDs []uint32) error {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateRevisionIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate template releases exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateRevisionIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template release id in %v is not exist", diffIDs)
	}

	return nil
}

// ValidateTemplateRevisionsExistWithTx validate if template releases exists with transaction
func (dao *validatorDao) ValidateTemplateRevisionsExistWithTx(kit *kit.Kit, tx *gen.QueryTx,
	templateRevisionIDs []uint32) error {
	m := tx.TemplateRevision
	q := tx.TemplateRevision.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateRevisionIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate template releases exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateRevisionIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template release id in %v is not exist", diffIDs)
	}

	return nil
}

// ValidateTemplateSetsExist validate if template sets exists
func (dao *validatorDao) ValidateTemplateSetsExist(kit *kit.Kit, templateSetIDs []uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateSetIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate template sets exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateSetIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template set id in %v is not exist", diffIDs)
	}

	return nil
}

// ValidateTemplateExist validate if one template exists
func (dao *validatorDao) ValidateTemplateExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(id)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template %d is not exist", id)
		}
		return fmt.Errorf("get template failed, err: %v", err)
	}

	return nil
}

// ValidateTemplateRevisionExist validate if one template release exists
func (dao *validatorDao) ValidateTemplateRevisionExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(id)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template release %d is not exist", id)
		}
		return fmt.Errorf("get template release failed, err: %v", err)
	}

	return nil
}

// ValidateTemplateSetExist validate if one template set exists
func (dao *validatorDao) ValidateTemplateSetExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(id)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template set %d is not exist", id)
		}
		return fmt.Errorf("get template set failed, err: %v", err)
	}

	return nil
}

// ValidateTemplateSpaceNoSubResource validate if one template space has not subresource
func (dao *validatorDao) ValidateTemplateSpaceNoSubResource(kit *kit.Kit, id uint32) error {
	var (
		tmplSetCnt, tmplCnt int64
		err                 error
	)

	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	if tmplSetCnt, err = q.Where(m.TemplateSpaceID.Eq(id)).Count(); err != nil {
		return fmt.Errorf("get template set count failed, err: %v", err)
	}
	// when the tmplSetCnt is 1, the template set must be default template set
	// in this scenario, allow the template space to be deleted
	if tmplSetCnt > 1 {
		return fmt.Errorf("there are template sets under the template space, need to delete them first")
	}

	tm := dao.genQ.Template
	tq := dao.genQ.Template.WithContext(kit.Ctx)
	if tmplCnt, err = tq.Where(tm.TemplateSpaceID.Eq(id)).Count(); err != nil {
		return fmt.Errorf("get template count failed, err: %v", err)
	}
	if tmplCnt > 0 {
		return fmt.Errorf("there are templates under the template space, need to delete them first")
	}

	return nil
}

// ValidateTemplatesBelongToTemplateSet validate if templates belong to a template set
func (dao *validatorDao) ValidateTemplatesBelongToTemplateSet(
	kit *kit.Kit, templateIDs []uint32, templateSetID uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	type existT struct {
		TemplateIDs types.Uint32Slice `json:"template_ids"`
	}
	var existIDs existT
	if err := q.Select(m.TemplateIDs).Where(m.ID.Eq(templateSetID)).Scan(&existIDs); err != nil {
		return fmt.Errorf("validate templates in a template set failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateIDs, existIDs.TemplateIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template id in %v is not belong to template set id %d", diffIDs, templateSetID)
	}

	return nil
}
