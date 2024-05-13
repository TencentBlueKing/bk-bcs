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

// Package etcd xxx
package etcd

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	clientv3 "go.etcd.io/etcd/client/v3"

	conf "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
)

var etcdClient *clientv3.Client

// Init init etcd client singleton
func Init(conf *conf.EtcdConfig) error {
	etcdEndpoints := stringx.SplitString(conf.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if len(conf.EtcdCa) != 0 && len(conf.EtcdCert) != 0 && len(conf.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(conf.EtcdCa, conf.EtcdCert, conf.EtcdKey, "")
		if err != nil {
			return err
		}
	}

	logging.Info("etcd endpoints for etcdClient: %v, with secure %t", etcdEndpoints, etcdSecure)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		TLS:         etcdTLS,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	etcdClient = cli
	return nil
}

// GetClient return etcd client singleton
func GetClient() (*clientv3.Client, error) {
	if etcdClient == nil {
		return nil, errors.New("etcd client not inited")
	}
	return etcdClient, nil
}

// Close close etcd connection
func Close() {
	if etcdClient != nil {
		_ = etcdClient.Close()
	}
}
