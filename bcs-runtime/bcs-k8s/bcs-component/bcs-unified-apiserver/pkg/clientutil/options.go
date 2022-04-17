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
	"errors"
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
