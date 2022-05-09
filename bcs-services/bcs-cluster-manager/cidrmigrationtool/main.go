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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/tke"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var dsn string
var filename string
var workmode string

var mongoAddress string
var mongoDatabase string
var mongoConnectTimeout int
var mongoUsername string
var mongoPassword string

// Cidr cidr struct in mysql
type Cidr struct {
	ID        uint      `gorm:"primary_key"`
	Vpc       string    `gorm:"not null" json:"vpc"`
	Cidr      string    `gorm:"not null" json:"cidr"`
	IPNumber  uint      `gorm:"not null" json:"ip_number"`
	Status    string    `gorm:"not null" json:"status"`
	Cluster   *string   `json:"cluster"`
	CreatedAt time.Time `json:"createAt"`
	UpdatedAt time.Time `json:"updateAt"`
}

func main() {
	flag.StringVar(&dsn, "dsn", "", "database dsn")
	flag.StringVar(&filename, "filename", "tke_cidr.json", "file name")
	flag.StringVar(&workmode, "mode", "", "dumps or upload")
	flag.StringVar(&mongoAddress, "mongo_address", "", "mongo address")
	flag.StringVar(&mongoDatabase, "mongo_database", "", "mongo database")
	flag.IntVar(&mongoConnectTimeout, "mongo_connecttime", 5, "mongo connect timeout")
	flag.StringVar(&mongoUsername, "mongo_username", "", "mongo username")
	flag.StringVar(&mongoPassword, "mongo_password", "", "mongo password")
	flag.Parse()

	blog.InitLogs(conf.LogConfig{
		LogDir:          "./logs",
		LogMaxSize:      500,
		LogMaxNum:       10,
		AlsoToStdErr:    true,
		Verbosity:       5,
		StdErrThreshold: "2",
	})

	switch workmode {
	case "dumps":
		dumpsMysql()
	case "upload":
		uploadMongo()
	default:
		blog.Fatalf("unknown work mode %s", workmode)
	}

}

func dumpsMysql() {
	if len(dsn) == 0 {
		blog.Fatalf("dsn cannot be empty")
	}

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		blog.Fatalf("err %s", err.Error())
	}
	db.DB().SetConnMaxLifetime(60 * time.Second)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(20)

	var cidrList []Cidr
	db.Table("tke_cidrs").Select("*").Scan(&cidrList)
	data, err := json.MarshalIndent(cidrList, "", "    ")
	if err != nil {
		blog.Fatalf("err %s", err.Error())
	}
	err = ioutil.WriteFile(filename, data, 755)
	if err != nil {
		blog.Fatalf("err %s", err.Error())
	}
	blog.Infof("dumps mysql tke_cidrs table successfully")
}

func uploadMongo() {
	if len(mongoAddress) == 0 {
		fmt.Printf("mongo_address cannot be empty")
	}
	if len(mongoDatabase) == 0 {
		fmt.Printf("mongo_database cannot be empty")
	}
	if len(mongoUsername) == 0 {
		fmt.Printf("mongo_username cannot be empty")
	}
	if len(mongoPassword) == 0 {
		fmt.Printf("mongo_password cannot be empty")
	}

	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(mongoAddress, ","),
		ConnectTimeoutSeconds: mongoConnectTimeout,
		Database:              mongoDatabase,
		Username:              mongoUsername,
		Password:              mongoPassword,
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		fmt.Printf("init mongo driver failed, err %s", err.Error())
		os.Exit(-1)
	}
	if err = mongoDB.Ping(); err != nil {
		blog.Fatalf("ping mongo db failed, err %s", err.Error())
	}
	blog.Infof("init mongo db successfully")

	var cidrList []Cidr
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		blog.Fatalf("read file %s failed, err %s", filename, err.Error())
	}
	err = json.Unmarshal(data, &cidrList)
	if err != nil {
		blog.Fatalf("json unmarshal failed, err %s", err.Error())
	}
	tkeStore := tke.New(mongoDB)
	for _, cidr := range cidrList {
		newCidr := &types.TkeCidr{
			VPC:        cidr.Vpc,
			CIDR:       cidr.Cidr,
			IPNumber:   uint32(cidr.IPNumber),
			Status:     cidr.Status,
			CreateTime: cidr.CreatedAt.String(),
			UpdateTime: cidr.UpdatedAt.String(),
		}
		if cidr.Cluster != nil {
			newCidr.Cluster = *cidr.Cluster
		}
		if err := tkeStore.CreateTkeCidr(context.TODO(), newCidr); err != nil {
			blog.Fatalf("create tke cidr %+v to mongo failed, err %s", newCidr, err.Error())
		}
	}
	blog.Infof("update cidr to mongo successfully")
}
