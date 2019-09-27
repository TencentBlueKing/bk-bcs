package store

import (
	"bk-bcs/bcs-common/common/blog"
	"fmt"
	"strings"
)

func GetRunAsAndAppIDbyTaskGroupID(taskGroupId string) (string, string) {
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
}
