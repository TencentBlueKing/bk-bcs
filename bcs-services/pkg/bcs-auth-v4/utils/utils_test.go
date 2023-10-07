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

package utils

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCalcIAMNsID(t *testing.T) {
	cases := []struct {
		Name      string
		ClusterID string
		Namespace string
		Want      string
	}{
		{
			Name:      "#1",
			ClusterID: "BCS-K8S-40000",
			Namespace: "default",
			Want:      "40000:5f03d33dde",
		},
		{
			Name:      "#2",
			ClusterID: "BCS-K8S-40000",
			Namespace: "de",
			Want:      "40000:9301fd7bde",
		},
		{
			Name:      "#3",
			ClusterID: "BCS-K8S-40000",
			Namespace: "d",
			Want:      "40000:0d750195d",
		},
		{
			Name:      "#4",
			ClusterID: "BCS-K8S-40000",
			Namespace: "",
			Want:      "40000:8f00b204",
		},
		{
			Name:      "#5",
			ClusterID: "BCS-40000",
			Namespace: "",
			Want:      "40000:8f00b204",
		},
		{
			Name:      "#6",
			ClusterID: "-40000",
			Namespace: "",
			Want:      "40000:8f00b204",
		},
		{
			Name:      "#7",
			ClusterID: "40000",
			Namespace: "",
			Want:      "40000:8f00b204",
		},
		{
			Name:      "#8",
			ClusterID: "0",
			Namespace: "",
			Want:      "0:8f00b204",
		},
		{
			Name:      "#9",
			ClusterID: "",
			Namespace: "",
			Want:      ":8f00b204",
		},
	}

	for _, v := range cases {
		t.Run(v.Name, func(t *testing.T) {
			if got := CalcIAMNsID(v.ClusterID, v.Namespace); got != v.Want {
				t.Errorf("got: %s, want: %s, clusterID: %s, namespace: %s", got, v.Want, v.ClusterID, v.Namespace)
			}
		})
	}
}

func TestGenerateString(t *testing.T) {
	appCode := "bcs"
	for i := 0; i < 5; i++ {
		result := GenerateEventID(appCode, uuid.NewString())
		if len(strings.Split(result, "-")) != 3 {
			t.Error("Generated string has invalid number of sections")
		}
		if !strings.HasPrefix(result, "bcs-") {
			t.Error("Generated string does not start with testapp-")
		}
		if len(strings.Split(result, "-")[1]) != 14 {
			t.Error("Timestamp section of generated string has invalid length")
		}
		if len(strings.Split(result, "-")[2]) != 16 {
			t.Error("MD5 substring section of generated string has invalid length")
		}
	}
}
