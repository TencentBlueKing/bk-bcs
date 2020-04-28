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

package service

import (
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

//NewOperator create business all operation interface
func NewOperator(op *option.Global) *AccessOperator {
	operator := &AccessOperator{
		Business: op.Business,
		User:     op.Operator,
		index:    op.Index,
		limit:    op.Limit,
	}
	return operator
}

//AccessOperator basic AccessServer tools
type AccessOperator struct {
	//yaml configuraiotn handler
	cfgHandler *viper.Viper
	//grpcConn for connecting AccessServer
	grpcConn *grpclb.GRPCConn
	//all common attributes shared in sub Class
	//Client access server client interface
	Client accessserver.AccessClient
	//Business name for operator
	Business string
	//User operator
	User string
	//only for list command
	index int32
	limit int32
}

//Init access all basic components
func (operator *AccessOperator) Init(cfgfile string) error {
	if err := operator.initYAMLParser(cfgfile); err != nil {
		return err
	}
	if err := operator.initLogger(); err != nil {
		return err
	}
	if err := operator.checkInitialParameter(); err != nil {
		return err
	}
	if err := operator.initAccessClient(); err != nil {
		return err
	}
	return nil
}

//Stop stop all basic components
func (operator *AccessOperator) Stop() {
	logger.CloseLogs()
	operator.grpcConn.Close()
}

//initYAMLParser init yaml parser viper
func (operator *AccessOperator) initYAMLParser(cfg string) error {
	//check config file existence
	operator.cfgHandler = viper.New()
	operator.cfgHandler.SetConfigFile(cfg)
	err := operator.cfgHandler.ReadInConfig()
	//todo(DeveloperJim): setting default confiuratio
	//	that we don't need this configuration file
	return err
}

//initLogger for debug mode
func (operator *AccessOperator) initLogger() error {
	if !operator.cfgHandler.IsSet("logger.directory") {
		operator.cfgHandler.SetDefault("logger.directory", "./log")
	}
	if !operator.cfgHandler.IsSet("logger.maxsize") {
		operator.cfgHandler.SetDefault("logger.maxsize", 200)
	}
	if !operator.cfgHandler.IsSet("logger.maxnum") {
		operator.cfgHandler.SetDefault("logger.maxnum", 5)
	}
	if !operator.cfgHandler.IsSet("logger.stderr") {
		operator.cfgHandler.SetDefault("logger.stderr", false)
	}
	if !operator.cfgHandler.IsSet("logger.alsoStderr") {
		operator.cfgHandler.SetDefault("logger.alsoStderr", false)
	}
	if !operator.cfgHandler.IsSet("logger.level") {
		operator.cfgHandler.SetDefault("logger.level", 0)
	}
	if !operator.cfgHandler.IsSet("logger.stderrThreshold") {
		operator.cfgHandler.SetDefault("logger.stderrThreshold", 2)
	}
	logger.InitLogger(logger.LogConfig{
		LogDir:          operator.cfgHandler.GetString("logger.directory"),
		LogMaxSize:      operator.cfgHandler.GetUint64("logger.maxsize"),
		LogMaxNum:       operator.cfgHandler.GetInt("logger.maxnum"),
		ToStdErr:        operator.cfgHandler.GetBool("logger.stderr"),
		AlsoToStdErr:    operator.cfgHandler.GetBool("logger.alsoStderr"),
		Verbosity:       operator.cfgHandler.GetInt32("logger.level"),
		StdErrThreshold: operator.cfgHandler.GetString("logger.stderrThreshold"),
		VModule:         operator.cfgHandler.GetString("logger.vmodule"),
		TraceLocation:   operator.cfgHandler.GetString("traceLocation"),
	})
	return nil
}

//checkInitialParameter check yaml
func (operator *AccessOperator) checkInitialParameter() error {
	//access configuration part
	if !operator.cfgHandler.IsSet("access.servicename") {
		operator.cfgHandler.SetDefault("access.servicename", "bk-bscp-accessserver")
	}
	if !operator.cfgHandler.IsSet("access.timeout") {
		operator.cfgHandler.SetDefault("access.timeout", time.Second*3)
	}

	//discovery etcd configuration part
	if !operator.cfgHandler.IsSet("etcd.endpoints") {
		return fmt.Errorf("client configration lost etcd endpoints")
	}
	if !operator.cfgHandler.IsSet("etcd.dialTimeout") {
		operator.cfgHandler.SetDefault("etcd.dialTimeout", time.Second*3)
	}
	if !operator.cfgHandler.IsSet("debug") {
		operator.cfgHandler.SetDefault("debug", false)
	}

	//other configuration part
	return nil
}

//initAccessClient create client for AccessServer
func (operator *AccessOperator) initAccessClient() error {
	//internal grpc lb context
	ctx := &grpclb.Context{
		Target: operator.cfgHandler.GetString("access.servicename"),
		EtcdConfig: clientv3.Config{
			Endpoints:   operator.cfgHandler.GetStringSlice("etcd.endpoints"),
			DialTimeout: operator.cfgHandler.GetDuration("etcd.dialtimeout"),
		},
	}
	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(operator.cfgHandler.GetDuration("access.timeout")),
	}
	// build gRPC client of datamanager.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		return err
	}
	operator.grpcConn = conn
	operator.Client = accessserver.NewAccessClient(conn.Conn())
	return nil
}
