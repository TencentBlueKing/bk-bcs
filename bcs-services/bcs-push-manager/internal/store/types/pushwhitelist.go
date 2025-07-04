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

// Package types defines the data structures for database models.
package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionPushWhitelist is the name of the push_whitelists collection in MongoDB.
const CollectionPushWhitelist = "push_whitelists"

// PushWhitelist represents a push whitelist record in the database.
type PushWhitelist struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WhitelistID     string             `bson:"whitelist_id" json:"whitelist_id"`
	Domain          string             `bson:"domain" json:"domain"`
	Dimension       Dimension          `bson:"dimension" json:"dimension"`
	Reason          string             `bson:"reason" json:"reason"`
	Applicant       string             `bson:"applicant" json:"applicant"`
	Approver        string             `bson:"approver" json:"approver"`
	WhitelistStatus *int               `bson:"whitelist_status" json:"whitelist_status"`
	ApprovalStatus  *int               `bson:"approval_status" json:"approval_status"`
	StartTime       time.Time          `bson:"start_time" json:"start_time"`
	EndTime         *time.Time         `bson:"end_time" json:"end_time"`
	ApprovedAt      *time.Time         `bson:"approved_at" json:"approved_at"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time         `bson:"deleted_at" json:"deleted_at"`
}
