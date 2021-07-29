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

package etcd

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckTransactionExist check if transaction exists
func (store *managerStore) CheckTransactionExist(namespace, name string) (string, bool) {
	trans, _ := store.fetchTransactionInDB(namespace, name)
	if trans != nil {
		return trans.ResourceVersion, true
	}
	return "", false
}

func (store *managerStore) fetchTransactionInDB(namespace, name string) (*types.Transaction, error) {
	client := store.BkbcsClient.BcsTransactions(namespace)
	v2Trans, err := client.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	obj := v2Trans.Spec.Transaction
	obj.ResourceVersion = v2Trans.ResourceVersion
	return &obj, nil
}

// FetchTransaction fetch transaction
func (store *managerStore) FetchTransaction(namespace, name string) (*types.Transaction, error) {
	trans := getCacheTransaction(namespace, name)
	if trans == nil {
		return trans, schStore.ErrNoFound
	}
	return trans, nil
}

// SaveTransaction save transaction
func (store *managerStore) SaveTransaction(transaction *types.Transaction) error {
	err := store.checkNamespace(transaction.Namespace)
	if err != nil {
		return err
	}
	client := store.BkbcsClient.BcsTransactions(transaction.Namespace)
	v2Trans := &v2.BcsTransaction{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdBcsTransaction,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      transaction.TransactionID,
			Namespace: transaction.Namespace,
		},
		Spec: v2.BcsTransactionSpec{
			Transaction: *transaction,
		},
	}

	rv, exist := store.CheckTransactionExist(transaction.Namespace, transaction.TransactionID)
	if exist {
		v2Trans.ResourceVersion = rv
		v2Trans, err = client.Update(context.Background(), v2Trans, metav1.UpdateOptions{})
	} else {
		v2Trans, err = client.Create(context.Background(), v2Trans, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	transaction.ResourceVersion = v2Trans.ResourceVersion
	saveCacheTransaction(transaction)
	return nil
}

// ListTransaction list transaction by namespace
func (store *managerStore) ListTransaction(ns string) ([]*types.Transaction, error) {
	if cacheMgr.isOK {
		return listCacheRunAsTransaction(ns)
	}

	client := store.BkbcsClient.BcsTransactions(ns)
	v2Trans, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	trans := make([]*types.Transaction, 0, len(v2Trans.Items))
	for _, tran := range v2Trans.Items {
		obj := tran.Spec.Transaction
		obj.ResourceVersion = tran.ResourceVersion
		trans = append(trans, &obj)
	}
	return trans, nil
}

// ListAllTransaction list all transaction
func (store *managerStore) ListAllTransaction() ([]*types.Transaction, error) {
	if cacheMgr.isOK {
		return listCacheTransactions()
	}

	client := store.BkbcsClient.BcsTransactions("")
	v2Trans, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	trans := make([]*types.Transaction, 0, len(v2Trans.Items))
	for _, tran := range v2Trans.Items {
		obj := tran.Spec.Transaction
		obj.ResourceVersion = tran.ResourceVersion
		trans = append(trans, &obj)
	}
	return trans, nil
}

// DeleteTransaction delete transaction
func (store *managerStore) DeleteTransaction(namespace, name string) error {
	client := store.BkbcsClient.BcsTransactions(namespace)
	err := client.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteCacheTransaction(namespace, name)
	return nil
}
