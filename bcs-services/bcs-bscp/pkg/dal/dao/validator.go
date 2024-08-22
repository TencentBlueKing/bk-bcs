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

package dao

import (
	"errors"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Validator supplies all the validate operations.
type Validator interface {
	// ValidateTmplSpacesExist validate if templates spaces exists
	ValidateTmplSpacesExist(kit *kit.Kit, templateSpaceIDs []uint32) error
	// ValidateTemplatesExist validate if templates exists
	ValidateTemplatesExist(kit *kit.Kit, templateIDs []uint32) error
	// ValidateTmplRevisionsExist validate if template releases exists
	ValidateTmplRevisionsExist(kit *kit.Kit, templateRevisionIDs []uint32) error
	// ValidateTmplRevisionsExistWithTx validate if template releases exists with transaction
	ValidateTmplRevisionsExistWithTx(kit *kit.Kit, tx *gen.QueryTx, templateRevisionIDs []uint32) error
	// ValidateTmplSetsExist validate if template sets exists
	ValidateTmplSetsExist(kit *kit.Kit, templateSetIDs []uint32) error
	// ValidateTemplateExist validate if one template exists
	ValidateTemplateExist(kit *kit.Kit, id uint32) error
	// ValidateTmplRevisionExist validate if one template release exists
	ValidateTmplRevisionExist(kit *kit.Kit, id uint32) error
	// ValidateTmplSetExist validate if one template set exists
	ValidateTmplSetExist(kit *kit.Kit, id uint32) error
	// ValidateTmplSpaceNoSubRes validate if one template space has not subresource
	ValidateTmplSpaceNoSubRes(kit *kit.Kit, id uint32) error
	// ValidateTmplsBelongToTmplSet validate if templates belong to a template set
	ValidateTmplsBelongToTmplSet(kit *kit.Kit, templateIDs []uint32, templateSetID uint32) error
}

var _ Validator = new(validatorDao)

type validatorDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ValidateTmplSpacesExist validate if templates spaces exists
func (dao *validatorDao) ValidateTmplSpacesExist(kit *kit.Kit, templateSpaceIDs []uint32) error {
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateSpaceIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "validate templates exist failed, err: %v", err))
	}

	diffIDs := tools.SliceDiff(templateSpaceIDs, existIDs)
	if len(diffIDs) > 0 {
		return errf.Errorf(errf.InvalidRequest, i18n.T(kit, "template space id in %v is not exist", diffIDs))
	}

	return nil
}

// ValidateTemplatesExist validate if templates exists
func (dao *validatorDao) ValidateTemplatesExist(kit *kit.Kit, templateIDs []uint32) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "validate templates exist failed, err: %v", err))
	}

	diffIDs := tools.SliceDiff(templateIDs, existIDs)
	if len(diffIDs) > 0 {
		return errf.Errorf(errf.InvalidRequest, i18n.T(kit, "template id in %v is not exist", diffIDs))
	}

	return nil
}

// ValidateTmplRevisionsExist validate if template releases exists
func (dao *validatorDao) ValidateTmplRevisionsExist(kit *kit.Kit, templateRevisionIDs []uint32) error {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateRevisionIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "validate template releases exist failed, err: %v", err))
	}

	diffIDs := tools.SliceDiff(templateRevisionIDs, existIDs)
	if len(diffIDs) > 0 {
		return errf.Errorf(errf.InvalidRequest, i18n.T(kit, "template revision id in %v is not exist", diffIDs))
	}

	return nil
}

// ValidateTmplRevisionsExistWithTx validate if template releases exists with transaction
func (dao *validatorDao) ValidateTmplRevisionsExistWithTx(kit *kit.Kit, tx *gen.QueryTx,
	templateRevisionIDs []uint32) error {
	m := tx.TemplateRevision
	q := tx.TemplateRevision.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateRevisionIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "validate template releases exist failed, err: %v", err))
	}

	diffIDs := tools.SliceDiff(templateRevisionIDs, existIDs)
	if len(diffIDs) > 0 {
		return errf.Errorf(errf.InvalidRequest, i18n.T(kit, "template revision id in %v is not exist", diffIDs))
	}

	return nil
}

// ValidateTmplSetsExist validate if template sets exists
func (dao *validatorDao) ValidateTmplSetsExist(kit *kit.Kit, templateSetIDs []uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateSetIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "validate template sets exist failed, err: %v", err))
	}

	diffIDs := tools.SliceDiff(templateSetIDs, existIDs)
	if len(diffIDs) > 0 {
		return errf.Errorf(errf.InvalidRequest, i18n.T(kit, "template set id in %v is not exist", diffIDs))
	}

	return nil
}

// ValidateTemplateExist validate if one template exists
func (dao *validatorDao) ValidateTemplateExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(id)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "template %d is not exist", id))
		}
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get template failed, err: %v", err))
	}

	return nil
}

// ValidateTmplRevisionExist validate if one template release exists
func (dao *validatorDao) ValidateTmplRevisionExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(id)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "template release %d is not exist", id))
		}
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get template release failed, err: %v", err))
	}

	return nil
}

// ValidateTmplSetExist validate if one template set exists
func (dao *validatorDao) ValidateTmplSetExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(id)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "template set %d is not exist", id))
		}
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get template set failed, err: %v", err))
	}

	return nil
}

// ValidateTmplSpaceNoSubRes validate if one template space has not subresource
func (dao *validatorDao) ValidateTmplSpaceNoSubRes(kit *kit.Kit, id uint32) error {
	var (
		tmplSetCnt, tmplCnt int64
		err                 error
	)

	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	if tmplSetCnt, err = q.Where(m.TemplateSpaceID.Eq(id)).Count(); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get template set count failed, err: %v", err))
	}
	if tmplSetCnt > 0 {
		return errf.Errorf(errf.InvalidRequest,
			i18n.T(kit, "there are template sets under the template space, need to delete them first"))
	}

	tm := dao.genQ.Template
	tq := dao.genQ.Template.WithContext(kit.Ctx)
	if tmplCnt, err = tq.Where(tm.TemplateSpaceID.Eq(id)).Count(); err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get template count failed, err: %v", err))
	}
	if tmplCnt > 0 {
		return errf.Errorf(errf.InvalidRequest,
			i18n.T(kit, "there are templates under the template space, need to delete them first"))
	}

	return nil
}

// ValidateTmplsBelongToTmplSet validate if templates belong to a template set
func (dao *validatorDao) ValidateTmplsBelongToTmplSet(
	kit *kit.Kit, templateIDs []uint32, templateSetID uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	type existT struct {
		TemplateIDs types.Uint32Slice `json:"template_ids"`
	}
	var existIDs existT
	if err := q.Select(m.TemplateIDs).Where(m.ID.Eq(templateSetID)).Scan(&existIDs); err != nil {
		return errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "validate templates in a template set failed, err: %v", err))
	}

	diffIDs := tools.SliceDiff(templateIDs, existIDs.TemplateIDs)
	if len(diffIDs) > 0 {
		return errf.Errorf(errf.InvalidRequest,
			i18n.T(kit, "template id in %v is not belong to template set id %d", diffIDs, templateSetID))
	}

	return nil
}
