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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers/mongodb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers/zookeeper"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"gopkg.in/mgo.v2"
)

const (
	configKeySep     = "/"
	mongodbConfigKey = "mongodb"
	zkConfigKey      = "zk"
)

type APIResource struct {
	Conf      *options.StorageOptions
	ActionsV1 []*httpserver.Action
	dbInfoMap map[string]*operator.DBInfo
}

var api = APIResource{}

// Get *APIResource that loads config from config-file and handlers from api-actions
func GetAPIResource() *APIResource {
	return &api
}

// Set storageConfig to APIResource
func (a *APIResource) SetConfig(op *options.StorageOptions) {
	a.Conf = op

	// parse config-map from file
	dbConfig := a.ParseDBConfig()
	blog.Infof("Begin to parse databases.")

	// parse db config from config-map
	a.dbInfoMap = make(map[string]*operator.DBInfo)
	for _, key := range dbConfig.KeyList {
		if _, ok := a.dbInfoMap[key]; ok {
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
				SetUnhealthy(mongodbConfigKey, err.Error())
			}
		case zkConfigKey:
			if err = a.parseZk(key, dbConfig); err != nil {
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
}

//
func (a *APIResource) InitActions() {
	a.ActionsV1 = append(a.ActionsV1, actions.GetApiV1Action()...)
}

func (a *APIResource) ParseDBConfig() (dbConf *conf.Config) {
	dbConf = new(conf.Config)
	if _, err := os.Stat(a.Conf.DBConfig); !os.IsNotExist(err) {
		blog.Infof("Parsing config file: %s", a.Conf.DBConfig)
		dbConf.InitConfig(a.Conf.DBConfig)
	} else {
		blog.Errorf("Config file not exists: %s", a.Conf.DBConfig)
	}
	return
}

func (a *APIResource) GetMongodbTankName(name string) string {
	return getDriverName(mongodbConfigKey, name)
}

func (a *APIResource) GetZkTankName(name string) string {
	return getDriverName(zkConfigKey, name)
}

func (a *APIResource) getDBInfo(key string) (info *operator.DBInfo) {
	var ok bool
	info, ok = a.dbInfoMap[key]
	if !ok {
		blog.Errorf("Database Config not exists: %s", key)
	}
	return
}

func (a *APIResource) parseMongodb(key string, dbConf *conf.Config) (err error) {
	address := dbConf.Read(key, "Addr")
	timeoutRaw := dbConf.Read(key, "ConnectTimeout")
	timeout, _ := strconv.Atoi(timeoutRaw)
	database := dbConf.Read(key, "Database")
	username := dbConf.Read(key, "Username")
	password := dbConf.Read(key, "Password")
	opLogCollection := dbConf.Read(key, "OpLogCollection")
	isListener := dbConf.Read(key, "IsListener")
	listenerName := dbConf.Read(key, "ListenerName")

	var mode, modeRaw int
	if modeRaw, err = strconv.Atoi(dbConf.Read(key, "Mode")); err != nil {
		mode = int(mgo.Strong)
	} else {
		mode = modeRaw
	}

	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	a.dbInfoMap[key] = &operator.DBInfo{
		Addr:           strings.Split(address, ","),
		ConnectTimeout: time.Second * time.Duration(timeout),
		Database:       database,
		Username:       username,
		Password:       password,
		Mode:           mode,
		ListenerName:   listenerName,
	}

	if !runWithTimeout(
		func() {
			if err = mongodb.RegisterMongodbTank(key, a.dbInfoMap[key]); err != nil {
				blog.Errorf("register db config failed: %s | %v", key, err)
			}
		},
		2*time.Second,
	) {
		blog.Errorf("register db config timeout: %s", key)
		return fmt.Errorf("db connect timeout: %s", key)
	}

	if err != nil {
		return
	}

	// check db connect and ping.
	tank := mongodb.NewMongodbTank(key)
	if err = tank.GetError(); err == nil {
		blog.Infof("Complete parse mongodb config: %s", key)

		// start watch if isWatch is set
		if !(isListener == "" || isListener == "false") {
			go mongodb.StartWatch(tank.Using("local").From(opLogCollection), listenerName)
			blog.Infof("Start mongodb watch: %s", key)
		}
	} else {
		blog.Errorf("Check db config failed: %s | %v", key, err)
	}
	return err
}

func (a *APIResource) parseZk(key string, dbConf *conf.Config) (err error) {
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

	a.dbInfoMap[key] = &operator.DBInfo{
		Addr:           strings.Split(address, ","),
		ConnectTimeout: time.Second * time.Duration(timeout),
		Database:       database,
		Username:       username,
		Password:       password,
	}

	if !runWithTimeout(
		func() {
			if err = zookeeper.RegisterZkTank(key, a.dbInfoMap[key]); err != nil {
				blog.Errorf("register db config failed: %s | %v", key, err)
			}
		},
		2*time.Second,
	) {
		blog.Errorf("register db config timeout: %s", key)
		return fmt.Errorf("db connect timeout: %s", key)
	}

	if err != nil {
		return
	}

	// check db connect and ping.
	if err = zookeeper.NewZkTank(key).GetError(); err == nil {
		blog.Infof("Complete parse zk config: %s", key)
	} else {
		blog.Errorf("Check db config failed: %s | %v", key, err)
	}
	return err
}

func getDriverName(prefix, name string) string {
	return prefix + configKeySep + name
}

func runWithTimeout(f func(), duration time.Duration) bool {
	c := make(chan struct{})
	go func() {
		f()
		c <- struct{}{}
	}()

	select {
	case <-c:
		return true
	case <-time.After(duration):
		return false
	}
}
