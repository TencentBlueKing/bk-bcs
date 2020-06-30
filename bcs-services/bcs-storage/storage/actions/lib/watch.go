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

package lib

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	restful "github.com/emicklei/go-restful"
)

// The ServerSide handler in a watch action.
type watchServer struct {
	tank operator.Tank
	opts *operator.WatchOptions
	req  *restful.Request
	resp *restful.Response

	Writer func(resp *restful.Response, event *operator.Event) bool
}

var (
	//DefaultWriter default json writer
	DefaultWriter = func(resp *restful.Response, event *operator.Event) bool {
		if err := codec.EncJsonWriter(event, resp.ResponseWriter); err != nil {
			blog.Errorf("defaultWriter error: %v", err)
			return false
		}
		return true
	}
)

//NewWatchServer create default json formate watchServer
func NewWatchServer(req *restful.Request, resp *restful.Response, tank operator.Tank) (*watchServer, error) {
	opts := new(operator.WatchOptions)
	if err := codec.DecJsonReader(req.Request.Body, opts); err != nil {
		return nil, err
	}
	return &watchServer{
		req:    req,
		resp:   resp,
		tank:   tank,
		opts:   opts,
		Writer: DefaultWriter,
	}, nil
}

//Go running watchServer
func (ws *watchServer) Go(ctx context.Context) {
	if ws.tank == nil || ws.req == nil || ws.resp == nil {
		blog.Errorf(ws.sprint("Go() error, tank or req or resp can not be nil"))
		return
	}
	if ws.opts == nil {
		ws.opts = &operator.WatchOptions{}
	}

	notify := ws.resp.CloseNotify()
	ws.resp.WriteHeader(http.StatusOK)
	ws.resp.ResponseWriter.(http.Flusher).Flush()

	blog.Infof(ws.sprint("begin to watch"))
	event, cancel := ws.tank.Watch(ws.opts)
	defer func() {
		cancel()
		blog.Infof(ws.sprint("watch end"))
		ws.Writer(ws.resp, operator.EventWatchBreak)
		ws.resp.ResponseWriter.(http.Flusher).Flush()
	}()
	for {
		select {
		case <-notify:
			blog.Infof(ws.sprint("stop watch by closing the connection in client side"))
			return
		case <-ctx.Done():
			blog.Infof(ws.sprint("stop watch by server"))
			return
		case e := <-event:
			if ws.Writer(ws.resp, e) {
				blog.Infof(ws.sprint(fmt.Sprintf("flush: %v", e)))
			}
			ws.resp.ResponseWriter.(http.Flusher).Flush()
		}
	}
}

func (ws *watchServer) sprint(s string) string {
	return fmt.Sprintf("watch server %s | %s", ws.req.Request.URL, s)
}
