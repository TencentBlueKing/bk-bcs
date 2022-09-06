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

package common

import (
	"encoding/json"
	"testing"
)

func TestVersionAPIError(t *testing.T) {
	responseRaw := []byte("{\"Response\":{\"Error\":{\"Code\":\"InternalError\",\"Message\":\"An internal error has occurred. Retry your request, but if the problem persists, contact us with details by posting a message on the Tencent cloud forums.\"},\"RequestId\":\"request-id-mock\"}}")

	versionErrorResponse := VersionAPIError{}

	err := json.Unmarshal(responseRaw, &versionErrorResponse)
	if err != nil {
		t.Fatal(err)
	}

	if (versionErrorResponse.Response.Error.Code != "") != true {
		t.Fatal("unable to detect versioned api error.")
	}
}
