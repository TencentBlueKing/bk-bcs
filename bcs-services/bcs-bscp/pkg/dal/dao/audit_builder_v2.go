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
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// AuditDo
type AuditDo interface {
	Do(tx *gen.Query) error
}

// AuditRes :
type AuditRes interface {
	AppID() uint32
	ResourceType() string
	ResourceID() uint32
}

// AuditPrepare
type AuditPrepare interface {
	PrepareCreate(obj AuditRes) AuditDo
	PrepareUpdate(obj AuditRes, oldObj interface{}) AuditDo
	PrepareDelete(obj AuditRes) AuditDo
}

// initAuditBuilder create a new audit builder instance.
func initAuditBuilderV2(kit *kit.Kit, bizID uint32, ad *audit) AuditPrepare {
	ab := &AuditBuilderV2{
		toAudit: &table.Audit{
			BizID:     bizID,
			CreatedAt: time.Now(),
			Operator:  kit.User,
			Rid:       kit.Rid,
			AppCode:   kit.AppCode,
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

// AuditBuilder is a wrapper decorator to handle all the resource's
// audit operation.
type AuditBuilderV2 struct {
	hitErr error

	toAudit *table.Audit
	bizID   uint32
	kit     *kit.Kit
	prev    interface{}
	changed map[string]interface{}
	ad      *audit
}

// Do save audit log to the db immediately.
func (ab *AuditBuilderV2) Do(tx *gen.Query) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	return ab.ad.One(ab.kit, ab.toAudit, &AuditOption{genM: tx})
}

func (ab *AuditBuilderV2) PrepareCreate(obj AuditRes) AuditDo {
	ab.toAudit.ResourceType = enumor.AuditResourceType(obj.ResourceType())
	ab.toAudit.ResourceID = obj.ResourceID()
	ab.toAudit.Action = enumor.Create
	ab.prev = obj

	detail := &table.AuditBasicDetail{
		Prev:    ab.prev,
		Changed: nil,
	}

	js, err := json.Marshal(detail)
	if err != nil {
		ab.hitErr = err
		return ab
	}
	ab.toAudit.Detail = string(js)

	return ab
}

func (ab *AuditBuilderV2) PrepareUpdate(obj AuditRes, oldObj interface{}) AuditDo {
	ab.toAudit.ResourceType = enumor.AuditResourceType(obj.ResourceType())
	ab.toAudit.ResourceID = obj.ResourceID()
	ab.toAudit.Action = enumor.Update
	ab.prev = oldObj

	changed, err := parseChangedSpecFields(oldObj, obj)
	if err != nil {
		ab.hitErr = err
		return ab
	}
	ab.changed = changed

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

func (ab *AuditBuilderV2) PrepareDelete(obj AuditRes) AuditDo {
	ab.toAudit.ResourceType = enumor.AuditResourceType(obj.ResourceType())
	ab.toAudit.ResourceID = obj.ResourceID()
	ab.toAudit.Action = enumor.Delete
	ab.prev = obj

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
