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

package table

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

const (
	// PublishByGroups means that publish given groups.
	PublishByGroups GrayPublishMode = "publish_by_groups"

	// PublishByLabels means that publish by given labels.
	PublishByLabels GrayPublishMode = "publish_by_labels"
)

// GrayPublishMode defines gray publish mode.
type GrayPublishMode string

// String returns gray publish mode string.
func (s GrayPublishMode) String() string {
	return string(s)
}

// Validate gray publish mode type.
func (s GrayPublishMode) Validate() error {
	switch s {
	case PublishByGroups:
	case PublishByLabels:
	default:
		return fmt.Errorf("unsupported publish mode: %s", s)
	}

	return nil
}

// Strategy defines a strategy for an app to publish.
// it contains the basic released information and the
// selector to define the scope of the matched instances.
type Strategy struct {
	// ID is an auto-increased value, which is a unique identity
	// of a strategy.
	ID         uint32              `db:"id" json:"id" gorm:"primaryKey"`
	Spec       *StrategySpec       `db:"spec" json:"spec" gorm:"embedded"`
	State      *StrategyState      `db:"state" json:"state" gorm:"embedded"`
	Attachment *StrategyAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision   *Revision           `db:"revision" json:"revision" gorm:"embedded"`
}

// TableName is the strategy's database table name.
func (s *Strategy) TableName() string {
	return "strategies"
}

// AppID AuditRes interface
func (s *Strategy) AppID() uint32 {
	return s.Attachment.AppID
}

// ResID AuditRes interface
func (s *Strategy) ResID() uint32 {
	return s.ID
}

// ResType AuditRes interface
func (s *Strategy) ResType() string {
	return "strategy"
}

// ValidateCreate validate strategy is valid or not when create it.
func (s *Strategy) ValidateCreate(kit *kit.Kit) error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if s.State == nil {
		return errors.New("state not set")
	}

	if err := s.State.Validate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := s.Attachment.Validate(); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate strategy is valid or not when update it.
func (s *Strategy) ValidateUpdate(kit *kit.Kit, asDefault bool, namespaced bool) error {

	if s.ID <= 0 {
		return errors.New("id should be set")
	}

	changed := false
	if s.Spec != nil {
		changed = true
		if err := s.Spec.ValidateUpdate(kit, asDefault, namespaced); err != nil {
			return err
		}
	}

	if s.State != nil {
		changed = true
		if err := s.State.Validate(); err != nil {
			return err
		}
	}

	if s.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if s.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	if !changed {
		return errors.New("nothing is found to be change")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the strategy's info when delete it.
func (s *Strategy) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("strategy id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

const (
	// ReservedNamespacePrefix defines the reserved namespaces which
	// is prefixed with 'bscp'.
	ReservedNamespacePrefix = "bscp"

	// DefaultNamespace is the default namespace's value when a strategy is
	// set to default strategy and works at the namespace mode at the same
	// time.
	DefaultNamespace = "bscp_default_ns"
)

// StrategySpec defines all the specifics for strategy set by user.
type StrategySpec struct {
	Name      string `db:"name" json:"name" gorm:"column:name"`
	ReleaseID uint32 `db:"release_id" json:"release_id" gorm:"column:release_id"`
	// AsDefault(=true) describes this strategy works as full release,
	// which means any instance can match this strategies
	AsDefault bool `db:"as_default" json:"as_default" gorm:"column:as_default"`

	// Scope must be empty when this strategy is a default strategy.
	// Scope must not be empty when this strategy is not a default strategy.
	Scope *Scope `db:"scope" json:"scope" gorm:"column:scope;type:json"`

	// Namespace defines which namespace this strategy works at.
	// It has the following features:
	// 1. if a strategy set works at namespace mode, then all the strategy
	//    belongs to it must be working at namespace mode at the same time.
	//    which means if StrategySpec.Mode = namespace, then StrategySpec.Namespace
	//    must not be empty.
	// 2. all the namespace in a strategy among the same strategy set is unique.
	//    it is not allowed to be duplicate.
	// 3. if this strategy is set to default strategy and works at namespace mode,
	//    then its namespace should be the reserved namespace DefaultNamespace(
	//    'bscp_default_ns')
	Namespace         string        `db:"namespace" json:"namespace" gorm:"column:namespace"`
	Memo              string        `db:"memo" json:"memo" gorm:"column:memo"`
	PublishType       PublishType   `db:"publish_type" json:"publish_type" gorm:"column:publish_type"`
	PublishTime       string        `db:"publish_time" json:"publish_time" gorm:"column:publish_time"`
	PublishStatus     PublishStatus `db:"publish_status" json:"publish_status" gorm:"column:publish_status"`
	RejectReason      string        `db:"reject_reason" json:"reject_reason" gorm:"column:reject_reason"`
	Approver          string        `db:"approver" json:"approver" approver:"column:approver"`
	ApproverProgress  string        `db:"approver_progress" json:"approver_progress" gorm:"column:approver_progress"`
	ItsmTicketType    string        `db:"itsm_ticket_type" json:"itsm_ticket_type" gorm:"column:itsm_ticket_type"`
	ItsmTicketUrl     string        `db:"itsm_ticket_url" json:"itsm_ticket_url" gorm:"column:itsm_ticket_url"`
	ItsmTicketSn      string        `db:"itsm_ticket_sn" json:"itsm_ticket_sn" gorm:"column:itsm_ticket_sn"`
	ItsmTicketStatus  string        `db:"itsm_ticket_status" json:"itsm_ticket_status" gorm:"column:itsm_ticket_status"`
	ItsmTicketStateID int           `db:"itsm_ticket_state_id" json:"itsm_ticket_state_id" gorm:"column:itsm_ticket_state_id"` // nolint
}

// ValidateCreate validate strategy spec when it is created.
func (s StrategySpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, s.Name); err != nil {
		return err
	}

	if s.ReleaseID <= 0 {
		return errors.New("invalid strategy release id")
	}

	if !s.AsDefault {
		if len(s.Scope.Groups) == 0 {
			return errors.New("strategy's scope can not be empty at gray release mode")
		}
		for _, group := range s.Scope.Groups {
			if err := group.ValidateCreate(kit); err != nil {
				return err
			}
		}
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	if err := s.ValidateSubmitPublishContent(); err != nil {
		return err
	}

	return nil
}

// ValidateSubmitPublishContent validate strategy spec public content when it is submit.
func (s StrategySpec) ValidateSubmitPublishContent() error {
	if err := s.PublishType.ValidatePublishType(); err != nil {
		return err
	}

	if err := s.PublishStatus.ValidatePublishStatus(); err != nil {
		return err
	}

	if s.PublishTime != "" {
		_, err := time.Parse(time.DateTime, s.PublishTime)
		if err != nil {
			return fmt.Errorf("publish time format invalid: %s, err: %s", s.PublishTime, err)
		}
	}

	// validate publish_type
	if s.Approver != "" && s.PublishType == Immediately {
		return fmt.Errorf("app need approve, invalid publish_type: %s", s.PublishType)
	}

	// app donnot need approve
	if s.Approver == "" &&
		(s.PublishType == Automatically || s.PublishType == Manually) {
		return fmt.Errorf("app donnot need approve, invalid publish_type: %s", s.PublishType)
	}

	return nil
}

// ValidateUpdate validate strategy spec when it is updated.
func (s StrategySpec) ValidateUpdate(kit *kit.Kit, asDefault bool, namespaced bool) error {

	if len(s.Name) != 0 {
		if err := validator.ValidateName(kit, s.Name); err != nil {
			return err
		}
	}

	if s.ReleaseID <= 0 {
		return errors.New("release id should be set")
	}

	if len(s.Namespace) != 0 {
		return errors.New("namespace can not be updated")
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	return nil
}

const (
	// Unpublished means this strategy is not published yet.
	// which means a strategy has not does any publish operation
	// before. this state exist only for once for a strategy.
	Unpublished PublishState = "unpublished"
	// Publishing means this strategy is during the publish
	// process, but have not finished.
	Publishing PublishState = "publishing"
	// Published means this strategy has already finishes the
	// publish process by the user.
	Published PublishState = "published"
)

// PublishState defines an app's strategy publish state.
type PublishState string

// Validate whether publish state is valid or not.
func (p PublishState) Validate() error {

	switch p {
	case Unpublished:
	case Publishing:
	case Published:
	default:
		return fmt.Errorf("unsupported publish state: %s", p)
	}

	return nil
}

// StrategyState defines the strategy's state
type StrategyState struct {
	PubState PublishState `db:"pub_state" json:"pub_state" gorm:"column:pub_state"`
}

// Validate whether strategy state is valid or not.
func (s StrategyState) Validate() error {
	if err := s.PubState.Validate(); err != nil {
		return err
	}

	return nil
}

// StrategyAttachment defines the strategy attachments.
type StrategyAttachment struct {
	BizID         uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID         uint32 `db:"app_id" json:"app_id" gorm:"column:app_id"`
	StrategySetID uint32 `db:"strategy_set_id" json:"strategy_set_id" gorm:"column:strategy_set_id"`
}

// IsEmpty test whether strategy attachment is empty or not.
func (s StrategyAttachment) IsEmpty() bool {
	return s.BizID == 0 && s.AppID == 0 && s.StrategySetID == 0
}

// Validate whether strategy attachment is valid or not.
func (s StrategyAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if s.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	if s.StrategySetID <= 0 {
		return errors.New("invalid attachment strategy set id")
	}

	return nil
}

// MaxScopeSelectorByteSize is the max size of a scope selector in byte.
// as is 1 KB.
const MaxScopeSelectorByteSize = 1 * 1024

// Scope defines a strategy's working groups.
type Scope struct {
	// Groups defines strategys's working scope
	Groups []*Group `db:"groups" json:"groups"`
}

// Scan is used to decode raw message which is read from db into a structured
// ScopeSelector instance.
func (s *Scope) Scan(raw interface{}) error {
	if s == nil {
		return errors.New("scope is not initialized")
	}

	if raw == nil {
		return errors.New("raw is nil, can not be decoded")
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("decode into scope failed, err: %v", err)

		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return fmt.Errorf("decode into scope failed, err: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported scope raw type: %T", v)
	}
}

// Value encode the scope selector to a json raw, so that it can be stored to db with
// json raw.
func (s *Scope) Value() (driver.Value, error) {
	if s == nil {
		return nil, errors.New("scope selector is not initialized, can not be encoded")
	}

	return json.Marshal(s)
}

// IsEmpty test whether this scope selector is empty or not.
func (s Scope) IsEmpty() bool {
	return len(s.Groups) == 0
}

// ValidateCreate validate strategy's selector when it is created.
func (s Scope) ValidateCreate(asDefault bool, namespaced bool) error {

	if s.IsEmpty() {
		return errors.New("strategy's groups is not set")
	}

	return nil
}

// ValidateUpdate validate strategy's selector when it is updated.
func (s Scope) ValidateUpdate(asDefault bool, namespaced bool) error {
	return nil
}

// SubStrategy is the sub-strategy of its parent strategy, it can not be
// used independently.
type SubStrategy struct {
	Spec *SubStrategySpec `db:"spec" json:"spec" gorm:"column:name"`
}

// IsEmpty test whether a sub-strategy is empty or not.
func (s SubStrategy) IsEmpty() bool {
	if s.Spec == nil {
		return true
	}

	if s.Spec != nil {
		if !s.Spec.IsEmpty() {
			return false
		}
	}

	return true
}

// ValidateCreate validate sub strategy when it is created.
func (s SubStrategy) ValidateCreate(kit *kit.Kit) error {
	if s.Spec == nil {
		return errors.New("sub strategy's spec is empty")
	}

	if err := s.Spec.Validate(kit); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate sub strategy when it is updated.
func (s SubStrategy) ValidateUpdate(kit *kit.Kit) error {
	if s.Spec == nil {
		return errors.New("sub strategy's spec is empty")
	}

	if err := s.Spec.Validate(kit); err != nil {
		return err
	}

	return nil
}

// SubStrategySpec is the sub-strategy's specifics defined by user.
type SubStrategySpec struct {
	Name string `db:"name" json:"name" gorm:"column:name"`
	// ReleaseID this sub strategy's released version id.
	ReleaseID uint32            `db:"release_id" json:"release_id" gorm:"column:release_id"`
	Scope     *SubScopeSelector `db:"scope" json:"scope" gorm:"embedded"`
	Memo      string            `db:"memo" json:"memo" gorm:"column:memo"`
}

// IsEmpty test whether a sub-strategy specific is empty or not.
func (s SubStrategySpec) IsEmpty() bool {
	if len(s.Name) != 0 {
		return false
	}

	if s.ReleaseID > 0 {
		return false
	}

	if s.Scope != nil {
		if !s.Scope.IsEmpty() {
			return false
		}
	}

	if len(s.Memo) != 0 {
		return false
	}

	return true
}

// Validate the sub strategy's specifics
func (s SubStrategySpec) Validate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, s.Name); err != nil {
		return err
	}

	if s.ReleaseID <= 0 {
		return errors.New("invalid sub strategy's release id")
	}

	if s.Scope == nil {
		return errors.New("sub strategy's scope is empty")
	}

	if err := s.Scope.Validate(); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	return nil
}

// SubScopeSelector is the sub-strategy's scope selector.
type SubScopeSelector struct {
	// Selector's scope must be part of the instances, which means
	// this select should not be matched all policy. and this selector
	// is required, should not be empty.
	// this selector has a max size limit, as is MaxScopeSelectorByteSize byte.
	Selector *selector.Selector `db:"selector" json:"selector" gorm:"column:selector;type:json"`
}

// IsEmpty test whether a sub-scope selector is empty or not.
func (s SubScopeSelector) IsEmpty() bool {
	if s.Selector == nil {
		return true
	}

	return s.Selector.IsEmpty()
}

// ErrSelectorByteSizeIsOverMaxLimit means the selector's byte size is over max
// limit.
var ErrSelectorByteSizeIsOverMaxLimit = errors.New("the selector's byte size is over the max limit error")

// Validate the sub scope selector
func (s SubScopeSelector) Validate() error {
	if s.Selector == nil {
		return errors.New("sub scope selector can not be empty")
	}

	if s.Selector.IsEmpty() {
		return errors.New("sub scope selector is empty, it is required")
	}

	if s.Selector.MatchAll {
		return errors.New("sub strategy's scope selector can not use match all, " +
			"should be part of all the instances")
	}

	raw, err := json.Marshal(s.Selector)
	if err != nil {
		return fmt.Errorf("marshal sub strategy selector failed, err: %v", err)
	}

	if len(raw) > MaxScopeSelectorByteSize {
		return ErrSelectorByteSizeIsOverMaxLimit
	}

	return nil
}

// PublishType defines an app's strategy publish type.
type PublishType string

const (
	// Manually means this strategy publish type is manual.
	Manually PublishType = "Manually"
	// Automatically means this strategy publish type is automatical.
	Automatically PublishType = "Automatically"
	// Periodically means this strategy publish type is periodical.
	Periodically PublishType = "Periodically"
	// Immediately means this strategy publish type is immediate.
	Immediately PublishType = "Immediately"
)

// ValidatePublishType validate publish type
func (p PublishType) ValidatePublishType() error {
	switch p {
	case Manually:
	case Automatically:
	case Periodically:
	case Immediately:
	default:
		return fmt.Errorf("unsupported publish type: %s", p)
	}
	return nil
}

// PublishStatus defines an app's strategy publish status.
type PublishStatus string

const (
	// PendApproval means this strategy publish status is pending.
	PendApproval PublishStatus = "PendApproval"
	// PendPublish means this strategy publish status is pending.
	PendPublish PublishStatus = "PendPublish"
	// RevokedPublish means this strategy publish status is revoked.
	RevokedPublish PublishStatus = "RevokedPublish"
	// RejectedApproval means this strategy publish status is rejected.
	RejectedApproval PublishStatus = "RejectedApproval"
	// AlreadyPublish means this strategy publish status is already publish.
	AlreadyPublish PublishStatus = "AlreadyPublish"
)

// ValidatePublishStatus validate publish status
func (p PublishStatus) ValidatePublishStatus() error {
	switch p {
	case PendApproval:
	case PendPublish:
	case RevokedPublish:
	case RejectedApproval:
	case AlreadyPublish:
	default:
		return fmt.Errorf("unsupported publish status: %s", p)
	}
	return nil
}
