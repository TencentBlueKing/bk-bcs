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

package table

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/criteria/validator"
	"bscp.io/pkg/runtime/selector"
)

// StrategyColumns defines Strategy's columns
var StrategyColumns = mergeColumns(StrategyColumnDescriptor)

// StrategyColumnDescriptor is Strategy's column descriptors.
var StrategyColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", StrategySpecColumnDescriptor),
	mergeColumnDescriptors("state", StrategyStateColumnDescriptor),
	mergeColumnDescriptors("attachment", StrategyAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

const (
	// maxNormalStrategiesLimitForApp defines the max limit of normal type strategy for an app for user to create.
	maxNormalStrategiesLimitForApp = 5
	// maxNormalStrategiesLimitForApp defines the max limit of namespace type strategy for an app for user to create.
	maxNamespaceStrategiesLimitForApp = 200
)

// ValidateAppStrategyNumber verify whether the current number of app strategies have reached the maximum.
func ValidateAppStrategyNumber(count uint32, mode AppMode) error {
	switch mode {
	case Normal:
		if count >= maxNormalStrategiesLimitForApp {
			return errf.New(errf.InvalidParameter, fmt.Sprintf("an application only create %d normal strategies",
				maxNormalStrategiesLimitForApp))
		}
	case Namespace:
		if count >= maxNamespaceStrategiesLimitForApp {
			return errf.New(errf.InvalidParameter, fmt.Sprintf("an application only create %d namespace strategies",
				maxNamespaceStrategiesLimitForApp))
		}
	default:
		return fmt.Errorf("unsupported strategy set mode: %s", mode)
	}
	return nil
}

// Strategy defines a strategy for an app to publish.
// it contains the basic released information and the
// selector to define the scope of the matched instances.
type Strategy struct {
	// ID is an auto-increased value, which is a unique identity
	// of a strategy.
	ID         uint32              `db:"id" json:"id"`
	Spec       *StrategySpec       `db:"spec" json:"spec"`
	State      *StrategyState      `db:"state" json:"state"`
	Attachment *StrategyAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision           `db:"revision" json:"revision"`
}

// TableName is the strategy's database table name.
func (s Strategy) TableName() Name {
	return StrategyTable
}

// ValidateCreate validate strategy is valid or not when create it.
func (s Strategy) ValidateCreate() error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
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
func (s Strategy) ValidateUpdate(asDefault bool, namespaced bool) error {

	if s.ID <= 0 {
		return errors.New("id should be set")
	}

	changed := false
	if s.Spec != nil {
		changed = true
		if err := s.Spec.ValidateUpdate(asDefault, namespaced); err != nil {
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
func (s Strategy) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("strategy id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// StrategySpecColumns defines StrategySpec's columns
var StrategySpecColumns = mergeColumns(StrategySpecColumnDescriptor)

// StrategySpecColumnDescriptor is StrategySpec's column descriptors.
var StrategySpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "release_id", NamedC: "release_id", Type: enumor.Numeric},
	{Column: "as_default", NamedC: "as_default", Type: enumor.Boolean},
	{Column: "scope", NamedC: "scope", Type: enumor.String},
	{Column: "mode", NamedC: "mode", Type: enumor.String},
	{Column: "namespace", NamedC: "namespace", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String}}

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
	Name      string `db:"name" json:"name"`
	ReleaseID uint32 `db:"release_id" json:"release_id"`
	// AsDefault(=true) describes this strategy works as a bottom strategy,
	// which means if an instance can not match the other strategies in
	// a strategy set, then this instance will use this strategy's
	// released configure version.
	// Note:
	// 1. if a strategy is set to as a bottom strategy, then this strategy
	// scope's selector must be 'matched all'.
	// 2. if the strategy works at namespace mode and set as default strategy
	// at the same time, its namespace must be the *reserved* namespace, as is
	// 'bscp_default_ns'
	AsDefault bool `db:"as_default" json:"as_default"`

	// Scope must be empty when this strategy is a default strategy.
	// Scope must not be empty when this strategy is not a default strategy.
	Scope *ScopeSelector `db:"scope" json:"scope"`

	// Mode defines what mode of this strategy works at, it is succeeded from
	// this strategy's app's mode.
	// it can not be updated once it is created.
	Mode AppMode `db:"mode" json:"mode"`

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
	Namespace string `db:"namespace" json:"namespace"`
	Memo      string `db:"memo" json:"memo"`
}

// ValidateCreate validate strategy spec when it is created.
func (s StrategySpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if s.ReleaseID <= 0 {
		return errors.New("invalid strategy release id")
	}

	if s.Scope == nil {
		return errors.New("this strategy's scope should be set")
	}

	if s.Scope.IsEmpty() {
		return errors.New("this strategy's scope is required")
	}

	if err := s.Mode.Validate(); err != nil {
		return err
	}

	switch s.Mode {
	case Normal:

		if err := s.Scope.ValidateCreate(s.AsDefault, false); err != nil {
			return fmt.Errorf("validate scope failed, err: %v", err)
		}

		if len(s.Namespace) != 0 {
			return errors.New("strategy set works at normal mode, namespace should be empty")
		}

	case Namespace:

		if s.AsDefault {
			// this strategy is a default strategy, then its namespace will be set by system
			// automatically. user can not set it.
			if len(s.Namespace) != 0 {
				return errors.New("strategy set works at namespace mode, but this is a default strategy, " +
					"namespace should be empty")
			}
		} else {
			// if the strategy set works at namespace mode, but not the default strategy,
			// then the strategy's namespace should be configured.
			if len(s.Namespace) == 0 {
				return errors.New("strategy set works at namespace mode, namespace is required, can not be empty")
			}

			if err := validator.ValidateNamespace(s.Namespace); err != nil {
				return err
			}
		}

		if err := s.Scope.ValidateCreate(s.AsDefault, true); err != nil {
			return fmt.Errorf("validate create scope failed, err: %v", err)
		}

		if strings.HasPrefix(strings.ToLower(s.Namespace), ReservedNamespacePrefix) {
			return fmt.Errorf("namespace prefixed with %s is the system reserved namespaces, can not be used",
				ReservedNamespacePrefix)
		}

	default:
		return fmt.Errorf("unsupported strategy set's type: %s", s.Mode)
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate strategy spec when it is updated.
func (s StrategySpec) ValidateUpdate(asDefault bool, namespaced bool) error {

	if len(s.Name) != 0 {
		if err := validator.ValidateName(s.Name); err != nil {
			return err
		}
	}

	if s.ReleaseID <= 0 {
		return errors.New("release id should be set")
	}

	if s.Scope != nil && !s.Scope.IsEmpty() {
		if err := s.Scope.ValidateUpdate(asDefault, namespaced); err != nil {
			return err
		}
	}

	if len(s.Mode) != 0 {
		return errors.New("strategy's mode can not be updated")
	}

	if len(s.Namespace) != 0 {
		return errors.New("namespace can not be updated")
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
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

// StrategyStateColumns defines StrategyState's columns
var StrategyStateColumns = mergeColumns(StrategyStateColumnDescriptor)

// StrategyStateColumnDescriptor is StrategyState's column descriptors.
var StrategyStateColumnDescriptor = ColumnDescriptors{{Column: "pub_state", NamedC: "pub_state", Type: enumor.String}}

// StrategyState defines the strategy's state
type StrategyState struct {
	PubState PublishState `db:"pub_state" json:"pub_state"`
}

// Validate whether strategy state is valid or not.
func (s StrategyState) Validate() error {
	if err := s.PubState.Validate(); err != nil {
		return err
	}

	return nil
}

// StrategyAttachmentColumns defines StrategyAttachment's columns
var StrategyAttachmentColumns = mergeColumns(StrategyAttachmentColumnDescriptor)

// StrategyAttachmentColumnDescriptor is StrategyAttachment's column descriptors.
var StrategyAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
	{Column: "strategy_set_id", NamedC: "strategy_set_id", Type: enumor.Numeric}}

// StrategyAttachment defines the strategy attachments.
type StrategyAttachment struct {
	BizID         uint32 `db:"biz_id" json:"biz_id"`
	AppID         uint32 `db:"app_id" json:"app_id"`
	StrategySetID uint32 `db:"strategy_set_id" json:"strategy_set_id"`
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

// ScopeSelector defines a strategy's working scope.
// 1. ScopeSelector is stored in mysql with json raw, and each selector,
// including Strategy selector and its sub strategy selector, has a max
// size limit, as is MaxScopeSelectorByteSize byte.
// 2. an instance's version info is matched with the following rules:
//   (1) check if this instance's uid or labels can match Selector at first,
//      if not matched, then skip this strategy. if matched, then do next rule.
//   (2) check if this strategy do not have the SubStrategy, then return with
//      this strategy's release version.
//      if this strategy do have a sub strategy, then check this instance's
//      uid or labels can match this sub strategy's selector or not. if not,
//      then return this strategy's release version. if yes, then return this
//      sub-strategy's release version.
type ScopeSelector struct {
	// Selector define the strategy's working scope. must be set.
	// Note:
	// 1. This Selector is required, must not be empty under any circumstances.
	// 2. If the strategy's parent strategy set is working at Namespace
	// strategy set type, then this selector must be
	// set with matched all policy, which means matched all the instances,
	// and SubStrategy can be chosen to be configured or not depends on need
	// under this circumstance.
	// 3. this Selector's value is decided by the strategy's AsDefault and Namespace
	// state, different combination of them have the different Selector rules.
	Selector *selector.Selector `db:"selector" json:"selector"`

	// SubStrategy is the sub-strategy for this strategy.
	// if the sub strategy is configured, then it's selector can not be set
	// to match all, and the selector is required.
	SubStrategy *SubStrategy `db:"sub_strategy" json:"sub_strategy"`
}

// Scan is used to decode raw message which is read from db into a structured
// ScopeSelector instance.
func (s *ScopeSelector) Scan(raw interface{}) error {
	if s == nil {
		return errors.New("scope selector is not initialized")
	}

	if raw == nil {
		return errors.New("raw is nil, can not be decoded")
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("decode into scope selector failed, err: %v", err)

		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return fmt.Errorf("decode into scope selector failed, err: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported scope selector raw type: %T", v)
	}
}

// Value encode the scope selector to a json raw, so that it can be stored to db with
// json raw.
func (s *ScopeSelector) Value() (driver.Value, error) {
	if s == nil {
		return nil, errors.New("scope selector is not initialized, can not be encoded")
	}

	return json.Marshal(s)
}

// IsEmpty test whether this scope selector is empty or not.
func (s ScopeSelector) IsEmpty() bool {
	if s.Selector == nil && s.SubStrategy == nil {
		return true
	}

	if s.Selector != nil {
		if !s.Selector.IsEmpty() {
			return false
		}
	}

	if s.SubStrategy == nil {
		return true
	}

	if !s.SubStrategy.IsEmpty() {
		return false
	}

	return true
}

// ValidateCreate validate strategy's selector when it is created.
func (s ScopeSelector) ValidateCreate(asDefault bool, namespaced bool) error {
	if s.Selector == nil {
		return errors.New("strategy's selector is empty")
	}

	if s.Selector.IsEmpty() {
		return errors.New("strategy's selector is not set")
	}

	if namespaced && !s.Selector.MatchAll {
		return errors.New("strategy works at 'namespace' mode, scope.selector should be set to match all")
	}

	if asDefault && !s.Selector.MatchAll {
		if !s.Selector.MatchAll {
			return errors.New("this is a default strategy, scope.selector should be set to match all")
		}

		// sub strategy should be empty
		if !(s.SubStrategy == nil || (s.SubStrategy != nil && s.SubStrategy.IsEmpty())) {
			return errors.New("strategy is the default strategy, its sub-strategy should not be configured")
		}
	}

	raw, err := json.Marshal(s.Selector)
	if err != nil {
		return fmt.Errorf("marshal strategy selector failed, err: %v", err)
	}

	if len(raw) > MaxScopeSelectorByteSize {
		return ErrSelectorByteSizeIsOverMaxLimit
	}

	if err := s.Selector.Validate(); err != nil {
		return fmt.Errorf("invalid strategy selector, err: %v", err)
	}

	if s.SubStrategy != nil {
		if err := s.SubStrategy.ValidateCreate(); err != nil {
			return fmt.Errorf("invalid strategy's sub-strategy, err: %v", err)
		}
	}

	return nil
}

// ValidateUpdate validate strategy's selector when it is updated.
func (s ScopeSelector) ValidateUpdate(asDefault bool, namespaced bool) error {
	if s.Selector == nil {
		return errors.New("strategy's selector is empty")
	}

	if s.Selector.IsEmpty() {
		return errors.New("strategy's selector is not set")
	}

	if namespaced && !s.Selector.MatchAll {
		return errors.New("strategy works at 'namespace' mode, scope.selector should be set to match all")
	}

	if asDefault {
		if !s.Selector.MatchAll {
			return errors.New("this is a default strategy, scope.selector should be set to match all")
		}

		// sub strategy should be empty
		if !(s.SubStrategy == nil || (s.SubStrategy != nil && s.SubStrategy.IsEmpty())) {
			return errors.New("strategy is the default strategy, its sub-strategy should not be configured")
		}
	}

	if err := s.Selector.Validate(); err != nil {
		return fmt.Errorf("invalid strategy selector, err: %v", err)
	}

	raw, err := json.Marshal(s.Selector)
	if err != nil {
		return fmt.Errorf("marshal strategy selector failed, err: %v", err)
	}

	if len(raw) > MaxScopeSelectorByteSize {
		return ErrSelectorByteSizeIsOverMaxLimit
	}

	if s.SubStrategy != nil {
		if err := s.SubStrategy.ValidateUpdate(); err != nil {
			return fmt.Errorf("invalid strategy's sub-strategy, err: %v", err)
		}
	}

	return nil
}

// SubStrategy is the sub-strategy of its parent strategy, it can not be
// used independently.
type SubStrategy struct {
	Spec *SubStrategySpec `db:"spec" json:"spec"`
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
func (s SubStrategy) ValidateCreate() error {
	if s.Spec == nil {
		return errors.New("sub strategy's spec is empty")
	}

	if err := s.Spec.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate sub strategy when it is updated.
func (s SubStrategy) ValidateUpdate() error {
	if s.Spec == nil {
		return errors.New("sub strategy's spec is empty")
	}

	if err := s.Spec.Validate(); err != nil {
		return err
	}

	return nil
}

// SubStrategySpec is the sub-strategy's specifics defined by user.
type SubStrategySpec struct {
	Name string `db:"name" json:"name"`
	// ReleaseID this sub strategy's released version id.
	ReleaseID uint32            `db:"release_id" json:"release_id"`
	Scope     *SubScopeSelector `db:"scope" json:"scope"`
	Memo      string            `db:"memo" json:"memo"`
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
func (s SubStrategySpec) Validate() error {
	if err := validator.ValidateName(s.Name); err != nil {
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

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
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
	Selector *selector.Selector `db:"selector" json:"selector"`
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

// CurrentPublishedStrategyColumns defines 's columns
var CurrentPublishedStrategyColumns = mergeColumns(CPSColumnDescriptor)

// CPSColumnDescriptor is CurrentPublishedStrategy's column descriptors.
var CPSColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "strategy_id", NamedC: "strategy_id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", StrategySpecColumnDescriptor),
	mergeColumnDescriptors("state", StrategyStateColumnDescriptor),
	mergeColumnDescriptors("attachment", StrategyAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", CreatedRevisionColumnDescriptor))

// CurrentPublishedStrategy defines the current published strategy details.
// It's a snapshot of the strategy details when do the publish operation.
// Note:
// 1. one strategy has no more than one record in this table.
// 2. when a strategy is published, if it already has a record in this table,
//    then the old record should be deleted and insert a new one with a new ID.
// 3. when the strategy is deleted, then it's published record in this table
//    should also be deleted at the same time.
// 4. it's not editable once created.
type CurrentPublishedStrategy struct {
	// ID is an auto-increased value, which is a unique identity
	// of the published strategy. it not represents the strategy id.
	ID uint32 `db:"id" json:"id"`
	// the published strategy's identity id. it's a unique index in this table.
	StrategyID uint32              `db:"strategy_id" json:"strategy_id"`
	Spec       *StrategySpec       `db:"spec" json:"spec"`
	State      *StrategyState      `db:"state" json:"state"`
	Attachment *StrategyAttachment `db:"attachment" json:"attachment"`
	Revision   *CreatedRevision    `db:"revision" json:"revision"`
}

// TableName is the current publish strategy's database table name.
func (c CurrentPublishedStrategy) TableName() Name {
	return CurrentPublishedStrategyTable
}

// Validate the current published strategy is valid or not.
func (c CurrentPublishedStrategy) Validate() error {
	if c.ID > 0 {
		return errors.New("id can not be set")
	}

	if c.StrategyID <= 0 {
		return errors.New("strategy id not set")
	}

	if c.Spec == nil {
		return errors.New("spec not set")
	}

	if err := c.Spec.ValidateCreate(); err != nil {
		return err
	}

	if c.State != nil {
		return errors.New("state not set")
	}

	if err := c.State.Validate(); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision == nil {
		return errors.New("revision not set")
	}

	if err := c.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// PubStrategyHistoryColumns defines PublishedStrategyHistory's columns
var PubStrategyHistoryColumns = mergeColumns(PSHColumnDescriptor)

// PSHColumnDescriptor is PublishedStrategyHistory's column descriptors.
var PSHColumnDescriptor = CPSColumnDescriptor

// PublishedStrategyHistory is used to store the published strategy.
// each publish strategy will be recorded here, which will be used
// to rollbacks the published release or check the published strategy
// histories.
type PublishedStrategyHistory CurrentPublishedStrategy

// TableName is the current publish strategy's database table name.
func (p PublishedStrategyHistory) TableName() Name {
	return PublishedStrategyHistoryTable
}
