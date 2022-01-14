package main

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/handler"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/proto"

	micro "go-micro.dev/v4"
	log "go-micro.dev/v4/logger"
)

var (
	service = "bcs-webconsole"
	version = "latest"
)

func main() {
	// Create service
	srv := micro.NewService(
		micro.Name(service),
		micro.Version(version),
	)
	srv.Init()

	// Register handler
	pb.RegisterBcsWebconsoleHandler(srv.Server(), new(handler.BcsWebconsole))

	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
