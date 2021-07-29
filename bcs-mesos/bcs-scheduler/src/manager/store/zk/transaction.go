/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zk

import (
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

func getTransactionRootPath() string {
	return "/" + bcsRootNode + "/" + transactionNode
}

// FetchTransaction fetch transaction
func (store *managerStore) FetchTransaction(namespace, name string) (*types.Transaction, error) {
	path := getTransactionRootPath() + "/" + namespace + "/" + name
	data, err := store.Db.Fetch(path)
	if err != nil {
		return nil, err
	}

	transaction := &types.Transaction{}
	if err := json.Unmarshal(data, transaction); err != nil {
		blog.Errorf("falied to unmarshal transaction %s, err %s", string(data), err.Error())
		return nil, err
	}
	return transaction, nil
}

// SaveTransaction save transaction
func (store *managerStore) SaveTransaction(transaction *types.Transaction) error {
	data, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	path := getTransactionRootPath() + "/" + transaction.Namespace + "/" + transaction.TransactionID

	return store.Db.Insert(path, string(data))
}

func (store *managerStore) listTransactions(ns string) ([]*types.Transaction, error) {
	path := getTransactionRootPath() + "/" + ns
	idList, err := store.Db.List(path)
	if err != nil {
		blog.Errorf("fail to list transacation ids in ns %s, err %s", ns, err.Error())
		return nil, err
	}
	if len(idList) == 0 {
		blog.V(3).Infof("no transaction in ns %s", ns)
		return nil, nil
	}

	var objs []*types.Transaction
	for _, id := range idList {
		obj, err := store.FetchTransaction(ns, id)
		if err != nil {
			blog.Warnf("failed to fetch transaction by ns %s id %s, err %s", ns, id, err.Error())
			continue
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

// ListTransaction list transaction by namespace
func (store *managerStore) ListTransaction(ns string) ([]*types.Transaction, error) {
	return store.listTransactions(ns)
}

// ListAllTransaction list all transaction
func (store *managerStore) ListAllTransaction() ([]*types.Transaction, error) {
	nss, err := store.ListObjectNamespaces(transactionNode)
	if err != nil {
		return nil, err
	}

	var retList []*types.Transaction
	for _, ns := range nss {
		blog.Infof("list all transaction ns %s", ns)
		objs, err := store.listTransactions(ns)
		if err != nil {
			blog.Errorf("failed to fetch service by ns %s", ns)
			continue
		}
		retList = append(retList, objs...)
	}
	return retList, nil
}

// DeleteTransaction delete transaction
func (store *managerStore) DeleteTransaction(namespace, name string) error {
	path := getTransactionRootPath() + "/" + namespace + "/" + name
	if err := store.Db.Delete(path); err != nil {
		blog.Errorf("failed to delete transaction %s/%s, err %s", namespace, name, err.Error())
		return err
	}
	return nil
}
