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

package common

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

// GetBCSProjectID get projectID from annotations
func GetBCSProjectID(data map[string]string) string {
	// projectID is required in Authentication(hash id, not readable).
	// so ProjectController store projectID in AppProject.Meta,
	// indexer is bkbcs.tencent.com/projectID
	if data == nil {
		return ""
	}
	projectID, ok := data[ProjectIDKey]
	if !ok {
		return ""
	}
	return projectID
}

// GetBCSProjectBusinessKey return the business id of project
func GetBCSProjectBusinessKey(data map[string]string) string {
	if data == nil {
		return ""
	}
	businessID, ok := data[ProjectBusinessIDKey]
	if !ok {
		return ""
	}
	return businessID
}

// AddCustomAnnotationForApplication add custom annotation for application
func AddCustomAnnotationForApplication(argoProj *v1alpha1.AppProject, app *v1alpha1.Application) {
	app.Annotations[ProjectIDKey] = GetBCSProjectID(argoProj.Annotations)
	app.Annotations[ProjectBusinessIDKey] = GetBCSProjectBusinessKey(argoProj.Annotations)
}
