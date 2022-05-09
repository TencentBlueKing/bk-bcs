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

package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	yaml "github.com/asim/go-micro/plugins/config/encoder/yaml/v4"
	etcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	mhttp "github.com/asim/go-micro/plugins/server/http/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/cmd"
	microConf "go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source/file"
	"go-micro.dev/v4/registry"
	"golang.org/x/sync/errgroup"
)

var (
	// 变量, 编译后覆盖
	service              = "webconsole.bkbcs.tencent.com"
	version              = "latest"
	credentialConfigPath = cli.StringSlice{}
)

// WebConsoleManager is an console struct
type WebConsoleManager struct {
	ctx           context.Context
	opt           *options.WebConsoleManagerOption
	microService  micro.Service
	microConfig   microConf.Config
	multiCredConf *options.MultiCredConf
}

// NewWebConsoleManager
func NewWebConsoleManager(opt *options.WebConsoleManagerOption) *WebConsoleManager {
	return &WebConsoleManager{
		ctx: context.Background(),
		opt: opt,
	}
}

func (c *WebConsoleManager) Init() error {
	// 初始化服务注册, 配置文件等
	microService, microConfig, multiCredConf := c.initMicroService()
	c.microService = microService
	c.microConfig = microConfig
	c.multiCredConf = multiCredConf

	// etcd 服务发现注册
	etcdRegistry, err := c.initEtcdRegistry()
	if err != nil {
		return err
	}

	if etcdRegistry != nil {
		microService.Init(micro.Registry(etcdRegistry))
	}

	// http 路由注册
	router, err := c.initHTTPService()
	if err != nil {
		return err
	}

	if err := micro.RegisterHandler(microService.Server(), router); err != nil {
		return err
	}

	return nil
}

func (c *WebConsoleManager) initMicroService() (micro.Service, microConf.Config, *options.MultiCredConf) {
	var (
		configPath string
	)

	// new config
	conf, _ := microConf.NewConfig(
		microConf.WithReader(json.NewReader(reader.WithEncoder(yaml.NewEncoder()))),
	)
	var multiCredConf *options.MultiCredConf

	cmdOptions := []cmd.Option{
		cmd.Description("bcs webconsole micro service"),
		cmd.Version(version),
	}

	microCmd := cmd.NewCmd(cmdOptions...)
	microCmd.App().Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "server_address",
			EnvVars: []string{"MICRO_SERVER_ADDRESS"},
			Usage:   "Bind address for the server. 127.0.0.1:8080",
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "config file path",
			Required:    true,
			Destination: &configPath,
		},
		&cli.StringSliceFlag{
			Name:        "credential-config",
			Usage:       "credential config file path",
			Required:    false,
			Destination: &credentialConfigPath,
		},
	}

	microCmd.App().Action = func(c *cli.Context) error {
		if err := conf.Load(file.NewSource(file.WithPath(configPath))); err != nil {
			return err
		}

		// 初始化配置文件
		if err := config.G.ReadFrom(conf.Bytes()); err != nil {
			logger.Errorf("config not valid, err: %s, exited", err)
			os.Exit(1)
		}

		logger.Infof("load conf from %s", configPath)

		// 授权信息
		if len(credentialConfigPath.Value()) > 0 {
			credConf, err := options.NewMultiCredConf(credentialConfigPath.Value())
			if err != nil {
				logger.Errorf("config not valid, err: %s, exited", err)
				os.Exit(1)
			}
			multiCredConf = credConf

		}
		return nil
	}

	srv := micro.NewService(micro.Server(mhttp.NewServer()))
	opts := []micro.Option{
		micro.Name(service),
		micro.Version(version),
		micro.Cmd(microCmd),
	}

	// 配置文件, 日志这里才设置完成
	srv.Init(opts...)

	return srv, conf, multiCredConf
}

// initHTTPService 初始化 gin Http 配置
func (c *WebConsoleManager) initHTTPService() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger(), cors.Default())
	router.Use(i18n.Localize())

	// 注册模板和静态资源
	router.SetHTMLTemplate(web.WebTemplate())

	// 静态资源
	routePrefix := config.G.Web.RoutePrefix
	if routePrefix == "" {
		routePrefix = "/webconsole"
	}

	// 支持路径 prefix 透传和 rewrite 的场景
	router.Group(routePrefix).StaticFS("/web/static", http.FS(web.WebStatic()))
	router.Group("").StaticFS("/web/static", http.FS(web.WebStatic()))

	handlerOpts := &route.Options{
		RoutePrefix: routePrefix,
		Client:      c.microService.Client(),
		Router:      router,
	}

	// 注册 HTTP 请求
	for _, r := range []route.Registrar{
		web.NewRouteRegistrar(handlerOpts),
		api.NewRouteRegistrar(handlerOpts),
	} {
		r.RegisterRoute(router.Group(routePrefix))
		r.RegisterRoute(router.Group(""))
	}

	return router, nil
}

// initEtcdRegistry etcd 服务注册
func (c *WebConsoleManager) initEtcdRegistry() (registry.Registry, error) {
	endpoints := c.microConfig.Get("etcd", "endpoints").String("")
	if endpoints == "" {
		return nil, nil
	}

	etcdRegistry := etcd.NewRegistry(registry.Addrs(strings.Split(endpoints, ",")...))

	ca := c.microConfig.Get("etcd", "ca").String("")
	cert := c.microConfig.Get("etcd", "cert").String("")
	key := c.microConfig.Get("etcd", "key").String("")
	if ca != "" && cert != "" && key != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(ca, cert, key, "")
		if err != nil {
			return nil, err
		}
		etcdRegistry.Init(registry.TLSConfig(tlsConfig))
	}

	return etcdRegistry, nil
}

// Run create a pid
func (c *WebConsoleManager) Run() error {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(c.ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("starting bcs-webconsole.")

	c.microService.Init(micro.AfterStop(func() error {
		// 会让 websocket 发送 EndOfTransmission, 不能保证一定发送成功
		logger.Info("receive interput, gracefully shutdown")
		<-ctx.Done()
		return nil
	}))

	eg, ctx := errgroup.WithContext(ctx)

	podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
	eg.Go(func() error {
		return podCleanUpMgr.Run()
	})

	if c.multiCredConf != nil {
		c.microService.Init(micro.AfterStop(func() error {
			c.multiCredConf.Stop()
			return nil
		}))

		eg.Go(func() error {
			return c.multiCredConf.Watch()
		})
	}

	eg.Go(func() error {
		if err := c.microService.Run(); err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		defer logger.CloseLogs()
		return err
	}
	return nil
}
