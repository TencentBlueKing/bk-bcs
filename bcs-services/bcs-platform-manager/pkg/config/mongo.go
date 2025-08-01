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

package config

// MongoConf mongo config
type MongoConf struct {
	Address        string `yaml:"address"`
	Replicaset     string `yaml:"replicaset"`
	AuthDatabase   string `yaml:"authdatabase"`
	Database       string `yaml:"database"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	ConnectTimeout uint   `yaml:"connecttimeout"`
	MaxPoolSize    uint   `yaml:"maxpoolsize"`
	MinPoolSize    uint   `yaml:"minpoolsize"`
}

// DefaultMongoConf :
func DefaultMongoConf() *MongoConf {
	// only for development
	return &MongoConf{}
}
