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

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	titleRoot           = "root"
	titleMetadata       = "metadata"
	titleContainerGroup = "containerGroup"
	titleContainers     = "containers"
	titleContainer      = "container"
)

var rootSubSchema = subSchema{
	Property: SchemaPropertyRoot,
	Source:   SchemaSourceRoot,

	Type:  TypeObject,
	Title: &titleRoot,
	Properties: map[string]*subSchema{
		"metadata":       &metadataSubSchema,
		"containerGroup": &containerGroupSubSchema,
	},
}

var metadataSubSchema = subSchema{
	Property: "metadata",
	Source:   SchemaSourceProperties,

	Type:       TypeObject,
	Title:      &titleMetadata,
	Properties: map[string]*subSchema{},
}

var containerGroupSubSchema = subSchema{
	Property: "containerGroup",
	Source:   SchemaSourceProperties,

	Type:  TypeObject,
	Title: &titleContainerGroup,
	Properties: map[string]*subSchema{
		"containers": &containersSubSchema,
	},
}

var containersSubSchema = subSchema{
	Property: "containers",
	Source:   SchemaSourceProperties,

	Type:  TypeArray,
	Title: &titleContainers,
	Items: &containerSubSchema,
}

var containerSubSchema = subSchema{
	Property: "container",
	Source:   SchemaSourceItems,

	Type:       TypeObject,
	Title:      &titleContainer,
	Properties: map[string]*subSchema{},
}

func TestGenNodePaths(t *testing.T) {
	metadataSubSchema.Parent = &rootSubSchema
	containerGroupSubSchema.Parent = &rootSubSchema
	containersSubSchema.Parent = &containerGroupSubSchema
	containerSubSchema.Parent = &containersSubSchema

	assert.Equal(
		t, "properties.containerGroup.properties.containers.items.title",
		genNodePaths(&containerSubSchema, "title"),
	)
	assert.Equal(
		t, "properties.metadata.type",
		genNodePaths(&metadataSubSchema, ".type"),
	)
}

func TestGenSubPath(t *testing.T) {
	assert.Equal(t, ".a.b.c", genSubPath("a.b", "c"))
	assert.Equal(t, ".a.b.(c.d)", genSubPath("a.b", "c.d"))
}

func TestGenSubPathWithIdx(t *testing.T) {
	assert.Equal(t, ".a.b.c[1]", genSubPathWithIdx("a.b", "c", 1))
	assert.Equal(t, ".a.b.(c.d)[3]", genSubPathWithIdx("a.b", "c.d", 3))
}
