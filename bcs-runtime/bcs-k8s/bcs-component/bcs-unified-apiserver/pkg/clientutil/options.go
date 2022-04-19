/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clientutil

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getListOptionsFromQueryParam 从查询参数获取 ListOptions
func GetListOptionsFromQueryParam(q url.Values) (*metav1.ListOptions, error) {
	var errReturn error = nil
	allowWatchBookmarksStr := strings.ToLower(q.Get("allowWatchBookmarks"))
	allowWatchBookmarksBool := allowWatchBookmarksStr == "true" || allowWatchBookmarksStr == "yes"
	continueStr := q.Get("continue")
	fieldSelector := q.Get("fieldSelector")
	labelSelector := q.Get("labelSelector")

	limitStr := q.Get("limit")
	var limitInt64 *int64 = nil
	if limitStr != "" {
		t, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			errReturn = errors.New("cannot parse 'limit' param")
			return nil, errReturn
		}
		limitInt64 = &t
	}

	timeoutSecondsStr := q.Get("timeoutSeconds")
	var timeoutSecondsInt64 *int64 = nil
	if timeoutSecondsStr != "" {
		t, err := strconv.ParseInt(timeoutSecondsStr, 10, 64)
		if err != nil {
			errReturn = errors.New("cannot parse 'timeoutSeconds'")
			return nil, errReturn
		}
		timeoutSecondsInt64 = &t
	}
	resourceVersion := q.Get("resourceVersion")

	watchStr := strings.ToLower(q.Get("watch"))
	watchBool := watchStr == "true" || watchStr == "yes"

	listOptions := metav1.ListOptions{
		AllowWatchBookmarks: allowWatchBookmarksBool,
		Continue:            continueStr,
		FieldSelector:       fieldSelector,
		LabelSelector:       labelSelector,
		ResourceVersion:     resourceVersion,
		TimeoutSeconds:      timeoutSecondsInt64,
		Watch:               watchBool,
	}
	if limitInt64 != nil {
		listOptions.Limit = *limitInt64
	}
	return &listOptions, errReturn
}

// GetDeleteOptionsFromReq 从查询参数获取 DeleteOptions
func GetDeleteOptionsFromReq(req *http.Request) (metav1.DeleteOptions, error) {
	var errReturn error = nil
	deleteOptions := metav1.DeleteOptions{}

	// 优先从 body 获取 DeleteOptions
	if req.ContentLength > 0 {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return deleteOptions, err
		}
		err = json.Unmarshal(body, &deleteOptions)
		if err != nil {
			return deleteOptions, err
		}
		return deleteOptions, nil
	}

	// body 为空时在从查询参数获取
	q := req.URL.Query()
	gracePeriodSecondsStr := q.Get("gracePeriodSeconds")
	var gracePeriodSecondsInt64 *int64
	if gracePeriodSecondsStr != "" {
		t, err := strconv.ParseInt(gracePeriodSecondsStr, 10, 64)
		if err != nil {
			errReturn = errors.New("cannot parse 'gracePeriodSeconds'")
			return metav1.DeleteOptions{}, errReturn
		}
		gracePeriodSecondsInt64 = &t
	} else {
		gracePeriodSecondsInt64 = nil
	}

	var dryRun []string
	dryRunStr := q.Get("dryRun")
	if dryRunStr == "" {
		dryRun = nil
	} else {
		dryRun = append(dryRun, dryRunStr)
	}

	var propagationPolicy *metav1.DeletionPropagation
	propagationPolicyStr := q.Get("propagationPolicy")
	if propagationPolicyStr == "" {
		propagationPolicy = nil
	} else {
		tmpPP := metav1.DeletionPropagation(propagationPolicyStr)
		propagationPolicy = &tmpPP
	}

	deleteOptions = metav1.DeleteOptions{
		DryRun:             dryRun,
		GracePeriodSeconds: gracePeriodSecondsInt64,
		OrphanDependents:   nil, // 1.7 版本后废弃，不再支持
		PropagationPolicy:  propagationPolicy,
	}
	return deleteOptions, errReturn
}
