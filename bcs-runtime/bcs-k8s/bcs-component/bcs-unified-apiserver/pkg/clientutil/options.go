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
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
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

// MakeCreateOptions 组装 Create 参数
func MakeCreateOptions(q url.Values) (*metav1.CreateOptions, error) {
	opts := &metav1.CreateOptions{}
	fieldManager := q.Get("fieldManager")
	if fieldManager != "" {
		opts.FieldManager = fieldManager
	}
	return opts, nil
}

// MakeCreateOpMakePatchOptionstions 组装 Patch 参数
func MakePatchOptions(q url.Values) (*metav1.PatchOptions, error) {
	opts := &metav1.PatchOptions{}
	fieldManager := q.Get("fieldManager")
	if fieldManager != "" {
		opts.FieldManager = fieldManager
	}
	return opts, nil
}

// GetDeleteOptionsFromReq 从查询参数获取 DeleteOptions
func GetDeleteOptionsFromReq(req *http.Request) (*metav1.DeleteOptions, error) {
	// 优先从 body 获取 DeleteOptions
	if req.ContentLength > 0 {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		deleteOptions := metav1.DeleteOptions{}
		err = json.Unmarshal(body, &deleteOptions)
		if err != nil {
			return nil, err
		}
		return &deleteOptions, nil
	}

	// body 为空时在从查询参数获取
	deleteOptions := &metav1.DeleteOptions{
		OrphanDependents: nil, // 1.7 版本后废弃，不再支持
		DryRun:           []string{},
	}

	q := req.URL.Query()
	gracePeriodSecondsStr := q.Get("gracePeriodSeconds")
	if gracePeriodSecondsStr != "" {
		t, err := strconv.ParseInt(gracePeriodSecondsStr, 10, 64)
		if err != nil {
			return nil, errors.Errorf("cannot parse 'gracePeriodSeconds', err: %s", err)
		}
		deleteOptions.GracePeriodSeconds = &t
	}

	dryRunStr := q.Get("dryRun")
	if dryRunStr != "" {
		deleteOptions.DryRun = append(deleteOptions.DryRun, dryRunStr)
	}

	propagationPolicyStr := q.Get("propagationPolicy")
	if propagationPolicyStr != "" {
		tmpPP := metav1.DeletionPropagation(propagationPolicyStr)
		deleteOptions.PropagationPolicy = &tmpPP
	}

	return deleteOptions, nil
}

// MakePodLogOptions 从查询参数获取 PodLogOptions
func MakePodLogOptions(q url.Values) (*v1.PodLogOptions, error) {
	options := &v1.PodLogOptions{
		Container: q.Get("container"),
	}

	if queryTail := q.Get("tailLines"); queryTail != "" {
		t, err := strconv.ParseInt(queryTail, 10, 64)
		if err != nil {
			return nil, errors.New("cannot parse 'tailLines'")
		}
		options.TailLines = &t
	}

	if limitBytesStr := q.Get("limitBytes"); limitBytesStr != "" {
		t, err := strconv.ParseInt(limitBytesStr, 10, 64)
		if err != nil {
			return nil, errors.New("cannot parse 'limitBytes'")
		}
		options.LimitBytes = &t
	}

	if sinceSecondsStr := q.Get("sinceSeconds"); sinceSecondsStr != "" {
		t, err := strconv.ParseInt(sinceSecondsStr, 10, 64)
		if err != nil {
			return nil, errors.New("cannot parse 'sinceSeconds'")
		}
		options.SinceSeconds = &t
	}

	if followStr := strings.ToLower(q.Get("follow")); followStr != "" {
		follow, err := strconv.ParseBool(followStr)
		if err != nil {
			return nil, errors.New("cannot parse 'follow'")
		}
		options.Follow = follow
	}

	if InsecureSkipTLSVerifyBackendStr := strings.ToLower(q.Get("insecureSkipTLSVerifyBackend")); InsecureSkipTLSVerifyBackendStr != "" {
		InsecureSkipTLSVerifyBackend, err := strconv.ParseBool(InsecureSkipTLSVerifyBackendStr)
		if err != nil {
			return nil, errors.New("cannot parse 'insecureSkipTLSVerifyBackend'")
		}
		options.InsecureSkipTLSVerifyBackend = InsecureSkipTLSVerifyBackend
	}

	if previousStr := strings.ToLower(q.Get("previous")); previousStr != "" {
		previous, err := strconv.ParseBool(previousStr)
		if err != nil {
			return nil, errors.New("cannot parse 'previous'")
		}
		options.Previous = previous
	}

	if timestampsStr := strings.ToLower(q.Get("timestamps")); timestampsStr != "" {
		timestamps, err := strconv.ParseBool(timestampsStr)
		if err != nil {
			return nil, errors.New("cannot parse 'previous'")
		}
		options.Timestamps = timestamps
	}

	return options, nil
}
