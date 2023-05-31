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
	"strconv"
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	filter2 "bscp.io/pkg/runtime/filter"
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

	case *table.Hook:
		sset := cur.(*table.Hook)
		ab.toAudit.ResourceID = sset.ID

	case *table.TemplateSpace:
		sset := cur.(*table.TemplateSpace)
		ab.toAudit.ResourceID = sset.ID

	case *table.Group:
		sset := cur.(*table.Group)
		ab.toAudit.ResourceID = sset.ID

	case []*table.ReleasedConfigItem:
		items := cur.([]*table.ReleasedConfigItem)
		ab.toAudit.AppID = items[0].Attachment.AppID
		ab.toAudit.ResourceID = items[0].ReleaseID

	case *table.Credential:
		sset := cur.(*table.Credential)
		ab.toAudit.ResourceID = sset.ID

	case *table.CredentialScope:
		sset := cur.(*table.CredentialScope)
		ab.toAudit.ResourceID = sset.ID

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

	case *table.Group:
		group := updatedTo.(*table.Group)
		if err := ab.decorateGroupUpdate(group); err != nil {
			ab.hitErr = err
			return ab
		}

	case *table.Credential:
		credential := updatedTo.(*table.Credential)
		if err := ab.decorateCredentialUpdate(credential); err != nil {
			ab.hitErr = err
			return ab
		}

	case *table.CredentialScope:
		credentialScope := updatedTo.(*table.CredentialScope)
		if err := ab.decorateCredentialScopeUpdate(credentialScope); err != nil {
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

func (ab *AuditBuilder) decorateCredentialUpdate(credential *table.Credential) error {
	ab.toAudit.ResourceID = credential.ID

	preCredential, err := ab.getCredential(credential.ID)
	if err != nil {
		return err
	}

	ab.prev = preCredential

	changed, err := parseChangedSpecFields(preCredential, credential)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse credential changed spec field failed, err: %v", err)
	}

	ab.changed = changed
	return nil
}

func (ab *AuditBuilder) decorateCredentialScopeUpdate(credentialScope *table.CredentialScope) error {
	ab.toAudit.ResourceID = credentialScope.ID

	preCredential, err := ab.getCredentialScope(credentialScope.ID)
	if err != nil {
		return err
	}

	ab.prev = preCredential

	changed, err := parseChangedSpecFields(preCredential, credentialScope)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse credential scope changed spec field failed, err: %v", err)
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

	case enumor.Group:
		group, err := ab.getGroup(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.ResourceID = group.ID
		ab.prev = group

	case enumor.CRInstance:
		cri, err := ab.getCRInstance(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.AppID = cri.Attachment.AppID
		ab.toAudit.ResourceID = cri.ID
		ab.prev = cri

	case enumor.Credential:
		credential, err := ab.getCredential(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.ResourceID = credential.ID
		ab.prev = credential

	case enumor.CredentialScope:
		credentialScope, err := ab.getCredentialScope(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}
		ab.toAudit.ResourceID = credentialScope.ID
		ab.prev = credentialScope

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
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(), " WHERE id = ", strconv.Itoa(int(appID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.App)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getConfigItem(configItemID uint32) (*table.ConfigItem, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ConfigItemColumns.NamedExpr(), " FROM ", table.ConfigItemTable.Name(),
		" WHERE id = ", strconv.Itoa(int(configItemID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.ConfigItem)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get config item details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getStrategySet(strategySetID uint32) (*table.StrategySet, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.StrategySetColumns.NamedExpr(), " FROM ", table.StrategySetTable.Name(),
		" WHERE id = ", strconv.Itoa(int(strategySetID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.StrategySet)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get strategy set details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getStrategy(strategyID uint32) (*table.Strategy, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.StrategyColumns.NamedExpr(), " FROM ", table.StrategyTable.Name(),
		" WHERE id = ", strconv.Itoa(int(strategyID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.Strategy)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get strategy details failed, err: %v", err)
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

func (ab *AuditBuilder) getCRInstance(criID uint32) (*table.CurrentReleasedInstance, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CurrentReleasedInstanceColumns.NamedExpr(),
		" FROM ", table.CurrentReleasedInstanceTable.Name(),
		" WHERE id = ", strconv.Itoa(int(criID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.CurrentReleasedInstance)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get current released instance details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getCredential(credentialID uint32) (*table.Credential, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CredentialColumns.NamedExpr(), " FROM ", table.CredentialTable.Name(),
		" WHERE id = ", strconv.Itoa(int(credentialID)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.Credential)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get credential details failed, err: %v", err)
	}

	return one, nil
}

func (ab *AuditBuilder) getCredentialScope(id uint32) (*table.CredentialScope, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CredentialScopeColumns.NamedExpr(), " FROM ", table.CredentialScopeTable.Name(),
		" WHERE id = ", strconv.Itoa(int(id)), " AND biz_id = ", strconv.Itoa(int(ab.bizID)))
	filter := filter2.SqlJoint(sqlSentence)

	one := new(table.CredentialScope)
	err := ab.ad.orm.Do(ab.ad.sd.MustSharding(ab.bizID)).Get(ab.kit.Ctx, one, filter)
	if err != nil {
		return nil, fmt.Errorf("get credential scope details failed, err: %v", err)
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
