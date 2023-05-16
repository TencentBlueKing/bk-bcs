/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

func TestFilterByOwnerRefs(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"name": "deploy-name-1",
						"kind": resCsts.Deploy,
					},
				},
				"name": "rs-name-1",
			},
		},
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"name": "deploy-name-2",
						"kind": resCsts.Deploy,
					},
				},
				"name": "rs-name-2",
			},
		},
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"name": "deploy-name-2",
						"kind": resCsts.Deploy,
					},
				},
				"name": "rs-name-3",
			},
		},
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"name": "deploy-name-3",
						"kind": resCsts.Deploy,
					},
				},
				"name": "rs-name-4",
			},
		},
	}
	ownerRefs := []map[string]string{
		{
			"name": "deploy-name-1",
			"kind": resCsts.Deploy,
		},
		{
			"name": "deploy-name-2",
			"kind": resCsts.Deploy,
		},
	}
	items = filterByOwnerRefs(items, ownerRefs)
	assert.Equal(t, 3, len(items))

	for idx, item := range items {
		assert.Equal(
			t,
			fmt.Sprintf("rs-name-%d", idx+1),
			mapx.GetStr(item.(map[string]interface{}), "metadata.name"),
		)
	}
}
