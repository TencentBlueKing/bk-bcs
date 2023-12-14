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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	filter2 "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/filter"
)

// initAuditBuilder create a new audit builder instance.
func initAuditBuilder(kit *kit.Kit, bizID uint32, res enumor.AuditResourceType, ad *audit) AuditDecorator {

	ab := &AuditBuilder{
		toAudit: &table.Audit{
			BizID:        bizID,
			ResourceType: res,
			CreatedAt:    time.Now().UTC(),
			Operator:     kit.User,
			Rid:          kit.Rid,
			AppCode:      kit.AppCode,
		},
		ad:    ad,
		bizID: bizID,
		kit:   kit,
	}

	if bizID <= 0 {
		ab.hitErr = errors.New("invalid audit biz id")
	}

	if len(kit.User) == 0 {
		ab.hitErr = errors.New("invalid audit operator")
	}

	return ab
}

// AuditDecorator is audit decorator interface, use to record audit.
type AuditDecorator interface {
	AuditCreate(cur interface{}, opt *AuditOption) error
	PrepareUpdate(updatedTo interface{}) AuditDecorator
	PrepareDelete(resID uint32) AuditDecorator
	AuditPublish(cur interface{}, opt *AuditOption) error
	Do(opt *AuditOption) error
}

// AuditBuilder is a wrapper decorator to handle all the resource's
// audit operation.
type AuditBuilder struct {
	hitErr error

	toAudit *table.Audit
	bizID   uint32
	kit     *kit.Kit
	prev    interface{}
	changed map[string]interface{}
	ad      *audit
}

// AuditCreate set the resource's current details.
// Note:
// 1. must call this after the resource has already been created.
// 2. cur should be a *struct.
func (ab *AuditBuilder) AuditCreate(cur interface{}, opt *AuditOption) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	ab.toAudit.Action = enumor.Create
	ab.prev = cur

	switch val := cur.(type) {
	case *table.ConfigItem:
		configItem := val
		ab.toAudit.AppID = configItem.Attachment.AppID
		ab.toAudit.ResourceID = configItem.ID

	case *table.Content:
		content := val
		ab.toAudit.AppID = content.Attachment.AppID
		ab.toAudit.ResourceID = content.ID

	case *table.Commit:
		commit := val
		ab.toAudit.AppID = commit.Attachment.AppID
		ab.toAudit.ResourceID = commit.ID

	case *table.Release:
		release := val
		ab.toAudit.AppID = release.Attachment.AppID
		ab.toAudit.ResourceID = release.ID

	case *table.Hook:
		sset := val
		ab.toAudit.ResourceID = sset.ID

	case *table.TemplateSpace:
		sset := val
		ab.toAudit.ResourceID = sset.ID

	case *table.Group:
		sset := val
		ab.toAudit.ResourceID = sset.ID

	case []*table.ReleasedConfigItem:
		items := val
		ab.toAudit.AppID = items[0].Attachment.AppID
		ab.toAudit.ResourceID = items[0].ReleaseID

	default:
		logs.Errorf("unsupported audit create resource: %s, type: %s, rid: %v", ab.toAudit.ResourceType,
			reflect.TypeOf(cur), ab.toAudit.Rid)
		return fmt.Errorf("unsupported audit create resource: %s", ab.toAudit.ResourceType)
	}

	detail := &table.AuditBasicDetail{
		Prev:    ab.prev,
		Changed: nil,
	}
	js, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("marshal audit detail failed, err: %v", err)
	}
	ab.toAudit.Detail = string(js)

	return ab.ad.One(ab.kit, ab.toAudit, opt)
}

// AuditPublish set the publish content's current details.
// Note:
// 1. must call this after the resource has already been published.
// 2. cur should be a *struct.
func (ab *AuditBuilder) AuditPublish(cur interface{}, opt *AuditOption) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	ab.toAudit.Action = enumor.Publish
	ab.prev = cur

	switch val := cur.(type) {
	case *table.Strategy:
		strategy := val
		ab.toAudit.AppID = strategy.Attachment.AppID
		ab.toAudit.ResourceID = strategy.ID

	default:
		logs.Errorf("unsupported audit publish resource: %s, type: %s, rid: %v", ab.toAudit.ResourceType,
			reflect.TypeOf(cur), ab.toAudit.Rid)
		return fmt.Errorf("unsupported audit publish resource: %s", ab.toAudit.ResourceType)
	}

	detail := &table.AuditBasicDetail{
		Prev:    ab.prev,
		Changed: nil,
	}
	js, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("marshal audit detail failed, err: %v", err)
	}
	ab.toAudit.Detail = string(js)

	return ab.ad.One(ab.kit, ab.toAudit, opt)
}

// PrepareUpdate prepare the resource's previous instance details by
// get the instance's detail from db and save it to ab.prev for later use.
// Note:
// 1. call this before resource is updated.
// 2. updatedTo means 'to be updated to data', it should be a *struct.
func (ab *AuditBuilder) PrepareUpdate(updatedTo interface{}) AuditDecorator {
	if ab.hitErr != nil {
		return ab
	}

	ab.toAudit.Action = enumor.Update

	switch val := updatedTo.(type) {
	case *table.ConfigItem:
		ci := val
		if err := ab.decorateConfigItemUpdate(ci); err != nil {
			ab.hitErr = err
			return ab
		}

	case *table.Group:
		group := val
		if err := ab.decorateGroupUpdate(group); err != nil {
			ab.hitErr = err
			return ab
		}

	default:
		logs.Errorf("unsupported audit update resource: %s, type: %s, rid: %v", ab.toAudit.ResourceType,
			reflect.TypeOf(updatedTo), ab.toAudit.Rid)
		ab.hitErr = fmt.Errorf("unsupported audit update resource: %s", ab.toAudit.ResourceType)
		return ab
	}

	detail := &table.AuditBasicDetail{
		Prev:    ab.prev,
		Changed: ab.changed,
	}

	js, err := json.Marshal(detail)
	if err != nil {
		ab.hitErr = fmt.Errorf("marshal audit detail failed, err: %v", err)
		return ab
	}
	ab.toAudit.Detail = string(js)

	return ab
}

func (ab *AuditBuilder) decorateConfigItemUpdate(ci *table.ConfigItem) error {
	ab.toAudit.AppID = ci.Attachment.AppID
	ab.toAudit.ResourceID = ci.ID

	prevCI, err := ab.getConfigItem(ci.ID)
	if err != nil {
		return err
	}

	ab.prev = prevCI

	changed, err := parseChangedSpecFields(prevCI, ci)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse config item changed spec field failed, err: %v", err)
	}

	ab.changed = changed
	return nil
}

func (ab *AuditBuilder) decorateGroupUpdate(group *table.Group) error {
	ab.toAudit.ResourceID = group.ID

	preGroup, err := ab.getGroup(group.ID)
	if err != nil {
		return err
	}

	ab.prev = preGroup

	changed, err := parseChangedSpecFields(preGroup, group)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse group changed spec field failed, err: %v", err)
	}

	ab.changed = changed
	return nil
}

// PrepareDelete prepare the resource's previous instance details by
// get the instance's detail from db and save it to ab.prev for later use.
// Note: call this before resource is deleted.
func (ab *AuditBuilder) PrepareDelete(resID uint32) AuditDecorator {
	if ab.hitErr != nil {
		return ab
	}

	ab.toAudit.Action = enumor.Delete

	switch ab.toAudit.ResourceType {
	case enumor.ConfigItem:
		configItem, err := ab.getConfigItem(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = configItem.Attachment.AppID
		ab.toAudit.ResourceID = configItem.ID
		ab.prev = configItem

	case enumor.Group:
		group, err := ab.getGroup(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.ResourceID = group.ID
		ab.prev = group

	default:
		ab.hitErr = fmt.Errorf("unsupported audit deleted resource: %s", ab.toAudit.ResourceType)
		return ab
	}

	detail := &table.AuditBasicDetail{
		Prev:    ab.prev,
		Changed: nil,
	}

	js, err := json.Marshal(detail)
	if err != nil {
		ab.hitErr = fmt.Errorf("marshal audit detail failed, err: %v", err)
		return ab
	}
	ab.toAudit.Detail = string(js)

	return ab
}

// Do save audit log to the db immediately.
func (ab *AuditBuilder) Do(opt *AuditOption) error {

	if ab.hitErr != nil {
		return ab.hitErr
	}

	return ab.ad.One(ab.kit, ab.toAudit, opt)

}

func (ab *AuditBuilder) getConfigItem(configItemID uint32) (*table.ConfigItem, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ConfigItemColumns.NamedExpr(), " FROM ",
		table.ConfigItemTable.Name(), " WHERE id = ", strconv.Itoa(int(configItemID)), " AND biz_id = ",
		strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.ConfigItem)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get config item details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getGroup(groupID uint32) (*table.Group, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(),
		" WHERE id = ", strconv.Itoa(int(groupID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.Group)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get group details failed, err: %v", err)
	}

	return one, nil
}

// parseChangedSpecFields parse the changed filed with pre and cur *structs' Spec field.
// both pre and curl should be a *struct, if not, it will 'panic'.
// Note:
//  1. the pre and cur should be the same structs' pointer, and should
//     have a 'Spec' struct field.
//  2. this func only compare 'Spec' field.
//  3. if one of the cur's Spec's filed value is zero, then this filed will be ignored.
//  4. the returned update field's key is this field's 'db' tag.
func parseChangedSpecFields(pre, cur interface{}) (map[string]interface{}, error) {
	preV := reflect.ValueOf(pre)
	if preV.Kind() != reflect.Ptr {
		return nil, errors.New("parse changed spec field, but pre data is not a *struct")
	}

	curV := reflect.ValueOf(cur)
	if curV.Kind() != reflect.Ptr {
		return nil, errors.New("parse changed spec field, but cur data is not a *struct")
	}

	// make sure the pre and data is the same struct.
	if !reflect.TypeOf(pre).AssignableTo(reflect.TypeOf(cur)) {
		return nil, errors.New("parse changed spec field, but pre and cur resource type is not different, " +
			"can not be compared")
	}

	prevSpec := preV.Elem().FieldByName("Spec")
	curSpec := curV.Elem().FieldByName("Spec")
	if prevSpec.IsZero() || curSpec.IsZero() {
		return nil, errors.New("pre or cur data do not has a 'Spec' struct field")
	}

	prevSpecV := prevSpec.Elem()
	curSpecV := curSpec.Elem()
	changedField := make(map[string]interface{})

	// compare spec's detail
	for i := 0; i < prevSpecV.NumField(); i++ {
		preName := prevSpecV.Type().Field(i).Name
		curFieldV := curSpecV.FieldByName(preName)
		if curFieldV.IsZero() {
			// if this filed value is a zero value, then skip it.
			// which means it is not updated.
			continue
		}

		if reflect.DeepEqual(prevSpecV.Field(i).Interface(), curFieldV.Interface()) {
			// this field's value is not changed.
			continue
		}

		dbTag := prevSpecV.Type().Field(i).Tag.Get("db")
		if len(dbTag) == 0 {
			// fallback to json tag
			dbTag = prevSpecV.Type().Field(i).Tag.Get("json")
		}

		if len(dbTag) == 0 {
			return nil, fmt.Errorf("filed: %s do not have a db or json tag, can not compare", preName)
		}

		changedField[dbTag] = curFieldV.Interface()
	}

	return changedField, nil
}
