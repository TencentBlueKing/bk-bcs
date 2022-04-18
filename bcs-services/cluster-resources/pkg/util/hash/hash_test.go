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

package hash_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/hash"
)

func TestMD5Digest(t *testing.T) {
	ret := hash.MD5Digest("ClusterResources")
	assert.Equal(t, "85360bdaa905e253dbc6c0917ed05d5e", ret)
	ret = hash.MD5Digest("default")
	assert.Equal(t, "c21f969b5f03d33d43e04f8f136e7682", ret)
}
