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

package mongo

const (
	fullDocumentKey      = "fullDocument"
	nsKey                = "ns"
	dbKey                = "db"
	collectionKey        = "coll"
	operationTypeKey     = "operationType"
	updateDescriptionKey = "updateDescription"
	updatedFieldsKey     = "updatedFields"
	removedFieldsKey     = "removedFields"

	operationTypeInsert       = "insert"
	operationTypeUpdate       = "update"
	operationTypeDelete       = "delete"
	operationTypeReplace      = "replace"
	operationTypeDrop         = "drop"
	operationTypeRename       = "rename"
	operationTypeDropDatabase = "dropDatabase"
	operationTypeInvalidte    = "invalidate"

	mongoAuthMichanismSha256 = "SCRAM-SHA-256"
	mongoAuthMichanismSah1   = "SCRAM-SHA-1"
	mongoAuthMichanismCr     = "MONGODB-CR"
	mongoAuthMichanismPlain  = "PLAIN"
	mongoAuthMichanismGssAPI = "GSSAPI"
	mongoAuthMichanismX509   = "MONGODB-X509"
)
