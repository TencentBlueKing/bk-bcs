package main

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/handler"

	mhttp "github.com/asim/go-micro/plugins/server/http/v4"
	"github.com/gin-gonic/gin"
	micro "go-micro.dev/v4"
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
	srv := micro.NewService(micro.Server(mhttp.NewServer()))
	opts := []micro.Option{
		micro.Name(service),
		micro.Address("0.0.0.0:8083"),
		micro.Version(version),
	}

	srv.Init(opts...)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	if err := handler.Register(handler.Options{Client: srv.Client(), Router: router}); err != nil {
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
