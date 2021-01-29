/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package testdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
)

func randName(prefix string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, time.Now().Format("2006-01-02-15:04:05"), time.Now().Nanosecond())
}

// CreateAppTestData returns test data for create application case.
func CreateAppTestData(bizID string) (string, error) {
	req := &pb.CreateAppReq{
		BizId:      bizID,
		Name:       randName("e2e-app"),
		DeployType: 0,
		Memo:       "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// QueryAppListTestData returns test data for list application case.
func QueryAppListTestData(bizID string, returnTotal bool, start, limit int32) (string, error) {
	req := &pb.QueryAppListReq{
		BizId: bizID,
		Page: &pbcommon.Page{
			ReturnTotal: returnTotal,
			Start:       start,
			Limit:       limit,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateAppTestData returns test data for update application case.
func UpdateAppTestData(bizID, appID string) (string, error) {
	req := &pb.UpdateAppReq{
		BizId:      bizID,
		AppId:      appID,
		Name:       randName("e2e-app"),
		DeployType: 0,
		Memo:       "e2e testing update",
		State:      0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateConfigTestData returns test data for create config case.
func CreateConfigTestData(bizID, appID, path string) (string, error) {
	req := &pb.CreateConfigReq{
		BizId: bizID,
		AppId: appID,
		Name:  randName("e2e-config"),
		Fpath: path,
		Memo:  "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// QueryConfigListTestData returns test data for list config case.
func QueryConfigListTestData(bizID, appID string, returnTotal bool, start, limit int32) (string, error) {
	req := &pb.QueryConfigListReq{
		BizId: bizID,
		AppId: appID,
		Page: &pbcommon.Page{
			ReturnTotal: returnTotal,
			Start:       start,
			Limit:       limit,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateConfigTestData returns test data for update config case.
func UpdateConfigTestData(bizID, appID, cfgID string) (string, error) {
	req := &pb.UpdateConfigReq{
		BizId: bizID,
		AppId: appID,
		CfgId: cfgID,
		Name:  randName("e2e-config"),
		Fpath: "/etc",
		Memo:  "e2e testing update",
		State: 0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateCommitTestData returns test data for create commit case.
func CreateCommitTestData(bizID, appID, cfgID string, commitMode pbcommon.CommitMode) (string, error) {
	req := &pb.CreateCommitReq{
		BizId:      bizID,
		AppId:      appID,
		CfgId:      cfgID,
		CommitMode: int32(commitMode),
		Memo:       "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// QueryHistoryCommitsTestData returns test data for query history commits case.
func QueryHistoryCommitsTestData(bizID, appID, cfgID string, returnTotal bool, start, limit int32) (string, error) {
	req := &pb.QueryHistoryCommitsReq{
		BizId: bizID,
		AppId: appID,
		CfgId: cfgID,
		Page: &pbcommon.Page{
			ReturnTotal: returnTotal,
			Start:       start,
			Limit:       limit,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateCommitTestData returns test data for update commit case.
func UpdateCommitTestData(bizID, appID, commitID string, commitMode pbcommon.CommitMode) (string, error) {
	req := &pb.UpdateCommitReq{
		BizId:      bizID,
		AppId:      appID,
		CommitId:   commitID,
		CommitMode: int32(commitMode),
		Memo:       "e2e testing update",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// ConfirmCommitTestData returns test data for confirm commit case.
func ConfirmCommitTestData(bizID, appID, commitID string) (string, error) {
	req := &pb.ConfirmCommitReq{
		BizId:    bizID,
		AppId:    appID,
		CommitId: commitID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CancelCommitTestData returns test data for cancel commit case.
func CancelCommitTestData(bizID, appID, commitID string) (string, error) {
	req := &pb.CancelCommitReq{
		BizId:    bizID,
		AppId:    appID,
		CommitId: commitID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateReleaseTestData returns test data for create release case.
func CreateReleaseTestData(bizID, appID, commitID, strategyID string) (string, error) {
	req := &pb.CreateReleaseReq{
		BizId:      bizID,
		AppId:      appID,
		CommitId:   commitID,
		Name:       randName("e2e-release"),
		StrategyId: strategyID,
		Memo:       "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// QueryHistoryReleasesTestData returns test data for query history releases case.
func QueryHistoryReleasesTestData(bizID, appID, cfgID string, returnTotal bool, start, limit int32) (string, error) {
	req := &pb.QueryHistoryReleasesReq{
		BizId: bizID,
		AppId: appID,
		CfgId: cfgID,
		Page: &pbcommon.Page{
			ReturnTotal: returnTotal,
			Start:       start,
			Limit:       limit,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateReleaseTestData returns test data for update release case.
func UpdateReleaseTestData(bizID, appID, releaseID string) (string, error) {
	req := &pb.UpdateReleaseReq{
		BizId:     bizID,
		AppId:     appID,
		ReleaseId: releaseID,
		Name:      randName("e2e-release"),
		Memo:      "e2e testing update",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// PublishReleaseTestData returns test data for publish release case.
func PublishReleaseTestData(bizID, appID, releaseID string) (string, error) {
	req := &pb.PublishReleaseReq{
		BizId:     bizID,
		AppId:     appID,
		ReleaseId: releaseID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CancelReleaseTestData returns test data for cancel release case.
func CancelReleaseTestData(bizID, appID, releaseID string) (string, error) {
	req := &pb.CancelReleaseReq{
		BizId:     bizID,
		AppId:     appID,
		ReleaseId: releaseID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateStrategyTestData returns test data for create strategy case.
func CreateStrategyTestData(bizID, appID string) (string, error) {
	req := &pb.CreateStrategyReq{
		BizId:     bizID,
		AppId:     appID,
		Name:      randName("e2e-strategy"),
		LabelsOr:  []*pbcommon.LabelsMap{},
		LabelsAnd: []*pbcommon.LabelsMap{},
		Memo:      "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// QueryStrategyListTestData returns test data for query strategy list case.
func QueryStrategyListTestData(bizID, appID string, returnTotal bool, start, limit int32) (string, error) {
	req := &pb.QueryStrategyListReq{
		BizId: bizID,
		AppId: appID,
		Page: &pbcommon.Page{
			ReturnTotal: returnTotal,
			Start:       start,
			Limit:       limit,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}
