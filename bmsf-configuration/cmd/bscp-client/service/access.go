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
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/pkg/logger"
)

const (
	// defaultDialTimeout is default dial timeout.
	defaultDialTimeout = 3 * time.Second
)

//NewOperator create business all operation interface
func NewOperator(op *option.Global) *AccessOperator {
	operator := &AccessOperator{
		Business: op.Business,
		User:     op.Operator,
		Token:    op.Token,
		index:    op.Index,
		limit:    op.Limit,
	}
	return operator
}

// CCredential is credential for per rpc.
type CCredential struct {
	auth string
}

// NewCCredential creates a new CCredential object.
func NewCCredential(auth string) *CCredential {
	return &CCredential{auth: auth}
}

// GetRequestMetadata returns metadatas for request.
func (c *CCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// Authorization: <type> <credentials>.
	// eg: grpcgateway-authorization: Basic YWRtaW46cGFzc3dvcmQ=
	return map[string]string{
		"grpcgateway-authorization": fmt.Sprintf("Basic %s", c.auth),
	}, nil
}

// RequireTransportSecurity returns flag for transport security.
func (c *CCredential) RequireTransportSecurity() bool {
	return false
}

//AccessOperator basic AccessServer tools
type AccessOperator struct {
	//yaml configuraiotn handler
	cfgHandler *viper.Viper
	//grpcConn for connecting AccessServer
	grpcConn *grpc.ClientConn
	//all common attributes shared in sub Class
	//Client access server client interface
	Client accessserver.AccessClient
	//Business name for operator
	Business string
	//User operator
	User string
	//Token
	Token string
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
	if !operator.cfgHandler.IsSet("host") {
		operator.cfgHandler.SetDefault("host", "client configration lost host")
	}
	if !operator.cfgHandler.IsSet("debug") {
		operator.cfgHandler.SetDefault("debug", false)
	}

	//other configuration part
	return nil
}

//initAccessClient create client for AccessServer
func (operator *AccessOperator) initAccessClient() error {
	bscpAuth := base64.StdEncoding.EncodeToString([]byte(operator.Token))
	conn, err := grpc.Dial(operator.cfgHandler.GetString("host"), grpc.WithInsecure(),
		grpc.WithTimeout(defaultDialTimeout), grpc.WithPerRPCCredentials(NewCCredential(bscpAuth)))
	if err != nil {
		return err
	}
	client := accessserver.NewAccessClient(conn)
	operator.grpcConn = conn
	operator.Client = client
	return nil
}
