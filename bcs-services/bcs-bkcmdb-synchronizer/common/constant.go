/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

const (
	// BCS_BKCMDB_SYNC_DIR zk path for synchronizer
	BCS_BKCMDB_SYNC_DIR = "/bcs/services/bkcmdb-synchronizer"
	// BCS_BKCMDB_SYNC_DIR_CLUSTER cluster dir for synchronizer
	BCS_BKCMDB_SYNC_DIR_CLUSTER = BCS_BKCMDB_SYNC_DIR + "/cluster"
	// BCS_BKCMDB_SYNC_DIR_WORKER synchronizer worker instance dir
	BCS_BKCMDB_SYNC_DIR_WORKER = BCS_BKCMDB_SYNC_DIR + "/worker"

	// BCS_BKCMDB_DEFAULT_SET_NAME bcs default set name in bk cmdb
	BCS_BKCMDB_DEFAULT_SET_NAME = "bkbcs"
	// BCS_BKCMDB_DEFAULT_MODLUE_NAME bcs default module name in bk cmdb
	BCS_BKCMDB_DEFAULT_MODLUE_NAME = "bkbcs"

	// BCS_BKCMDB_ANNOTATIONS_SET_KEY key of bcs annotations for bk cmdb
	BCS_BKCMDB_ANNOTATIONS_SET_KEY = "set.bkcmdb.bkbcs.tencent.com"
	// BCS_BKCMDB_ANNOTATIONS_MODULE_KEY key of bcs annotations for bk cmdb
	BCS_BKCMDB_ANNOTATIONS_MODULE_KEY = "module.bkcmdb.bkbcs.tencent.com"
)
