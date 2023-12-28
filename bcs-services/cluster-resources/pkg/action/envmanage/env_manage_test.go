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

// Package envmanage environment manage
package envmanage

import (
	"testing"

	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestValidateAssociate(t *testing.T) {
	associates := []*clusterRes.Associate{
		{
			Cluster:   "1",
			Namespace: "1",
		}, {
			Cluster:   "1",
			Namespace: "2",
		}, {
			Cluster:   "2",
			Namespace: "1",
		}, {
			Cluster:   "2",
			Namespace: "2",
		},
	}
	err := validateAssociate(associates)
	if err != nil {
		t.Errorf("validateAssociate err: %s", err)
	}
}
