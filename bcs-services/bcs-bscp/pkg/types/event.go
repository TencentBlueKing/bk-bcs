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

package types

import (
	"errors"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// Event defines a event's details info.
type Event struct {
	Spec       *table.EventSpec       `json:"spec"`
	Attachment *table.EventAttachment `json:"attachment"`
	Revision   *table.CreatedRevision `json:"revision"`
}

// Validate an event is valid or not.
func (e Event) Validate() error {
	if e.Spec == nil {
		return errors.New("invalid event spec, is nil")
	}

	if err := e.Spec.Validate(); err != nil {
		return err
	}

	if e.Attachment == nil {
		return errors.New("invalid event attachment, is nil")
	}

	if err := e.Attachment.Validate(); err != nil {
		return err
	}

	if e.Revision == nil {
		return errors.New("invalid event revision, is nil")
	}

	if err := e.Revision.Validate(); err != nil {
		return err
	}

	return nil
}
