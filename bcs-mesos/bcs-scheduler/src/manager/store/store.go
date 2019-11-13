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
 *
 */

package store

import (
	"errors"
)

var (
	ErrNoFound = errors.New("store: Not Found")
)

/*func GetRunAsAndAppIDbyTaskGroupID(taskGroupId string) (string, string) {
	appID := ""
	runAs := ""

	szSplit := strings.Split(taskGroupId, ".")
	//RunAs
	if len(szSplit) >= 3 {
		runAs = szSplit[2]
	}

	//appID
	if len(szSplit) >= 2 {
		appID = szSplit[1]
	}

	return runAs, appID
}

func GetRunAsAndAppIDbyTaskID(taskId string) (string, string) {
	appID := ""
	runAs := ""

	szSplit := strings.Split(taskId, ".")
	//RunAs
	if len(szSplit) >= 6 {
		runAs = szSplit[4]
	}

	//appID
	if len(szSplit) >= 6 {
		appID = szSplit[3]
	}

	return runAs, appID
}

func GetTaskGroupID(taskID string) string {

	splitID := strings.Split(taskID, ".")
	if len(splitID) < 6 {
		blog.Error("TaskID %s format error", taskID)
		return ""
	}
	// appInstances, appID, appRunAs, appClusterId, idTime
	taskGroupID := fmt.Sprintf("%s.%s.%s.%s.%s", splitID[2], splitID[3], splitID[4], splitID[5], splitID[0])

	return taskGroupID
}*/
