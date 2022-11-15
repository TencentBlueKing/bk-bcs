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
	"errors"
	"fmt"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/criteria/validator"
)

// CurrentReleasedInstanceColumns defines CurrentReleasedInstance's columns
var CurrentReleasedInstanceColumns = mergeColumns(CRInstanceColumnDescriptor)

// CRInstanceColumnDescriptor is CurrentReleasedInstance's column descriptors.
var CRInstanceColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", RISpecColumnDescriptor),
	mergeColumnDescriptors("attachment", ReleaseAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", CreatedRevisionColumnDescriptor))

// maxReleasedInstancesLimitForApp defines the max limit of instance for an app for the user to configure.
const maxReleasedInstancesLimitForApp = 200

// ValidateAppCRINumber verify whether the current number of app current released instance has reached the maximum.
func ValidateAppCRINumber(count uint32) error {
	if count >= maxReleasedInstancesLimitForApp {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("an application only create %d current "+
			"released instance", maxReleasedInstancesLimitForApp))
	}
	return nil
}

// CurrentReleasedInstance defines a released version for a unique app's
// instance. each app can have less than MaxReleasedInstancesLimitForApp.
type CurrentReleasedInstance struct {
	// ID is an auto-increased value, which is a unique identity
	// of a released instance.
	ID         uint32                `db:"id" json:"id"`
	Spec       *ReleasedInstanceSpec `db:"spec" json:"spec"`
	Attachment *ReleaseAttachment    `db:"attachment" json:"attachment"`
	Revision   *CreatedRevision      `db:"revision" json:"revision"`
}

// TableName is the current publish strategy's database table name.
func (c CurrentReleasedInstance) TableName() Name {
	return CurrentReleasedInstanceTable
}

// ValidateCreate the current released instance details when create it.
func (c CurrentReleasedInstance) ValidateCreate() error {
	if c.ID > 0 {
		return errors.New("id should not be set")
	}

	if c.Spec == nil {
		return errors.New("spec is not set")
	}

	if err := c.Spec.Validate(); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment is not set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision == nil {
		return errors.New("revision is not set")
	}

	if err := c.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate the current released instance details when update it.
func (c CurrentReleasedInstance) ValidateUpdate() error {
	if c.ID <= 0 {
		return errors.New("id is not set")
	}

	if c.Spec == nil {
		return errors.New("spec is not set")
	}

	if err := c.Spec.Validate(); err != nil {
		return err
	}

	if c.Attachment != nil {
		if !c.Attachment.IsEmpty() {
			return errors.New("attachment can not be updated")
		}
	}

	if c.Revision == nil {
		return errors.New("revision is not set")
	}

	if err := c.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the current released instance's info when delete it.
func (c CurrentReleasedInstance) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("current released instance id should be set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if c.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	return nil
}

// RISpecColumns defines ReleasedInstanceSpec's columns
var RISpecColumns = mergeColumns(RISpecColumnDescriptor)

// RISpecColumnDescriptor is ReleasedInstanceSpec's column descriptors.
var RISpecColumnDescriptor = ColumnDescriptors{
	{Column: "uid", NamedC: "uid", Type: enumor.String},
	{Column: "release_id", NamedC: "release_id", Type: enumor.Numeric},
	{Column: "memo", NamedC: "memo", Type: enumor.String}}

// ReleasedInstanceSpec defines a released instance's specifics.
type ReleasedInstanceSpec struct {
	// Uid is the unique id of an app's instance identity.
	Uid       string `db:"uid" json:"uid"`
	ReleaseID uint32 `db:"release_id" json:"release_id"`
	Memo      string `db:"memo" json:"memo"`
}

// Validate this released instance's specifics
func (r ReleasedInstanceSpec) Validate() error {

	if err := validator.ValidateUid(r.Uid); err != nil {
		return err
	}

	if r.ReleaseID <= 0 {
		return errors.New("invalid release id")
	}

	if err := validator.ValidateMemo(r.Memo, false); err != nil {
		return err
	}

	return nil
}
