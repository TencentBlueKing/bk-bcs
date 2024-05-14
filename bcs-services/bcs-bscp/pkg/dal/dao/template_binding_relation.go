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
	"gorm.io/datatypes"
	rawgen "gorm.io/gen"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// TemplateBindingRelation supplies all the template binding relation query operations.
type TemplateBindingRelation interface {
	// GetTemplateBoundUnnamedAppCount get bound unnamed app count of the target template.
	GetTemplateBoundUnnamedAppCount(kit *kit.Kit, bizID, templateID uint32) (uint32, error)
	// GetTemplateBoundNamedAppCount get bound named app count of the target template.
	GetTemplateBoundNamedAppCount(kit *kit.Kit, bizID, templateID uint32) (uint32, error)
	// GetTemplateBoundTemplateSetCount get bound template set count of the target template.
	GetTemplateBoundTemplateSetCount(kit *kit.Kit, bizID, templateID uint32) (uint32, error)
	// GetTemplateRevisionBoundUnnamedAppCount get bound unnamed app count of the target template release.
	GetTemplateRevisionBoundUnnamedAppCount(kit *kit.Kit, bizID, templateRevisionID uint32) (uint32, error)
	// GetTemplateRevisionBoundNamedAppCount get bound named app count of the target template release.
	GetTemplateRevisionBoundNamedAppCount(kit *kit.Kit, bizID, templateRevisionID uint32) (uint32, error)
	// GetTemplateSetBoundUnnamedAppCount get bound unnamed app count of the target template set.
	GetTemplateSetBoundUnnamedAppCount(kit *kit.Kit, bizID, templateSetID uint32) (uint32, error)
	// GetTemplateSetBoundNamedAppCount get bound named app count of the target template set.
	GetTemplateSetBoundNamedAppCount(kit *kit.Kit, bizID, templateSetID uint32) (uint32, error)
	// ListTmplBoundUnnamedApps list bound unnamed app details of the target template.
	ListTmplBoundUnnamedApps(kit *kit.Kit, bizID, templateID uint32) ([]*types.TmplBoundUnnamedAppDetail, error)
	// ListTmplBoundNamedApps list bound named app details of the target template.
	ListTmplBoundNamedApps(kit *kit.Kit, bizID, templateID uint32) ([]*types.TmplBoundNamedAppDetail, error)
	// ListTmplBoundTmplSets list bound template set details of the target template.
	ListTmplBoundTmplSets(kit *kit.Kit, bizID, templateID uint32) ([]uint32, error)
	// ListTmplRevisionBoundUnnamedApps list bound unnamed app details of the target template release.
	ListTmplRevisionBoundUnnamedApps(kit *kit.Kit, bizID, templateRevisionID uint32) ([]uint32, error)
	// ListTmplRevisionBoundNamedApps list bound named app details of the target template release.
	ListTmplRevisionBoundNamedApps(kit *kit.Kit, bizID, templateRevisionID uint32) (
		[]*types.TmplRevisionBoundNamedAppDetail, error)
	// ListTmplSetBoundUnnamedApps list bound unnamed app details of the target template set.
	ListTmplSetBoundUnnamedApps(kit *kit.Kit, bizID, templateSetID uint32) ([]uint32, error)
	// ListTmplSetBoundNamedApps list bound named app details of the target template set.
	ListTmplSetBoundNamedApps(kit *kit.Kit, bizID, templateSetID uint32) ([]*types.TmplSetBoundNamedAppDetail, error)
	// ListLatestTmplBoundUnnamedApps list bound unnamed app details of the latest target template.
	ListLatestTmplBoundUnnamedApps(kit *kit.Kit, bizID, templateID uint32) ([]*table.AppTemplateBinding, error)
	// ListTemplateSetsBoundATBs list bound app template bindings of the target template sets.
	ListTemplateSetsBoundATBs(kit *kit.Kit, bizID uint32, templateSetIDs []uint32) ([]*table.AppTemplateBinding, error)
	// ListTemplatesBoundATBs list bound app template bindings of the target templates.
	ListTemplatesBoundATBs(kit *kit.Kit, bizID uint32, templateIDs []uint32) ([]*table.AppTemplateBinding, error)
	// ListTemplateSetInvisibleATBs list invisible atbs of the target template set when update its app visible scope.
	ListTemplateSetInvisibleATBs(kit *kit.Kit, bizID, templateSetID uint32, boundApps []uint32) (
		[]*table.AppTemplateBinding, error)
}

var _ TemplateBindingRelation = new(templateBindingRelationDao)

type templateBindingRelationDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

type countResult struct {
	Counts uint32 `json:"counts"`
}

// GetTemplateBoundUnnamedAppCount get bound unnamed app count of the target template.
func (dao *templateBindingRelationDao) GetTemplateBoundUnnamedAppCount(kit *kit.Kit, bizID, templateID uint32) (
	uint32, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.AppID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(templateID))...).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// GetTemplateBoundNamedAppCount get bound named app count of the target template.
func (dao *templateBindingRelationDao) GetTemplateBoundNamedAppCount(kit *kit.Kit, bizID, templateID uint32) (
	uint32, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.AppID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// GetTemplateBoundTemplateSetCount get bound template set count of the target template.
func (dao *templateBindingRelationDao) GetTemplateBoundTemplateSetCount(kit *kit.Kit, bizID, templateID uint32) (
	uint32, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.ID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(templateID))...).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// GetTemplateRevisionBoundUnnamedAppCount get bound unnamed app count of the target template release.
// nolint
func (dao *templateBindingRelationDao) GetTemplateRevisionBoundUnnamedAppCount(kit *kit.Kit, bizID,
	templateRevisionID uint32) (uint32, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.AppID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_revision_ids").Contains(templateRevisionID))...).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// GetTemplateRevisionBoundNamedAppCount get bound named app count of the target template release.
// nolint
func (dao *templateBindingRelationDao) GetTemplateRevisionBoundNamedAppCount(kit *kit.Kit, bizID,
	templateRevisionID uint32) (uint32, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.AppID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID), m.TemplateRevisionID.Eq(templateRevisionID)).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// GetTemplateSetBoundUnnamedAppCount get bound unnamed app count of the target template set.
func (dao *templateBindingRelationDao) GetTemplateSetBoundUnnamedAppCount(kit *kit.Kit, bizID,
	templateSetID uint32) (uint32, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.AppID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_set_ids").Contains(templateSetID))...).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// GetTemplateSetBoundNamedAppCount get bound named app count of the target template set.
func (dao *templateBindingRelationDao) GetTemplateSetBoundNamedAppCount(kit *kit.Kit, bizID, templateSetID uint32) (
	uint32, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var rs countResult
	if err := q.Select(m.AppID.Distinct().Count().As("counts")).
		Where(m.BizID.Eq(bizID), m.TemplateSetID.Eq(templateSetID)).
		Scan(&rs); err != nil {
		return 0, err
	}

	return rs.Counts, nil
}

// ListTmplBoundUnnamedApps list bound unnamed app details of the target template.
func (dao *templateBindingRelationDao) ListTmplBoundUnnamedApps(kit *kit.Kit, bizID, templateID uint32) (
	[]*types.TmplBoundUnnamedAppDetail, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var rs []*types.TmplBoundUnnamedAppDetail
	if err := q.Select(m.AppID, m.TemplateRevisionIDs).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(templateID))...).
		Scan(&rs); err != nil {
		return nil, err
	}

	return rs, nil
}

// ListTmplBoundNamedApps list bound named app details of the target template.
func (dao *templateBindingRelationDao) ListTmplBoundNamedApps(kit *kit.Kit, bizID, templateID uint32) (
	[]*types.TmplBoundNamedAppDetail, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var rs []*types.TmplBoundNamedAppDetail
	if err := q.Select(m.AppID, m.ReleaseID, m.TemplateRevisionID).
		Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).
		Scan(&rs); err != nil {
		return nil, err
	}

	return rs, nil
}

// ListTmplBoundTmplSets list bound template set details of the target template.
func (dao *templateBindingRelationDao) ListTmplBoundTmplSets(kit *kit.Kit, bizID, templateID uint32) (
	[]uint32, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	var templateSetIDs []uint32
	if err := q.Select(m.ID).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(templateID))...).
		Pluck(m.ID, &templateSetIDs); err != nil {
		return nil, err
	}

	return templateSetIDs, nil
}

// ListTmplRevisionBoundUnnamedApps list bound unnamed app details of the target template release.
func (dao *templateBindingRelationDao) ListTmplRevisionBoundUnnamedApps(kit *kit.Kit, bizID,
	templateRevisionID uint32) ([]uint32, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var appIDs []uint32
	if err := q.Select(m.AppID.Distinct()).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_revision_ids").Contains(templateRevisionID))...).
		Pluck(m.AppID, &appIDs); err != nil {
		return nil, err
	}

	return appIDs, nil
}

// ListTmplRevisionBoundNamedApps list bound named app details of the target template release.
func (dao *templateBindingRelationDao) ListTmplRevisionBoundNamedApps(kit *kit.Kit, bizID,
	templateRevisionID uint32) ([]*types.TmplRevisionBoundNamedAppDetail, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var rs []*types.TmplRevisionBoundNamedAppDetail
	if err := q.Select(m.AppID, m.ReleaseID).
		Where(m.BizID.Eq(bizID), m.TemplateRevisionID.Eq(templateRevisionID)).
		Scan(&rs); err != nil {
		return nil, err
	}

	return rs, nil
}

// ListTmplSetBoundUnnamedApps list bound unnamed app details of the target template set.
func (dao *templateBindingRelationDao) ListTmplSetBoundUnnamedApps(kit *kit.Kit, bizID,
	templateSetID uint32) ([]uint32, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	var appIDs []uint32
	if err := q.Select(m.AppID.Distinct()).
		Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_set_ids").Contains(templateSetID))...).
		Pluck(m.AppID, &appIDs); err != nil {
		return nil, err
	}

	return appIDs, nil
}

// ListTmplSetBoundNamedApps list bound named app details of the target template set.
func (dao *templateBindingRelationDao) ListTmplSetBoundNamedApps(kit *kit.Kit, bizID, templateSetID uint32) (
	[]*types.TmplSetBoundNamedAppDetail, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var rs []*types.TmplSetBoundNamedAppDetail
	if err := q.Select(m.AppID, m.ReleaseID).
		Where(m.BizID.Eq(bizID), m.TemplateSetID.Eq(templateSetID)).
		Scan(&rs); err != nil {
		return nil, err
	}

	return rs, nil
}

// ListLatestTmplBoundUnnamedApps list bound unnamed app details of the latest target template.
func (dao *templateBindingRelationDao) ListLatestTmplBoundUnnamedApps(kit *kit.Kit, bizID, templateID uint32) (
	[]*table.AppTemplateBinding, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("latest_template_ids").Contains(templateID))...).
		Find()
}

// ListTemplateSetsBoundATBs list bound app template bindings of the target template sets.
func (dao *templateBindingRelationDao) ListTemplateSetsBoundATBs(kit *kit.Kit, bizID uint32, templateSetIDs []uint32) (
	[]*table.AppTemplateBinding, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	details := make([]*table.AppTemplateBinding, 0)
	for _, id := range templateSetIDs {
		atbs, err := q.Where(m.BizID.Eq(bizID)).
			Where(rawgen.Cond(datatypes.JSONArrayQuery("template_set_ids").Contains(id))...).
			Find()
		if err != nil {
			return nil, err
		}
		details = append(details, atbs...)
	}
	if len(details) <= 1 {
		return details, nil
	}

	// remove duplicated atb
	uniqueDetails := make([]*table.AppTemplateBinding, 0)
	existMap := make(map[uint32]bool)
	for _, d := range details {
		if existMap[d.ID] {
			continue
		}
		existMap[d.ID] = true
		uniqueDetails = append(uniqueDetails, d)
	}

	return uniqueDetails, nil
}

// ListTemplatesBoundATBs list bound app template bindings of the target templates.
func (dao *templateBindingRelationDao) ListTemplatesBoundATBs(kit *kit.Kit, bizID uint32, templateIDs []uint32) (
	[]*table.AppTemplateBinding, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	details := make([]*table.AppTemplateBinding, 0)
	for _, id := range templateIDs {
		atbs, err := q.Where(m.BizID.Eq(bizID)).
			Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(id))...).
			Find()
		if err != nil {
			return nil, err
		}
		details = append(details, atbs...)
	}
	if len(details) <= 1 {
		return details, nil
	}

	// remove duplicated atb
	uniqueDetails := make([]*table.AppTemplateBinding, 0)
	existMap := make(map[uint32]bool)
	for _, d := range details {
		if existMap[d.ID] {
			continue
		}
		existMap[d.ID] = true
		uniqueDetails = append(uniqueDetails, d)
	}

	return uniqueDetails, nil
}

// ListTemplateSetInvisibleATBs list invisible atbs of the target template set when update its app visible scope.
func (dao *templateBindingRelationDao) ListTemplateSetInvisibleATBs(
	kit *kit.Kit, bizID, templateSetID uint32, boundApps []uint32) ([]*table.AppTemplateBinding, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_set_ids").Contains(templateSetID))...).
		Where(m.AppID.NotIn(boundApps...)).
		Find()
}
