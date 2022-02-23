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

package main

import (
	"context"
	"net/http"
	"os/signal"
	"path"
	"strings"
	"syscall"

	consoleConf "github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	yaml "github.com/asim/go-micro/plugins/config/encoder/yaml/v4"
	etcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	mhttp "github.com/asim/go-micro/plugins/server/http/v4"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source/file"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
)

var (
	// 变量, 编译后覆盖
	service = "bcs-webconsole"
	version = "latest"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)

	blogConf := conf.LogConfig{
		Verbosity:    3,
		AlsoToStdErr: true,
		LogDir:       "",
		LogMaxSize:   100,
		LogMaxNum:    7,
	}
	blog.InitLogs(blogConf)
	// CloseLogs() can assure you that you can not lose any log.
	defer blog.CloseLogs()

	blog.Info("starting bcs-webconsole.")

	var configPath string

	// new config
	conf, _ := config.NewConfig(
		config.WithReader(json.NewReader(reader.WithEncoder(yaml.NewEncoder()))),
	)

	confFlags := micro.Flags(
		&cli.StringFlag{
			Name:        "bcs-conf",
			Usage:       "bcs-config path",
			Required:    true,
			Destination: &configPath,
		},
	)

	confAction := micro.Action(func(c *cli.Context) error {
		logger.Info("load conf from ", configPath)
		if err := conf.Load(file.NewSource(file.WithPath(configPath))); err != nil {
			return err
		}

		// 初始化配置文件
		consoleConf.G.ReadFrom(conf.Bytes())

		return nil
	})

	srv := micro.NewService(micro.Server(mhttp.NewServer()), confFlags)
	opts := []micro.Option{
		micro.Name(service),
		micro.Version(version),
		confAction,
	}
	srv.Init(opts...)

	// etcd 服务注册
	endpoints := conf.Get("etcd", "endpoints").String("127.0.0.1:2379")
	etcdRegistry := etcd.NewRegistry(registry.Addrs(strings.Split(endpoints, ",")...))

	ca := conf.Get("etcd", "ca").String("")
	cert := conf.Get("etcd", "cert").String("")
	key := conf.Get("etcd", "key").String("")
	if ca != "" && cert != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(ca, cert, key, "")
		if err != nil {
			logger.Fatal(err)
		}
		etcdRegistry.Init(registry.TLSConfig(tlsConfig))
	}

	etcdR := micro.Registry(etcdRegistry)
	srv.Init(etcdR)

	srv.Init(micro.AfterStop(func() error {
		// 会让 websocket 发送 EndOfTransmission, 不能保证一定发送成功
		logger.Info("receive interput, gracefully shutdown")
		<-ctx.Done()
		return nil
	}))

	podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
	eg.Go(func() error {
		return podCleanUpMgr.Run()
	})

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(i18n.Localize())

	// 注册模板和静态资源
	router.SetHTMLTemplate(web.WebTemplate())

	// 静态资源
	routePrefix := conf.Get("web", "route_prefix").String("")
	if routePrefix == "" {
		routePrefix = "/" + service
	}

	// 支持路径 prefix 透传和 rewrite 的场景
	router.StaticFS(path.Join(routePrefix, "/web/static"), http.FS(web.WebStatic()))
	router.StaticFS("/web/static", http.FS(web.WebStatic()))

	handlerOpts := &route.Options{
		RoutePrefix: routePrefix,
		Client:      srv.Client(),
		Router:      router,
	}

	if err := handler.Register(handlerOpts); err != nil {
		logger.Fatal(err)
	}
	if err := micro.RegisterHandler(srv.Server(), router); err != nil {
		logger.Fatal(err)
	}

	eg.Go(func() error {
		if err := srv.Run(); err != nil {
			logger.Fatal(err)
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		logger.Fatal(err)
		return
	}

}
