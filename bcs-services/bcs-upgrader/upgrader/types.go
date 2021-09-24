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

package upgrader

const (
	// SystemTypeVersion is a field representing version
	SystemTypeVersion = "version"
	// UpgraderTableName is the db table name for upgrader module
	UpgraderTableName = "bcs_upgrader"
	// VersionPrefix is the prefix for bcs upgrade program
	VersionPrefix = "u"
	// InitialVersion is the initial version for the first time to upgrade
	InitialVersion = "u1.21.199912121010"
)
