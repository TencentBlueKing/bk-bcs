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

package config

// RepositoryConf 对象存储配置
type RepositoryConf struct {
	StorageType string     `yaml:"storage_type"`
	Bkrepo      BkRepoConf `yaml:"bkrepo"`
	Cos         CosConf    `yaml:"cos"`
}

// BkRepoConf bkrepo配置
type BkRepoConf struct {
	Endpoint string `yaml:"endpoint"`
	Project  string `yaml:"project"`
	UserName string `yaml:"user_name"`
	Password string `yaml:"password"`
	Repo     string `yaml:"repo"`
}

// CosConf cos配置
type CosConf struct {
	BucketName string `yaml:"bucket_name"`
	Endpoint   string `yaml:"endpoint"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
}
