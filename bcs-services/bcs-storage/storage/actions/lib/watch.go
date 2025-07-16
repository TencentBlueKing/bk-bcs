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
 */

package lib

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
)

// WatchServerOption options for watch server
type WatchServerOption struct {
	Store     *Store
	TableName string
	Cond      operator.M
	Req       *restful.Request
	Resp      *restful.Response
}

// WatchServer the server side handler in a watch action.
type WatchServer struct {
	store     *Store
	tableName string
	cond      operator.M
	req       *restful.Request
	resp      *restful.Response
	opts      *bcstypes.WatchOptions
	Writer    func(resp *restful.Response, event *Event) bool
}

var (
	// DefaultWriter default json writer
	DefaultWriter = func(resp *restful.Response, event *Event) bool {
		if err := codec.EncJsonWriter(event, resp.ResponseWriter); err != nil {
			blog.Errorf("defaultWriter error: %v", err)
			return false
		}
		return true
	}
)

// NewWatchServer create default json formate watchServer
func NewWatchServer(wopt *WatchServerOption) (*WatchServer, error) {
	if wopt.Req == nil || wopt.Resp == nil || wopt.Store == nil {
		return nil, fmt.Errorf("field in watch server option cannot be empty")
	}
	opts := new(bcstypes.WatchOptions)
	if err := codec.DecJsonReader(wopt.Req.Request.Body, opts); err != nil {
		return nil, err
	}
	return &WatchServer{
		store:     wopt.Store,
		tableName: wopt.TableName,
		cond:      wopt.Cond,
		req:       wopt.Req,
		resp:      wopt.Resp,
		opts:      opts,
		Writer:    DefaultWriter,
	}, nil
}

// Go running watchServer
func (ws *WatchServer) Go(ctx context.Context) {
	if ws.store == nil || ws.req == nil || ws.resp == nil {
		blog.Errorf(ws.sprint("Go() error, tank or req or resp can not be nil"))
		return
	}
	if ws.opts == nil {
		ws.opts = &bcstypes.WatchOptions{}
	}

	// metrics
	metrics.ReportWatchRequestInc(ws.req.SelectedRoutePath(), ws.tableName)

	notify := ws.resp.CloseNotify()
	ws.resp.WriteHeader(http.StatusOK)
	ws.resp.ResponseWriter.(http.Flusher).Flush()

	blog.Infof(ws.sprint("begin to watch"))

	watchOption := &StoreWatchOption{
		Cond:      ws.cond,
		SelfOnly:  ws.opts.SelfOnly,
		MaxEvents: ws.opts.MaxEvents,
		Timeout:   ws.opts.Timeout,
		MustDiff:  ws.opts.MustDiff,
	}

	defer func() {
		blog.Infof(ws.sprint("watch end"))
		metrics.ReportWatchRequestDec(ws.req.SelectedRoutePath(), ws.tableName)
		ws.Writer(ws.resp, EventWatchBreak)
		ws.resp.ResponseWriter.(http.Flusher).Flush()
	}()
	event, err := ws.store.Watch(ctx, ws.tableName, watchOption)
	if err != nil {
		blog.Errorf("watch failed, err %s", err.Error())
		return
	}

	for {
		select {
		case <-notify:
			blog.Infof(ws.sprint("stop watch by closing the connection in client side"))
			return
		case <-ctx.Done():
			blog.Infof(ws.sprint("stop watch by server"))
			return
		case e := <-event:
			if e.Type == Brk {
				blog.Infof(ws.sprint("stop watch by event break"))
				return
			}
			if ws.Writer(ws.resp, e) {
				blog.V(5).Infof(ws.sprint(fmt.Sprintf("flush: %v", e)))
			}
			ws.resp.ResponseWriter.(http.Flusher).Flush()
			metrics.ReportWatchHTTPResponseSize(ws.req.SelectedRoutePath(), ws.tableName, int64(ws.resp.ContentLength()))
		}
	}
}

func (ws *WatchServer) sprint(s string) string {
	return fmt.Sprintf("watch server %s | %s", ws.req.Request.URL, s)
}
