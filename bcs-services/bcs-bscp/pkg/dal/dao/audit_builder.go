/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
)

// initAuditBuilder create a new audit builder instance.
func initAuditBuilder(kit *kit.Kit, bizID uint32, res enumor.AuditResourceType, ad *audit) AuditDecorator {

	ab := &AuditBuilder{
		toAudit: &table.Audit{
			BizID:        bizID,
			ResourceType: res,
			CreatedAt:    time.Now(),
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
	AuditFinishPublish(strID, appID uint32, opt *AuditOption) error
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

	switch cur.(type) {
	case *table.App:
		app := cur.(*table.App)
		ab.toAudit.AppID = app.ID
		ab.toAudit.ResourceID = app.ID

	case *table.ConfigItem:
		configItem := cur.(*table.ConfigItem)
		ab.toAudit.AppID = configItem.Attachment.AppID
		ab.toAudit.ResourceID = configItem.ID

	case *table.Content:
		content := cur.(*table.Content)
		ab.toAudit.AppID = content.Attachment.AppID
		ab.toAudit.ResourceID = content.ID

	case *table.Commit:
		commit := cur.(*table.Commit)
		ab.toAudit.AppID = commit.Attachment.AppID
		ab.toAudit.ResourceID = commit.ID

	case *table.Release:
		release := cur.(*table.Release)
		ab.toAudit.AppID = release.Attachment.AppID
		ab.toAudit.ResourceID = release.ID

	case *table.Strategy:
		strategy := cur.(*table.Strategy)
		ab.toAudit.AppID = strategy.Attachment.AppID
		ab.toAudit.ResourceID = strategy.ID

	case *table.StrategySet:
		sset := cur.(*table.StrategySet)
		ab.toAudit.AppID = sset.Attachment.AppID
		ab.toAudit.ResourceID = sset.ID

	case []*table.ReleasedConfigItem:
		items := cur.([]*table.ReleasedConfigItem)
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

	switch cur.(type) {
	case *table.Strategy:
		strategy := cur.(*table.Strategy)
		ab.toAudit.AppID = strategy.Attachment.AppID
		ab.toAudit.ResourceID = strategy.ID

	case *table.CurrentReleasedInstance:
		cri := cur.(*table.CurrentReleasedInstance)
		ab.toAudit.AppID = cri.Attachment.AppID
		ab.toAudit.ResourceID = cri.ID

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

// AuditFinishPublish set finish publish strategy id.
// Note:
// 1. must call this after the resource has already been finish published.
func (ab *AuditBuilder) AuditFinishPublish(strID, appID uint32, opt *AuditOption) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	ab.toAudit.Action = enumor.FinishPublish

	ab.toAudit.AppID = appID
	ab.toAudit.ResourceID = strID

	detail := &table.AuditBasicDetail{
		Prev:    nil,
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

	switch updatedTo.(type) {
	case *table.App:
		app := updatedTo.(*table.App)
		if err := ab.decorateAppUpdate(app); err != nil {
			ab.hitErr = err
			return ab
		}

	case *table.ConfigItem:
		ci := updatedTo.(*table.ConfigItem)
		if err := ab.decorateConfigItemUpdate(ci); err != nil {
			ab.hitErr = err
			return ab
		}

	case *table.StrategySet:
		ss := updatedTo.(*table.StrategySet)
		if err := ab.decorateStrategySetUpdate(ss); err != nil {
			ab.hitErr = err
			return ab
		}

	case *table.Strategy:
		strategy := updatedTo.(*table.Strategy)
		if err := ab.decorateStrategyUpdate(strategy); err != nil {
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

func (ab *AuditBuilder) decorateAppUpdate(app *table.App) error {
	ab.toAudit.AppID = app.ID
	ab.toAudit.ResourceID = app.ID

	prevApp, err := ab.getApp(app.ID)
	if err != nil {
		return err
	}

	ab.prev = prevApp

	changed, err := parseChangedSpecFields(prevApp, app)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse app changed spec field failed, err: %v", err)
	}

	ab.changed = changed
	return nil
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

func (ab *AuditBuilder) decorateStrategySetUpdate(ss *table.StrategySet) error {
	ab.toAudit.AppID = ss.Attachment.AppID
	ab.toAudit.ResourceID = ss.ID

	prevSS, err := ab.getStrategySet(ss.ID)
	if err != nil {
		return err
	}

	ab.prev = prevSS

	changed, err := parseChangedSpecFields(prevSS, ss)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse strategy set changed spec field failed, err: %v", err)
	}

	ab.changed = changed
	return nil
}

func (ab *AuditBuilder) decorateStrategyUpdate(strategy *table.Strategy) error {
	ab.toAudit.AppID = strategy.Attachment.AppID
	ab.toAudit.ResourceID = strategy.ID

	preStrategy, err := ab.getStrategy(strategy.ID)
	if err != nil {
		return err
	}

	ab.prev = preStrategy

	changed, err := parseChangedSpecFields(preStrategy, strategy)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse strategy changed spec field failed, err: %v", err)
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
	case enumor.App:
		app, err := ab.getApp(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = app.ID
		ab.toAudit.ResourceID = app.ID
		ab.prev = app

	case enumor.ConfigItem:
		configItem, err := ab.getConfigItem(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = configItem.Attachment.AppID
		ab.toAudit.ResourceID = configItem.ID
		ab.prev = configItem

	case enumor.Strategy:
		strategy, err := ab.getStrategy(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = strategy.Attachment.AppID
		ab.toAudit.ResourceID = strategy.ID
		ab.prev = strategy

	case enumor.StrategySet:
		ss, err := ab.getStrategySet(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = ss.Attachment.AppID
		ab.toAudit.ResourceID = ss.ID
		ab.prev = ss

	case enumor.CRInstance:
		cri, err := ab.getCRInstance(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = cri.Attachment.AppID
		ab.toAudit.ResourceID = cri.ID
		ab.prev = cri

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

func (ab *AuditBuilder) getApp(appID uint32) (*table.App, error) {
	filter := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d`,
		table.AppColumns.NamedExpr(), table.AppTable, appID, ab.bizID)

	one := new(table.App)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getConfigItem(configItemID uint32) (*table.ConfigItem, error) {
	filter := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d`,
		table.ConfigItemColumns.NamedExpr(), table.ConfigItemTable, configItemID, ab.bizID)

	one := new(table.ConfigItem)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get config item details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getStrategySet(strategySetID uint32) (*table.StrategySet, error) {
	filter := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d`,
		table.StrategySetColumns.NamedExpr(), table.StrategySetTable, strategySetID, ab.bizID)

	one := new(table.StrategySet)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get strategy set details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getStrategy(strategyID uint32) (*table.Strategy, error) {
	filter := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d`,
		table.StrategyColumns.NamedExpr(), table.StrategyTable, strategyID, ab.bizID)

	one := new(table.Strategy)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get strategy details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getCRInstance(criID uint32) (*table.CurrentReleasedInstance, error) {
	filter := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d`,
		table.CurrentReleasedInstanceColumns.NamedExpr(), table.CurrentReleasedInstanceTable, criID, ab.bizID)

	one := new(table.CurrentReleasedInstance)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get current released instance details failed, err: %v", err)
	}

	return one, nil
}

// parseChangedSpecFields parse the changed filed with pre and cur *structs' Spec field.
// both pre and curl should be a *struct, if not, it will 'panic'.
// Note:
// 1. the pre and cur should be the same structs' pointer, and should
//    have a 'Spec' struct field.
// 2. this func only compare 'Spec' field.
// 3. if one of the cur's Spec's filed value is zero, then this filed will be ignored.
// 4. the returned update field's key is this field's 'db' tag.
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
			return nil, fmt.Errorf("filed: %s do not have a db tag, can not compare", preName)
		}

		changedField[dbTag] = curFieldV.Interface()
	}

	return changedField, nil
}
