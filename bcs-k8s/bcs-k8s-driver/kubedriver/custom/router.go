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

package custom

import (
	"regexp"
	"strings"
)

type APIUrl string

// key can't be same with K8S URI
// Warning: It can not be start with "api" or "apis"
var UrlHandlerMap = map[string]APIHandler{
	"cluster/resources": &ClusterResourceAPIHandler{},
	"bcsclient/.+":      &BcsClientAPIHandler{},
}

type APIRouterInterface interface {
	Route(subPath string) *APIHandler
}

type APIRouter struct {
	UrlHandlerMap map[string]APIHandler
}

func NewRouter() (ar *APIRouter) {
	ar = &APIRouter{}
	ar.InitRegisteredUrls()
	return
}

func (ar *APIRouter) InitRegisteredUrls() {
	ar.UrlHandlerMap = UrlHandlerMap
}

func (ar *APIRouter) Route(subPath string) (handler APIHandler) {
	subPathWithoutQuery := strings.Split(subPath, "?")[0]
	for url, handler := range ar.UrlHandlerMap {
		if m, _ := regexp.MatchString(url, subPathWithoutQuery); m {
			return handler
		}
	}
	return nil
}
