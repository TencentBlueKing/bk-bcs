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

package option

const (
	// TerraformBinPath  terraform 命令存放目录
	TerraformBinPath = "/usr/local/bin/terraform"

	// RepositoryStorePath 代码存储路径
	RepositoryStorePath = "/data/bcs/terraform"
)

// GetRepoStoragePath 返回仓库的存储位置
func GetRepoStoragePath(tfName, tfUID string) string {
	return RepositoryStorePath + "/" + tfName + "/" + tfUID
}
