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

package upgrader

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/app/options"

	restful "github.com/emicklei/go-restful"
)

// Upgrader is a data struct of bcs upgrader server
type Upgrader struct {
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopCh        chan struct{}
	opt           *options.UpgraderOptions
	httpServer    *httpserver.HttpServer
	db            drivers.DB
	upgradeHelper *Helper
}

// NewUpgrader create upgrader server object
func NewUpgrader(op *options.UpgraderOptions) (*Upgrader, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Upgrader{
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopCh:        make(chan struct{}),
		opt:           op,
	}

	// Http server
	s.httpServer = httpserver.NewHttpServer(s.opt.Port, s.opt.Address, "")
	if s.opt.ServerCert.IsSSL {
		s.httpServer.SetSsl(
			s.opt.ServerCert.CAFile,
			s.opt.ServerCert.CertFile,
			s.opt.ServerCert.KeyFile,
			s.opt.ServerCert.CertPwd)
	}

	return s, nil
}

// initHTTPServer init all http service
func (u *Upgrader) initHTTPService() error {
	u.initUpgraderService()

	if u.opt.DebugMode {
		u.initPprofService()
	}
	return nil
}

// initUpgraderService init the upgrader service
func (u *Upgrader) initUpgraderService() {
	actions := []*httpserver.Action{
		httpserver.NewAction("POST", "/upgrade", nil, u.Upgrade),
	}
	u.httpServer.RegisterWebServer("/upgrader/v1", nil, actions)
}

// initPprofService init the pprof service
func (u *Upgrader) initPprofService() {
	action := []*httpserver.Action{
		httpserver.NewAction("GET", "/debug/pprof/", nil, getRouteFunc(pprof.Index)),
		httpserver.NewAction("GET", "/debug/pprof/{uri:*}", nil, getRouteFunc(pprof.Index)),
		httpserver.NewAction("GET", "/debug/pprof/cmdline", nil, getRouteFunc(pprof.Cmdline)),
		httpserver.NewAction("GET", "/debug/pprof/profile", nil, getRouteFunc(pprof.Profile)),
		httpserver.NewAction("GET", "/debug/pprof/symbol", nil, getRouteFunc(pprof.Symbol)),
		httpserver.NewAction("GET", "/debug/pprof/trace", nil, getRouteFunc(pprof.Trace)),
	}
	u.httpServer.RegisterWebServer("", nil, action)
}

// init mongo client
func (u *Upgrader) initMongoClient() error {
	if len(u.opt.MongoAddress) == 0 {
		return fmt.Errorf("mongo address cannot be empty")
	}
	if len(u.opt.MongoDatabase) == 0 {
		return fmt.Errorf("mongo database cannot be empty")
	}

	password := u.opt.MongoPassword
	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}
	mongoOptions := &mongo.Options{
		AuthMechanism:         u.opt.MongoAuthMechanism,
		Hosts:                 strings.Split(u.opt.MongoAddress, ","),
		ConnectTimeoutSeconds: int(u.opt.MongoConnectTimeout),
		Database:              u.opt.MongoDatabase,
		Username:              u.opt.MongoUsername,
		Password:              password,
		MaxPoolSize:           uint64(u.opt.MongoMaxPoolSize),
		MinPoolSize:           uint64(u.opt.MongoMinPoolSize),
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		blog.Errorf("init mongo client failed, err: %s", err)
		return err
	}

	if err = mongoDB.Ping(); err != nil {
		blog.Errorf("ping mongodb failed, err: %s", err)
		return err
	}
	u.db = mongoDB
	blog.Infof("init mongo client successfully")

	return nil
}

func (u *Upgrader) initSigalHandler() {
	// listen system signal
	// to run in the container, should not trap SIGTERM
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case e := <-interrupt:
			blog.Infof("receive interrupt %s, do close", e.String())
			u.stopCh <- struct{}{}
			u.close()
		}
	}()
}

func (u *Upgrader) initUpgradeHelper() {
	opt := &HelperOpt{
		DB:     u.db,
		config: u.opt.HttpCliConfig,
	}
	u.upgradeHelper = NewUpgradeHelper(opt)
}

func (u *Upgrader) close() {
	u.ctxCancelFunc()
}

// Start to run upgrader server
func (u *Upgrader) Start() error {
	// init system signale handler
	u.initSigalHandler()

	if err := u.initMongoClient(); err != nil {
		blog.Errorf("init mongo clinet failed, err: %s", err)
		return err
	}

	u.initUpgradeHelper()

	if err := u.initHTTPService(); err != nil {
		blog.Errorf("init http service failed, err: %s", err)
		return err
	}

	go func() {
		err := u.httpServer.ListenAndServe()
		blog.Errorf("http listen and service failed, err: %s", err)
		u.stopCh <- struct{}{}
	}()

	select {
	case <-u.stopCh:
		return nil
	}
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	}
}
