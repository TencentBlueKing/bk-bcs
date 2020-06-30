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

package etcd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/storage"
	"hash/fnv"
	"log"
	"sort"
	"time"

	"github.com/coredns/coredns/plugin/etcd/msg"
	etcdcv3 "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

const (
	defaultPrefix = "/bluekingdns"
)

//NewStorage create new persistence storage
func NewStorage(prefix string, hosts []string, ca, pubKey, priKey string) (storage.Storage, error) {
	//create etcd client
	var cc *tls.Config
	var err error
	if ca != "" {
		cc, err = ssl.ClientTslConfVerity(ca, pubKey, priKey, "")
		if err != nil {
			return nil, err
		}
	}
	etcdCfg := etcdcv3.Config{
		Endpoints: hosts,
		TLS:       cc,
	}
	cli, cerr := etcdcv3.New(etcdCfg)
	if cerr != nil {
		return nil, cerr
	}
	s := &etcdStorage{
		prefix:    prefix,
		etcdLinks: hosts,
		client:    cli,
	}
	if prefix == "" {
		s.prefix = defaultPrefix
	}
	return s, nil
}

//etcdStorage tool for DNS data storage in etcd
//now only support A
//todo(developer): support SRV, PTR
type etcdStorage struct {
	prefix    string          //storage prefix
	etcdLinks []string        //etcd host links
	client    *etcdcv3.Client //client for etcd
}

//AddService add service dns data
func (es *etcdStorage) AddService(domain string, msgs []msg.Service) error {
	if len(msgs) == 0 {
		log.Printf("[WARN] no dns service for %s adding to etcd", domain)
		return nil
	}
	sort.Sort(storage.ServiceList(msgs))
	//create path and data.
	for _, item := range msgs {
		data, err := json.Marshal(item)
		if err != nil {
			//todo(developer): return directly?
			log.Printf("[ERROR] domain %s get json marshal err, %s", domain, err.Error())
			continue
		}
		hashKey := es.value2Hash(data)
		value := string(data)
		etcdDomain := fmt.Sprintf("%s.%s", hashKey, domain)
		etcdPath := msg.Path(etcdDomain, defaultPrefix)
		if err := es.writeRecord(etcdPath, value); err != nil {
			//todo(developer): give up all data if one record err?
			log.Printf("[ERROR] domain %s: %s write to storage err, %s", etcdDomain, item.Host, err.Error())
			continue
		}
		//todo(developer): SRV record supporting
	}
	return nil
}

//UpdateService update service dns data
func (es *etcdStorage) UpdateService(domain string, old, cur []msg.Service) error {
	//create maps
	oldMap := make(map[string]string)
	for _, item := range old {
		data, err := json.Marshal(item)
		if err != nil {
			log.Printf("[ERROR] domain %s get old [%s] json marshal err, %s", domain, item.Host, err.Error())
			continue
		}
		hashKey := es.value2Hash(data)
		oldMap[hashKey] = string(data)
	}
	curMap := make(map[string]string)
	for _, item := range cur {
		data, err := json.Marshal(item)
		if err != nil {
			log.Printf("[ERROR] domain %s get cur [%s] json marshal err, %s", domain, item.Host, err.Error())
			continue
		}
		hashKey := es.value2Hash(data)
		curMap[hashKey] = string(data)
	}
	//clean duplicated data in oldMap
	for key := range curMap {
		if _, ok := oldMap[key]; ok {
			delete(oldMap, key)
		}
	}
	//now all data in oldMap need to delete
	cxt, _ := context.WithTimeout(context.Background(), time.Second*5)
	for key := range oldMap {
		etcdDomain := fmt.Sprintf("%s.%s", key, domain)
		etcdPath := msg.Path(etcdDomain, defaultPrefix)
		if _, err := es.client.Delete(cxt, etcdPath, etcdcv3.WithPrefix()); err != nil {
			log.Printf("[ERROR] etcdStorage clean %s in Update err, %s", etcdPath, err.Error())
			continue
		}
		log.Printf("[WARN] etcdStorage clean %s success", etcdPath)
	}
	//all data in curMap need to update
	for key, value := range curMap {
		etcdDomain := fmt.Sprintf("%s.%s", key, domain)
		etcdPath := msg.Path(etcdDomain, defaultPrefix)
		if err := es.writeRecord(etcdPath, value); err != nil {
			log.Printf("[ERROR] domain %s: %s write to storage err, %s", etcdDomain, value, err.Error())
		}
	}
	return nil
}

//DeleteService update service dns data
func (es *etcdStorage) DeleteService(domain string, msgs []msg.Service) error {
	//just delete
	cxt, _ := context.WithTimeout(context.Background(), time.Second*10)
	etcdPath := msg.Path(domain, defaultPrefix)
	if _, err := es.client.Delete(cxt, etcdPath, etcdcv3.WithPrefix()); err != nil {
		log.Printf("[ERROR] etcdStorage clean %s err, %s", etcdPath, err.Error())
		return err
	}
	return nil
}

//ListServiceByName list service dns data by service name
func (es *etcdStorage) ListServiceByName(domain string) (svcList []msg.Service, err error) {
	return svcList, err
}

//ListServiceByNamespace list service dns data under namespace
func (es *etcdStorage) ListServiceByNamespace(namespace, cluster, zone string) (svcList []msg.Service, err error) {
	return svcList, err
}

//ListService list all service dns data from etcd
func (es *etcdStorage) ListService(cluster, zone string) ([]msg.Service, error) {
	return nil, nil
}

//Close close connection, release all context
func (es *etcdStorage) Close() {
}

//*****************************************************************************
// inner methods
//*****************************************************************************

//writeRecord write record with key/value to etcd
func (es *etcdStorage) writeRecord(key, value string) error {
	cxt, _ := context.WithTimeout(context.Background(), time.Second*5)
	_, err := es.client.Put(cxt, key, value)
	return err
}

func (es *etcdStorage) value2Hash(value []byte) string {
	h := fnv.New32a()
	if _, err := h.Write(value); err != nil {
		log.Printf("[ERROR] CONVERT %s to simple hash failed, %s", value, err)
	}
	return fmt.Sprintf("%x", h.Sum32())
}
