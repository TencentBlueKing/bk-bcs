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

package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	// PendingStatus rule pending status
	PendingStatus = "PENDING"
	// RunningStatus rule running status
	RunningStatus = "RUNNING"
	// FailedStatus rule fail status
	FailedStatus = "FAILED"
	// SuccessStatus rule success status
	SuccessStatus = "SUCCESS"
	// TerminatedStatus rule terminated status
	TerminatedStatus = "TERMINATED"
	// DeletedStatus rule deleted status
	DeletedStatus = "DELETED"

	// create or update rule timeout
	defaultTimeout = time.Minute * 10
)

// LogRule log rule
type LogRule struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	Name               string             `json:"name" bson:"name"`
	RuleName           string             `json:"rule_name" bson:"ruleName"`
	RuleID             int                `json:"rule_id" bson:"ruleID"`
	Description        string             `json:"description" bson:"description"`
	ProjectID          string             `json:"project_id" bson:"projectID"`
	ProjectCode        string             `json:"project_code" bson:"projectCode"`
	ClusterID          string             `json:"cluster_id" bson:"clusterID"`
	Rule               bklog.LogRule      `json:"rule" bson:"rule"`
	FileIndexSetID     int                `json:"file_index_set_id" bson:"fileIndexSetID"`
	STDIndexSetID      int                `json:"std_index_set_id" bson:"stdIndexSetID"`
	RuleFileIndexSetID int                `json:"rule_file_index_set_id" bson:"ruleFileIndexSetID"`
	RuleSTDIndexSetID  int                `json:"rule_std_index_set_id" bson:"ruleSTDIndexSetID"`
	CreatedAt          utils.JSONTime     `json:"created_at" bson:"createdAt"`
	UpdatedAt          utils.JSONTime     `json:"updated_at" bson:"updatedAt"`
	Creator            string             `json:"creator" bson:"creator"`
	Updator            string             `json:"updator" bson:"updator"`
	Status             string             `json:"status" bson:"status"`
	Message            string             `json:"message" bson:"message"`
	FromRule           string             `json:"from_rule" bson:"fromRule"`
}

// FixStatus fix log rule status when status is stucked
func (r *LogRule) FixStatus() {
	if r.Status == PendingStatus && r.UpdatedAt.Before(time.Now().Add(-defaultTimeout)) {
		r.Status = FailedStatus
		r.Message = "apply log rule timeout"
	}
}
