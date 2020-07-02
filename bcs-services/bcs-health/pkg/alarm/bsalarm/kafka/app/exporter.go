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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/bsalarm"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/emicklei/go-restful"
)

func NewExporter() (*Exporter, error) {
	exporter := &Exporter{stop: make(chan struct{})}

	api := new(restful.WebService).Path("/v1").Produces(restful.MIME_JSON)
	api.Route(api.POST("create").To(exporter.ReceiveEvent))
	server := http.Server{
		Handler: restful.NewContainer().Add(api),
	}

	unixListener, err := net.Listen("unix", bsalarm.DefaultSock)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := server.Serve(unixListener); err != nil {
			fmt.Fprintf(os.Stderr, "serve unix socket failed, err: %v\n", err)
			return
		}
	}()

	return exporter, nil
}

type Exporter struct {
	client   publisher.Client
	dataChan chan *bsalarm.StdInput
	stop     chan struct{}
}

func (e *Exporter) Initialize(b *beat.Beat, localConfig *common.Config) (beat.Beater, error) {
	fmt.Fprintf(os.Stdout, "data platform name: %v, version: %v", b.Name, b.Version)
	return e, nil
}

func (e *Exporter) Run(b *beat.Beat) error {
	e.client = b.Publisher.Connect()
	select {}
}

func (e *Exporter) Stop() {
	e.client.Close()
	close(e.stop)
	fmt.Fprintf(os.Stderr, "event exporter stopped.")
}

func (e *Exporter) Reload(*common.Config) {
	fmt.Fprintf(os.Stdout, "event exporter reloaded.")
	return
}

func (e *Exporter) ReceiveEvent(req *restful.Request, resp *restful.Response) {
	event := new(bsalarm.StdInput)
	output := new(bsalarm.StdOutput)
	output.UUID = event.UUID
	if err := json.NewDecoder(req.Request.Body).Decode(event); err != nil {
		blog.Errorf("decode event failed, err: %v", err)
		output.Message = err.Error()
		resp.WriteEntity(output)
		return
	}
	pkg := common.MapStr{
		"dataid":    event.DataID,
		"bcs_event": event.Data,
	}

	if nil != e.client {
		if !e.client.PublishEvent(pkg) {
			js, _ := json.Marshal(event.Data)
			blog.Errorf("publish bs alarm event failed. uuid: %s, event: %s", event.UUID, string(js))
			output.Success = false
			resp.WriteEntity(output)
			return
		}
		output.Success = true
		resp.WriteEntity(output)
		return
	}

	output.Success = false
	output.Message = "gse client is nil"
	resp.WriteEntity(output)
}
