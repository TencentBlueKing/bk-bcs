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

package apiserver

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/check"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/watchbus"
)

const (
	configKeySep     = "/"
	mongodbConfigKey = "mongodb"
	zkConfigKey      = "zk"
	queueConfigKey   = "queue"
)

// MessageQueue queue object
type MessageQueue struct {
	QueueFlag       bool
	ResourceToQueue map[string]string
	MsgQueue        msgqueue.MessageQueue
}

// APIResource api resource object
type APIResource struct {
	Conf      *options.StorageOptions
	ActionsV1 []*httpserver.Action
	storeMap  map[string]store.Store
	dbMap     map[string]drivers.DB
	ebusMap   map[string]*watchbus.EventBus
	msgQueue  *MessageQueue
}

var api = APIResource{}

// GetAPIResource that loads config from config-file and handlers from api-actions
func GetAPIResource() *APIResource {
	return &api
}

// SetConfig Set storageConfig to APIResource
func (a *APIResource) SetConfig(op *options.StorageOptions) error {
	a.Conf = op
	// parse config-map from file
	dbConfig := a.ParseDBConfig()

	blog.Infof("Begin to parse databases.")

	a.storeMap = make(map[string]store.Store)
	a.dbMap = make(map[string]drivers.DB)
	a.ebusMap = make(map[string]*watchbus.EventBus)
	for _, key := range dbConfig.KeyList {
		if _, ok := a.storeMap[key]; ok {
			blog.Warnf("Store config duplicated: %s", key)
			continue
		}
		if _, ok := a.dbMap[key]; ok {
			blog.Warnf("Database config duplicated: %s", key)
			continue
		}
		s := strings.Split(key, configKeySep)
		if len(s) != 2 {
			blog.Errorf("Database config invalid: %s | Format like mongodb/dynamic.", key)
			continue
		}
		var err error
		switch s[0] {
		case mongodbConfigKey:
			if err = a.parseMongodb(key, dbConfig); err != nil {
				blog.Errorf("parse mongodb failed, err %s", err.Error())
				SetUnhealthy(mongodbConfigKey, err.Error())
			}
		case zkConfigKey:
			if err = a.parseZk(key, dbConfig); err != nil {
				blog.Errorf("parse zookeeper failed, err %s", err.Error())
				SetUnhealthy(zkConfigKey, err.Error())
			}
		default:
			err = storageErr.DatabaseConfigUnknown
			blog.Errorf("%v: %s", err, key)
			SetUnhealthy("unknown_config", fmt.Sprintf("%v: %s", err, key))
		}
		if err != nil {
			check.Occur(err)
		}
	}
	blog.Infof("Databases parsing completed.")

	// parse config-map from queue file
	queueConfig := a.ParseQueueConfig()

	blog.Infof("Begin to parse queueConfig.")

	for _, key := range queueConfig.KeyList {
		var err error
		switch key {
		case queueConfigKey:
			if err = a.parseQueueInit(key, queueConfig); err != nil {
				queueErr := fmt.Errorf("parse queue config failed, err %s", err.Error())
				blog.Errorf(queueErr.Error())
				SetUnhealthy(queueConfigKey, queueErr.Error())
				return queueErr
			}
		default:
			err = storageErr.QueueConfigUnknown
			blog.Errorf("%v: %s", err, key)
			SetUnhealthy("unknown_config", fmt.Sprintf("%v: %s", err, key))
		}
		if err != nil {
			check.Occur(err)
		}
	}
	blog.Infof("MsgQueue parsing completed.")
	return nil
}

// InitActions init actions
func (a *APIResource) InitActions() {
	a.ActionsV1 = append(a.ActionsV1, actions.GetApiV1Action()...)
}

// ParseDBConfig parse db config
func (a *APIResource) ParseDBConfig() *conf.Config {
	dbConf := new(conf.Config)
	if _, err := os.Stat(a.Conf.DBConfig); !os.IsNotExist(err) {
		blog.Infof("Parsing config file: %s", a.Conf.DBConfig)
		dbConf.InitConfig(a.Conf.DBConfig)
	} else {
		blog.Errorf("Config file not exists: %s", a.Conf.DBConfig)
	}
	return dbConf
}

// ParseQueueConfig parse queue config
func (a *APIResource) ParseQueueConfig() *conf.Config {
	queueConf := new(conf.Config)

	if _, err := os.Stat(a.Conf.QueueConfig); !os.IsNotExist(err) {
		blog.Infof("Parsing queueConfig file: %s", a.Conf.QueueConfig)
		queueConf.InitConfig(a.Conf.QueueConfig)
	} else {
		blog.Errorf("Config file not exists: %s", a.Conf.QueueConfig)
	}

	return queueConf
}

// GetMsgQueue get queue client
func (a *APIResource) GetMsgQueue() *MessageQueue {
	return a.msgQueue
}

func (a *APIResource) getMsgQueueFlag(key string, queueConf *conf.Config) bool {
	flagRaw := queueConf.Read(key, "QueueFlag")
	if flagRaw == "" {
		return false
	}

	flag, err := strconv.ParseBool(flagRaw)
	if err != nil {
		return false
	}

	blog.Infof("queueFlag is [%v]", flag)
	return flag
}

func (a *APIResource) parseQueueInit(key string, queueConf *conf.Config) error {
	queueFlag := a.getMsgQueueFlag(key, queueConf)
	if !queueFlag {
		a.msgQueue = &MessageQueue{
			QueueFlag: queueFlag,
			MsgQueue:  nil,
		}

		return nil
	}

	commonOptions, err := getQueueCommonOptions(key, queueConf)
	if err != nil {
		return err
	}
	commonOption := msgqueue.CommonOpts(commonOptions)

	exchangeOption, err := getQueueExchangeOptions(key, queueConf)
	if err != nil {
		return err
	}

	natStreamingOption, err := getNatStreamingOptions(key, queueConf)
	if err != nil {
		return err
	}

	publishOption, err := getPublishOptions(key, queueConf)
	if err != nil {
		return err
	}

	subscribeOption, err := getQueueSubscribeOptions(key, queueConf)
	if err != nil {
		return err
	}

	msgQueue, err := msgqueue.NewMsgQueue(commonOption, exchangeOption, natStreamingOption, publishOption, subscribeOption)
	if err != nil {
		msgErr := fmt.Errorf("create queue failed, err %s", err.Error())
		blog.Errorf("create queue failed, err %s", err.Error())
		return msgErr
	}
	queueKind, _ := msgQueue.String()

	a.msgQueue = &MessageQueue{
		QueueFlag:       queueFlag,
		MsgQueue:        msgQueue,
		ResourceToQueue: commonOptions.ResourceToQueue,
	}

	blog.Infof("init queue[%s] successfully, sub queue[%v]", queueKind, a.msgQueue.ResourceToQueue)
	return nil
}

// GetEventBus get event bus by key
func (a *APIResource) GetEventBus(key string) *watchbus.EventBus {
	return a.ebusMap[key]
}

// GetDBClient get db client by key
func (a *APIResource) GetDBClient(key string) drivers.DB {
	return a.dbMap[key]
}

// GetStoreClient get store client by keys
func (a *APIResource) GetStoreClient(key string) store.Store {
	return a.storeMap[key]
}

func (a *APIResource) parseMongodb(key string, dbConf *conf.Config) error {
	address := dbConf.Read(key, "Addr")
	replicaset := dbConf.Read(key, "Replicaset")
	timeoutRaw := dbConf.Read(key, "ConnectTimeout")
	timeout, err := strconv.Atoi(timeoutRaw)
	if err != nil {
		return err
	}
	authDatabase := dbConf.Read(key, "AuthDatabase")
	database := dbConf.Read(key, "Database")
	username := dbConf.Read(key, "Username")
	password := dbConf.Read(key, "Password")
	maxPoolSizeRaw := dbConf.Read(key, "MaxPoolSize")
	maxPoolSize := 0
	if len(maxPoolSizeRaw) != 0 {
		maxPoolSize, err = strconv.Atoi(maxPoolSizeRaw)
		if err != nil {
			return err
		}
	}
	minPoolSizeRaw := dbConf.Read(key, "MinPoolSize")
	minPoolSize := 0
	if len(minPoolSizeRaw) != 0 {
		minPoolSize, err = strconv.Atoi(minPoolSizeRaw)
		if err != nil {
			return err
		}
	}

	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	mongoOptions := &mongo.Options{
		AuthDatabase:          authDatabase,
		Hosts:                 strings.Split(address, ","),
		Replicaset:            replicaset,
		ConnectTimeoutSeconds: timeout,
		Database:              database,
		Username:              username,
		Password:              password,
		MaxPoolSize:           uint64(maxPoolSize),
		MinPoolSize:           uint64(minPoolSize),
	}

	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		blog.Errorf("create mongo db with %s failed, err %s", key, err.Error())
		return fmt.Errorf("create mongo db with %s failed, err %s", key, err.Error())
	}

	err = mongoDB.Ping()
	if err != nil {
		blog.Errorf("ping mongo db failed, err %s", err.Error())
		return fmt.Errorf("ping mongo db failed, err %s", err.Error())
	}

	ignoreDeleteEventRaw := dbConf.Read(key, "IgnoreDeleteEvent")

	ebus := watchbus.NewEventBus(mongoDB)
	if ignoreDeleteEventRaw == "true" {
		ebus.SetCondition(operator.NewBranchCondition(operator.Mat,
			operator.NewLeafCondition(operator.Ne, operator.M{"operationType": "delete"})))
	}
	a.dbMap[key] = mongoDB
	a.ebusMap[key] = ebus
	blog.Infof("init mongo db with key %s successfully", key)
	return nil
}

func (a *APIResource) parseZk(key string, dbConf *conf.Config) error {
	address := dbConf.Read(key, "Addr")
	timeoutRaw := dbConf.Read(key, "ConnectTimeout")
	timeout, _ := strconv.Atoi(timeoutRaw)
	database := dbConf.Read(key, "Database")
	username := dbConf.Read(key, "Username")
	password := dbConf.Read(key, "Password")

	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	zkOpt := &zookeeper.Options{
		BasePath:              database,
		Addrs:                 strings.Split(address, ","),
		ConnectTimeoutSeconds: timeout,
		Database:              database,
		Username:              username,
		Password:              password,
	}

	zkStore, err := zookeeper.NewStore(zkOpt)
	if err != nil {
		return err
	}
	a.storeMap[key] = zkStore
	blog.Infof("init zookeeper with key %s successfully", key)
	return nil
}
