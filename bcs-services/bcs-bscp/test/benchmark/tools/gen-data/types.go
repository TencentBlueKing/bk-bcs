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

package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

const (
	memo             = "stress-test"
	stressBizId      = 11
	stressInstanceID = "961b6dd3ede3cb8ecbaacbd68de040cd78eb2ed5889130cceb4c49268ea4d506"
	namespacePrefix  = "namespace"
)

// RequestID generate request id for stress test.
func RequestID() string {
	return uuid.UUID()
}

// Header generate request header for api client.
func Header(rid string) http.Header {
	header := http.Header{}
	header.Set(constant.UserKey, constant.BKUserForTestPrefix+"stress")
	header.Set(constant.RidKey, rid)
	header.Set(constant.AppCodeKey, "test")
	header.Add("Cookie", "bk_token="+constant.BKTokenForTest)

	return header
}

// randName generate rand resource name.
func randName(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("2006-01-02-15_04_05"), uuid.UUID())
}

// strategy scope selector rules define, used to stress.
var (
	elements = []selector.Element{element1, element2, element3, element4}
	element1 = selector.Element{Key: "biz", Op: new(selector.EqualType), Value: "2001"}
	element2 = selector.Element{Key: "set", Op: new(selector.InType), Value: []string{"1", "2", "3"}}
	element3 = selector.Element{Key: "module", Op: new(selector.GreaterThanType), Value: 1}
	element4 = selector.Element{Key: "game", Op: new(selector.NotEqualType), Value: "stress"}
	element5 = selector.Element{Key: "sub", Op: new(selector.EqualType), Value: "true"}
)

func genSelector(els []selector.Element) (*pbstruct.Struct, error) {
	if len(els) == 0 {
		return nil, errors.New("element rule is required")
	}

	ft := selector.Selector{
		MatchAll: false,
		LabelsOr: els,
	}

	return ft.MarshalPB()
}
