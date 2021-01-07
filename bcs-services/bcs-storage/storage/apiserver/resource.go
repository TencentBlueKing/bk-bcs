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

package apiserver

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/check"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers/mongo"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
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

// APIResource api resource object
type APIResource struct {
	Conf      *options.StorageOptions
	ActionsV1 []*httpserver.Action
	storeMap  map[string]store.Store
	dbMap     map[string]drivers.DB
	ebusMap   map[string]*watchbus.EventBus
	msgQueue  msgqueue.MessageQueue
}

var api = APIResource{}

// GetAPIResource that loads config from config-file and handlers from api-actions
func GetAPIResource() *APIResource {
	return &api
}

// SetConfig Set storageConfig to APIResource
func (a *APIResource) SetConfig(op *options.StorageOptions) {
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
			if err := a.parseQueueInit(key, queueConfig); err != nil {
				blog.Errorf("parse queue config failed, err %s", err.Error())
				SetUnhealthy(queueConfigKey, err.Error())
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
	blog.Infof("MsgQueue parsing completed.")
}

// InitActions init actions
func (a *APIResource) InitActions() {
	a.ActionsV1 = append(a.ActionsV1, actions.GetApiV1Action()...)
}

// ParseDBConfig parse db config
func (a *APIResource) ParseDBConfig() (dbConf *conf.Config) {
	dbConf = new(conf.Config)
	if _, err := os.Stat(a.Conf.DBConfig); !os.IsNotExist(err) {
		blog.Infof("Parsing dbConfig file: %s", a.Conf.DBConfig)
		dbConf.InitConfig(a.Conf.DBConfig)
	} else {
		blog.Errorf("Config file not exists: %s", a.Conf.DBConfig)
	}
	return
}

// ParseQueueConfig parse queue config
func (a *APIResource) ParseQueueConfig() (queueConf *conf.Config) {
	queueConf = new(conf.Config)

	if _, err := os.Stat(a.Conf.QueueConfig); !os.IsNotExist(err) {
		blog.Infof("Parsing queueConfig file: %s", a.Conf.QueueConfig)
		queueConf.InitConfig(a.Conf.QueueConfig)
	} else {
		blog.Errorf("Config file not exists: %s", a.Conf.QueueConfig)
	}

	return
}

// GetMsgQueue get queue client
func (a *APIResource) GetMsgQueue() msgqueue.MessageQueue {
	return a.msgQueue
}

func (a *APIResource) parseQueueInit(key string, queueConf *conf.Config) error {
	flagRaw := queueConf.Read(key, "QueueFlag")
	kind := queueConf.Read(key, "QueueKind")

	flag, err := strconv.ParseBool(flagRaw)
	if err != nil {
		return err
	}

	resource := queueConf.Read(key, "Resource")
	resourceToQueue := map[string]string{}
	arrayResource := strings.Split(resource, ",")
	for _, r := range arrayResource {
		resourceToQueue[r] = r
	}

	address := queueConf.Read(key, "Address")

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

	msgQueue, err := msgqueue.NewMsgQueue(flag, msgqueue.QueueKind(kind), address, resourceToQueue,
		exchangeOption, natStreamingOption, publishOption, subscribeOption)
	if err != nil {
		msgErr := fmt.Errorf("create queue failed, err %s", err.Error())
		blog.Errorf("create queue failed, err %s", err.Error())
		return msgErr
	}

	a.msgQueue = msgQueue
	blog.Infof("init queue successfully, queue kind[%s] queue flag[%v]", kind, flag)
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
	timeoutRaw := dbConf.Read(key, "ConnectTimeout")
	timeout, err := strconv.Atoi(timeoutRaw)
	if err != nil {
		return err
	}
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
		Hosts:                 strings.Split(address, ","),
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

func getPublishOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	publishDeliveryRaw := queueConf.Read(key, "PublishDelivery")
	publishDelivery, err := strconv.Atoi(publishDeliveryRaw)
	if err != nil {
		return nil, err
	}

	return msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			DeliveryMode: uint8(publishDelivery),
		}), nil
}

func getNatStreamingOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	clusterID := queueConf.Read(key, "ClusterId")
	connectTimeoutRaw := queueConf.Read(key, "ConnectTimeout")
	connectTimeout, err := strconv.Atoi(connectTimeoutRaw)
	if err != nil {
		return nil, err
	}
	connectRetryRaw := queueConf.Read(key, "ConnectRetry")
	connectRetry, err := strconv.ParseBool(connectRetryRaw)
	if err != nil {
		return nil, err
	}

	return msgqueue.NatsOpts(
		&msgqueue.NatsOptions{
			ClusterID:      clusterID,
			ConnectTimeout: time.Duration(connectTimeout) * time.Second,
			ConnectRetry:   connectRetry,
		}), nil
}

func getQueueExchangeOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	exchangeName := queueConf.Read(key, "ExchangeName")
	exchangeDurableRaw := queueConf.Read(key, "ExchangeDurable")
	exchangeDurable, err := strconv.ParseBool(exchangeDurableRaw)
	if err != nil {
		return nil, err
	}
	exchagePrefetchCountRaw := queueConf.Read(key, "ExchangePrefetchCount")
	exchagePrefetchCount, err := strconv.Atoi(exchagePrefetchCountRaw)
	if err != nil {
		return nil, err
	}
	exchangePrefetchGlobalRaw := queueConf.Read(key, "ExchangePrefetchGlobal")
	exchangePrefetchGlobal, err := strconv.ParseBool(exchangePrefetchGlobalRaw)
	if err != nil {
		return nil, err
	}

	return msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           exchangeName,
			Durable:        exchangeDurable,
			PrefetchCount:  exchagePrefetchCount,
			PrefetchGlobal: exchangePrefetchGlobal,
		}), nil

}

func getQueueSubscribeOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	subDurableRaw := queueConf.Read(key, "SubDurable")
	subDurable, err := strconv.ParseBool(subDurableRaw)
	if err != nil {
		return nil, err
	}
	subDisableAutoAckRaw := queueConf.Read(key, "SubDisableAutoAck")
	subDisableAutoAck, err := strconv.ParseBool(subDisableAutoAckRaw)
	if err != nil {
		return nil, err
	}
	subAckOnSuccessRaw := queueConf.Read(key, "SubAckOnSuccess")
	subAckOnSuccess, err := strconv.ParseBool(subAckOnSuccessRaw)
	if err != nil {
		return nil, err
	}

	subRequeueOnErrorRaw := queueConf.Read(key, "SubRequeueOnError")
	subRequeueOnError, err := strconv.ParseBool(subRequeueOnErrorRaw)
	if err != nil {
		return nil, err
	}

	subDeliverAllMessageRaw := queueConf.Read(key, "SubDeliverAllMessage")
	subDeliverAllMessage, err := strconv.ParseBool(subDeliverAllMessageRaw)
	if err != nil {
		return nil, err
	}

	subManualAckModeRaw := queueConf.Read(key, "SubManualAckMode")
	subManualAckMode, err := strconv.ParseBool(subManualAckModeRaw)
	if err != nil {
		return nil, err
	}
	subEnableAckWaitRaw := queueConf.Read(key, "SubEnableAckWait")
	subEnableAckWait, err := strconv.ParseBool(subEnableAckWaitRaw)
	if err != nil {
		return nil, err
	}

	subAckWaitDurationRaw := queueConf.Read(key, "SubAckWaitDuration")
	subAckWaitDuration, err := strconv.Atoi(subAckWaitDurationRaw)
	if err != nil {
		return nil, err
	}

	subMaxInFlightRaw := queueConf.Read(key, "SubMaxInFlight")
	subMaxInFlight, err := strconv.Atoi(subMaxInFlightRaw)
	if err != nil {
		return nil, err
	}

	// parse queueArguments
	arguments := make(map[string]interface{})
	queueArgumentsRaw := queueConf.Read(key, "QueueArguments")
	queueArguments := strings.Split(queueArgumentsRaw, ";")
	if len(queueArguments) > 0 {
		for _, data := range queueArguments {
			dList := strings.Split(data, ":")
			if len(dList) == 2 {
				arguments[dList[0]] = dList[1]
			}
		}
	}

	return msgqueue.SubscribeOpts(
		&msgqueue.SubscribeOptions{
			DisableAutoAck:    subDisableAutoAck,
			Durable:           subDurable,
			AckOnSuccess:      subAckOnSuccess,
			RequeueOnError:    subRequeueOnError,
			DeliverAllMessage: subDeliverAllMessage,
			ManualAckMode:     subManualAckMode,
			EnableAckWait:     subEnableAckWait,
			AckWaitDuration:   time.Duration(subAckWaitDuration) * time.Second,
			MaxInFlight:       subMaxInFlight,
			QueueArguments:    arguments,
		}), nil
}
