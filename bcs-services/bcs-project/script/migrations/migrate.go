/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
迁移数据，从 bcs cc 服务模块(mysql 存储)，迁移到 bcs project模块(mongo 存储)
允许重复执行
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mysqlTableName = "projects"
	mongoTableName = "bcsproject_project"
	timeLayout     = "2006-01-02T15:04:05Z"
)

var (
	mysqlHost   string
	mysqlPort   uint
	mysqlUser   string
	mysqlPwd    string
	mysqlDBName string
	mongoHost   string
	mongoPort   uint
	mongoUser   string
	mongoPwd    string
	mongoDBName string
)

func parseFlags() {
	// mysql
	flag.StringVar(&mysqlHost, "mysql_host", "", "mysql host")
	flag.UintVar(&mysqlPort, "mysql_port", 0, "mysql port")
	flag.StringVar(&mysqlUser, "mysql_user", "", "access mysql username")
	flag.StringVar(&mysqlPwd, "mysql_pwd", "", "access mysql password")
	flag.StringVar(&mysqlDBName, "mysql_db_name", "", "access mysql db name")

	// mongo
	flag.StringVar(&mongoHost, "mongo_host", "", "mongo host")
	flag.UintVar(&mongoPort, "mongo_port", 0, "mongo port")
	flag.StringVar(&mongoUser, "mongo_user", "", "access mongo username")
	flag.StringVar(&mongoPwd, "mongo_pwd", "", "access mongo password")
	flag.StringVar(&mongoDBName, "mongo_db_name", "", "access mongo db name")

	flag.Parse()
}

func main() {
	parseFlags()
	// 获取数据
	fmt.Println("migrate start ...")
	p, err := fetchBCSCCData()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 写入 mongo
	if err := insertProject(p); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("migrate success!")
}

type BCSCCProjectData struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Name        string `json:"name" gorm:"size:64;unique"`
	EnglishName string `json:"english_name" gorm:"size:64;unique;index"`
	Creator     string `json:"creator" gorm:"size:32"`
	Updator     string `json:"updator" gorm:"size:32"`
	Description string `json:"desc" sql:"type:text"`
	ProjectType uint   `json:"project_type"`
	IsOfflined  bool   `json:"is_offlined" gorm:"default:false"`
	ProjectID   string `json:"project_id" gorm:"size:32;unique;index"`
	UseBK       bool   `json:"use_bk" gorm:"default:true"`
	CCAppID     uint   `json:"cc_app_id"`
	Kind        uint   `json:"kind"`
	DeployType  string `json:"deploy_type"`
	BGID        uint   `json:"bg_id"`
	BGName      string `json:"bg_name"`
	DeptID      uint   `json:"dept_id"`
	DeptName    string `json:"dept_name"`
	CenterID    uint   `json:"center_id"`
	CenterName  string `json:"center_name"`
	DataID      uint   `json:"data_id"`
	IsSecrecy   bool   `json:"is_secrecy" gorm:"default:false"`
}

// bcs cc 中查询数据
func fetchBCSCCData() ([]BCSCCProjectData, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", mysqlUser, mysqlPwd, mysqlHost, mysqlPort, mysqlDBName)

	// 连接 db
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("access mysql error, %s", err.Error())
	}
	db.DB().SetConnMaxLifetime(10 * time.Second)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(20)

	// 读取数据
	var p []BCSCCProjectData
	db.Table(mysqlTableName).Select("*").Scan(&p)
	return p, nil
}

// 插入到 mongo
func insertProject(p []BCSCCProjectData) error {
	var (
		client     *mongo.Client
		err        error
		collection *mongo.Collection
	)
	upsert := true
	opts := options.ReplaceOptions{Upsert: &upsert}
	// 建立连接
	dsn := fmt.Sprintf("mongodb://%s:%s@%s:%d", mongoUser, mongoPwd, mongoHost, mongoPort)
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(dsn).SetConnectTimeout(10*time.Second)); err != nil {
		return fmt.Errorf("access mongo error, %s", err.Error())
	}
	if err := client.Ping(context.TODO(), nil); err != nil {
		return fmt.Errorf("mongo ping error, %s", err.Error())
	}
	collection = client.Database(mongoDBName).Collection(mongoTableName)
	// 组装数据
	for _, i := range p {
		data := map[string]interface{}{
			"projectID":   i.ProjectID,
			"name":        i.Name,
			"projectCode": i.EnglishName,
			"creator":     i.Creator,
			"updater":     i.Updator,
			"managers":    constructManagers(i.Creator, i.Updator),
			"projectType": i.ProjectType,
			"useBKRes":    i.UseBK,
			"description": i.Description,
			"isOffline":   i.IsOfflined,
			"kind":        getStrKind(i.Kind),
			"businessID":  strconv.Itoa(int(i.CCAppID)),
			"deployType":  getDeployType(i.DeployType),
			"bgID":        strconv.Itoa(int(i.BGID)),
			"bgName":      i.BGName,
			"deptID":      strconv.Itoa(int(i.DeptID)),
			"deptName":    i.DeptName,
			"centerID":    strconv.Itoa(int(i.CenterID)),
			"centerName":  i.CenterName,
			"isSecret":    i.IsSecrecy,
			"createTime":  i.CreatedAt.Format(timeLayout),
			"updateTime":  i.CreatedAt.Format(timeLayout),
		}
		// 插入数据，允许重复操作
		if _, err := collection.ReplaceOne(
			context.TODO(), map[string]string{"projectID": i.ProjectID}, data, &opts,
		); err != nil {
			return err
		}
	}

	return nil
}

// 获取字符串类型 kind，1 => k8s 2 => mesos
func getStrKind(kind uint) string {
	if kind == 1 {
		return "k8s"
	} else if kind == 2 {
		return "mesos"
	}
	return ""
}

func stringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

// 组装manager
func constructManagers(creator string, updater string) string {
	managers := []string{creator}
	if updater != "" {
		if !stringInSlice(updater, managers) {
			managers = append(managers, updater)
		}
	}
	return strings.Join(managers, ";")
}

// 获取 int 型 deployType
func getDeployType(deployType string) uint32 {
	if deployType == "null" {
		return 1
	}
	return 2
}
