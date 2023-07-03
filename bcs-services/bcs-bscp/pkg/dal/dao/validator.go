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

	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/tools"
)

// Validator supplies all the validate operations.
type Validator interface {
	// ValidateTemplatesExist validate if templates exists
	ValidateTemplatesExist(kit *kit.Kit, templateIDs []uint32) error
	// ValidateTemplateReleasesExist validate if template releases exists
	ValidateTemplateReleasesExist(kit *kit.Kit, templateReleaseIDs []uint32) error
	// ValidateTemplateSetsExist validate if template sets exists
	ValidateTemplateSetsExist(kit *kit.Kit, templateSetIDs []uint32) error
	// ValidateTemplateExist validate if one template exists
	ValidateTemplateExist(kit *kit.Kit, id uint32) error
	// ValidateTemplateReleaseExist validate if one template release exists
	ValidateTemplateReleaseExist(kit *kit.Kit, id uint32) error
	// ValidateTemplateSetExist validate if one template set exists
	ValidateTemplateSetExist(kit *kit.Kit, id uint32) error
}

var _ Validator = new(validatorDao)

type validatorDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
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

// ValidateTemplateReleasesExist validate if template releases exists
func (dao *validatorDao) ValidateTemplateReleasesExist(kit *kit.Kit, templateReleaseIDs []uint32) error {
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

// ValidateTemplateReleaseExist validate if one template release exists
func (dao *validatorDao) ValidateTemplateReleaseExist(kit *kit.Kit, id uint32) error {
	m := dao.genQ.TemplateRelease
	q := dao.genQ.TemplateRelease.WithContext(kit.Ctx)

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
