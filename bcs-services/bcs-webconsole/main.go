package main

import (
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/handler"
	yaml "github.com/asim/go-micro/plugins/config/encoder/yaml/v4"
	"github.com/urfave/cli/v2"

	mhttp "github.com/asim/go-micro/plugins/server/http/v4"
	"github.com/gin-gonic/gin"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source/file"
	"go-micro.dev/v4/logger"
)

var (
	service = "bcs-webconsole"
	version = "latest"
)

func main() {
	// Create service
	// srv := micro.NewService(
	// 	micro.Name(service),
	// 	micro.Version(version),
	// )

	var configPath string

	// new yaml encoder
	enc := yaml.NewEncoder()
	// new config
	conf, _ := config.NewConfig(
		config.WithReader(
			json.NewReader( // json reader for internal config merge
				reader.WithEncoder(enc),
			),
		),
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
		return nil
	})

	srv := micro.NewService(micro.Server(mhttp.NewServer()), confFlags)

	opts := []micro.Option{
		micro.Name(service),
		micro.Version(version),
		confAction,
	}
	srv.Init(opts...)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// 注册模板和静态资源
	router.SetHTMLTemplate(web.WebTemplate())
	router.StaticFS("/web_console/web/static", http.FS(web.WebStatic()))

	if err := handler.Register(handler.Options{Client: srv.Client(), Config: conf, Router: router}); err != nil {
		logger.Fatal(err)
	}
	if err := micro.RegisterHandler(srv.Server(), router); err != nil {
		logger.Fatal(err)
	}
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}

	// Register handler
	// pb.RegisterBcsWebconsoleHandler(srv.Server(), new(handler.BcsWebconsole))

	// // Run service
	// if err := srv.Run(); err != nil {
	// 	log.Fatal(err)
	// }
}
