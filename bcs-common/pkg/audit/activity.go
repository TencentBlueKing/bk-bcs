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

// Package audit xxx
package audit

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/klog"
)

// ActivityStatus is the activity status
type ActivityStatus string

// ActivityType is the activity type
type ActivityType string

// ResourceType is the resource type
type ResourceType string

const (
	// ActivityStatusUnknow means the activity status is unknow
	ActivityStatusUnknow ActivityStatus = "unknow"
	// ActivityStatusSuccess means the activity is success
	ActivityStatusSuccess ActivityStatus = "success"
	// ActivityStatusFailed means the activity is failed
	ActivityStatusFailed ActivityStatus = "failed"
	// ActivityStatusPending means the activity is pending
	ActivityStatusPending ActivityStatus = "pending"

	// ActivityTypeView means the activity type is view
	ActivityTypeView ActivityType = "view"
	// ActivityTypeCreate means the activity type is create
	ActivityTypeCreate ActivityType = "create"
	// ActivityTypeUpdate means the activity type is update
	ActivityTypeUpdate ActivityType = "update"
	// ActivityTypeDelete means the activity type is delete
	ActivityTypeDelete ActivityType = "delete"
	// ActivityTypeStart means the activity type is start
	ActivityTypeStart ActivityType = "start"
	// ActivityTypeStop means the activity type is stop
	ActivityTypeStop ActivityType = "stop"

	// ResourceTypeProject means the resource type is project
	ResourceTypeProject ResourceType = "project"
	// ResourceTypeCluster means the resource type is cluster
	ResourceTypeCluster ResourceType = "cluster"
	// ResourceTypeNode means the resource type is node
	ResourceTypeNode ResourceType = "node"
	// ResourceTypeNodeGroup means the resource type is node group
	ResourceTypeNodeGroup ResourceType = "node_group"
	// ResourceTypeCloudAccount means the resource type is cloud account
	ResourceTypeCloudAccount ResourceType = "cloud_account"
	// ResourceTypeNamespace means the resource type is namespace
	ResourceTypeNamespace ResourceType = "namespace"
	// ResourceTypeTemplateSet means the resource type is template set
	ResourceTypeTemplateSet ResourceType = "templateset"
	// ResourceTypeVariable means the resource type is variable
	ResourceTypeVariable ResourceType = "variable"
	// ResourceTypeK8SResource means the resource type is k8s resource
	ResourceTypeK8SResource ResourceType = "k8s_resource"
	// ResourceTypeHelm means the resource type is helm
	ResourceTypeHelm ResourceType = "helm"
	// ResourceTypeAddons means the resource type is addons
	ResourceTypeAddons ResourceType = "addons"
	// ResourceTypeChart means the resource type is chart
	ResourceTypeChart ResourceType = "chart"
	// ResourceTypeWebConsole means the resource type is web console
	ResourceTypeWebConsole ResourceType = "web_console"
	// ResourceTypeLogRule means the resource type is log rule
	ResourceTypeLogRule ResourceType = "log_rule"
	// ResourceTypeGitOps means the source type is gitops
	ResourceTypeGitOps ResourceType = "gitops"
)

// Activity is the struct of activity
type Activity struct {
	ProjectCode  string         `json:"project_code"`
	ResourceType ResourceType   `json:"resource_type"`
	ResourceName string         `json:"resource_name"`
	ResourceID   string         `json:"resource_id"`
	ActivityType ActivityType   `json:"activity_type"`
	Status       ActivityStatus `json:"status"`
	Username     string         `json:"username"`
	Description  string         `json:"description"`
	SourceIP     string         `json:"source_ip"`
	UserAgent    string         `json:"user_agent"`
	Extra        string         `json:"extra"`
}

// ActivityReq is the request of activity
type ActivityReq struct {
	Activities []Activity `json:"activities"`
}

// ErrorResponse is the error response for restful response
type ErrorResponse struct {
	Error Error `json:"error"`
}

// Error is the error response for restful response
type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var (
	activityChan       = make(chan Activity, 10000)
	activityChanClosed int32
	activityOnce       sync.Once

	bcsHost string
	token   string
)

func start() {
	activityOnce.Do(func() {
		go func() {
			// consumeActivity every 10 seconds
			for range time.Tick(10 * time.Second) {
				consumeActivity()
			}
		}()
	})
}

func stop() {
	close(activityChan)
	atomic.AddInt32(&activityChanClosed, int32(1))
	// consume remaining activity
	consumeActivity()
}

func consumeActivity() {
	activity := make([]Activity, 0)

	// push activity
	do := func() {
		batchPushActivity(activity)
		if len(activity) > 0 {
			klog.Infof("push activity success, total %d", len(activity))
			// reset activity
			activity = activity[:0]
		}
	}

	// if activityChan is closed or empty, push activity
	for {
		select {
		case a, ok := <-activityChan:
			if !ok {
				do()
				return
			}
			activity = append(activity, a)
		default:
			do()
			return
		}
	}
}

// pushActivity push activity to queue
func pushActivity(activity Activity) {
	if a := atomic.LoadInt32(&activityChanClosed); a > 0 {
		klog.Warningf("activity chan is closed, ignore activity")
		return
	}
	go func() {
		activityChan <- activity
	}()
}

// batchPushActivity splits the input slice of Activity into smaller slices of 100 elements each and creates a new
// goroutine for each slice to push the activities to the audit log
func batchPushActivity(activity []Activity) {
	activities := SplitSlice(activity, 100)
	for _, v := range activities {
		go func(data []Activity) {
			if err := createActivity(data); err != nil {
				klog.Errorf("push activity failed, %s", err.Error())
			}
		}(v)
	}
}

// createActivity call bcs-api to create activity
func createActivity(activity []Activity) error {
	body := ActivityReq{
		Activities: activity,
	}
	url := fmt.Sprintf("%s/bcsapi/v4/usermanager/v3/activity_logs", bcsHost)
	resp, err := GetClient().R().SetAuthToken(token).SetBody(body).Post(url)
	if err != nil {
		return err
	}

	requestID := resp.Header().Get("x-request-id")
	if resp.StatusCode() != 200 {
		var errorResponse ErrorResponse
		if err = json.Unmarshal(resp.Body(), &errorResponse); err != nil {
			return fmt.Errorf("unmarshal error response failed, %s", err.Error())
		}
		return fmt.Errorf("push activity failed, requestID: %s, code: %s, message: %s", requestID,
			errorResponse.Error.Code, errorResponse.Error.Message)
	}
	return nil
}
