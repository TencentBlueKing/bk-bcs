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

// Package main xxx
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RichardKnop/machinery/v2/backends/mongo"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	etcdbroker "github.com/Tencent/bk-bcs/bcs-common/common/task/brokers/etcd"
	etcdlock "github.com/Tencent/bk-bcs/bcs-common/common/task/locks/etcd"
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	mongostore "github.com/Tencent/bk-bcs/bcs-common/common/task/store/mongo"
	mysqlstore "github.com/Tencent/bk-bcs/bcs-common/common/task/store/mysql"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	bcsmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
)

/*
场景测试
1. 正常分发任务并成功执行
2. 任务执行失败并暂停执行任务
3. 重试失败任务 / 设置skipOnFailed并重试成功
4. 任务跳过失败测试
5. step超时控制
6. task超时控制
7. 任务回调机制
*/

var (
	moduleName   = "example"
	queueAddress = "amqp://guest:guest@127.0.0.1:5672"
	mongoHosts   = []string{"127.0.0.1:27017"}
	mysqlDSN     = "root:%s@tcp(127.0.0.1:3306)/bk-env-manager-1?charset=utf8mb4&parseTime=True&loc=Local"
)

func main() {
	pwd := os.Getenv("MONGO_PASSWORD")
	if pwd == "" {
		pwd = "12345"
	}

	mongoOpts := &bcsmongo.Options{
		Hosts:                 mongoHosts,
		ConnectTimeoutSeconds: 10,
		Database:              "cluster",
		Username:              "root",
		Password:              pwd,
		MaxPoolSize:           0,
		MinPoolSize:           0,
	}
	mongoDB, err := bcsmongo.NewDB(mongoOpts)
	if err != nil {
		panic(err)
	}
	mongoCli, err := mongostore.NewMongoCli(mongoOpts)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	serverConfig := &config.Config{
		DefaultQueue:    "machinery_tasks",
		Broker:          "http://127.0.0.1:2379",
		Lock:            "http://127.0.0.1:2379",
		ResultsExpireIn: 3600 * 48,
		MongoDB: &config.MongoDBConfig{
			Client:   mongoCli,
			Database: mongoOpts.Database,
		},
	}
	broker, err := etcdbroker.New(ctx, serverConfig)
	if err != nil {
		panic(err)
	}
	lock, err := etcdlock.New(ctx, serverConfig, 3)
	if err != nil {
		panic(lock)
	}
	backend, err := mongo.New(serverConfig)
	if err != nil {
		panic(err)
	}
	store := mongostore.New(mongoDB, moduleName)

	if mysqlPwd := os.Getenv("MYSQL_PASSWORD"); mysqlPwd != "" {
		dns := fmt.Sprintf(mysqlDSN, mysqlPwd)
		store, err = mysqlstore.New(dns)
		if err != nil {
			panic(err)
		}
		err = store.EnsureTable(ctx)
		if err != nil {
			panic(err)
		}

	}

	btm := task.NewTaskManager(ctx)
	config := &task.ManagerConfig{
		ModuleName: moduleName,
		WorkerNum:  100,
		Broker:     broker,
		Backend:    backend,
		Lock:       lock,
		Store:      store,
	}
	// register step worker && callback
	config.CallBacks = registerCallbacks()

	// init task manager
	err = btm.Init(config)
	if err != nil {
		panic(err)
	}

	// run task manager
	btm.Run()

	// wait task server run
	time.Sleep(3 * time.Second)

	// build tak && run
	sum := NewExampleTask("3", "5")

	info := &types.TaskInfo{
		TaskIndex: "example",
		TaskType:  "example-test",
		TaskName:  "example",
		Creator:   "bcs",
	}
	sumTask, err := sum.BuildTask(info, types.WithTaskMaxExecutionSeconds(0),
		types.WithTaskCallBackFunc(callBackName))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = btm.Dispatch(sumTask)
	if err != nil {
		fmt.Println(err)
		return
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	btm.Stop()

	fmt.Printf("Got OS shutdown signal, shutting down server gracefully...")
}

func registerSteps() []istep.StepWorkerInterface {
	steps := make([]istep.StepWorkerInterface, 0)

	sum := NewSumStep()
	steps = append(steps, sum)

	hello := NewHelloStep()
	steps = append(steps, hello)

	return steps
}

func registerCallbacks() []istep.CallbackInterface {
	callbacks := make([]istep.CallbackInterface, 0)
	callbacks = append(callbacks, &callBack{})

	return callbacks
}
