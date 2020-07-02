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

package tcpserver

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/protocol"
	"io"
	"net"
	"time"
)

// tcpServer 服务对象定义
type tcpServer struct {
	listenIP    string
	listenPort  uint
	headLength  int
	netListener net.Listener
	handler     protocol.HandlerIf
}

// New tcpServer
func New(listenIP string, listenPort uint, handlerIf protocol.HandlerIf) ServerIf {
	return &tcpServer{listenIP: listenIP, listenPort: listenPort, handler: handlerIf}
}

// HandleConnection 新连接注册
func (cli *tcpServer) handleConnection(conn net.Conn) {

	// 为每个conn 创建一个ProtocolBuffer的
	tmpConn := &connection{conn: conn, handler: cli.handler, lastTime: time.Now()}
	protoBuffer := protocol.CreateProtocolBuffer(tmpConn)

	// 启动协成检测心跳
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context, tmpCon *connection) {
		for {
			select {
			case <-ctx.Done():
				blog.Warn("finish check heart beat")
				return
			case <-time.After(10 * time.Second): // 10 秒钟检测一次

				blog.Debug("check heart beat")
				if !tmpCon.isAlive() {
					// 心跳超时，需要关闭连接
					blog.Warn("connection[%s] is not alive", tmpCon.conn.RemoteAddr().String())
					tmpCon.conn.Close()
				}
			}
		}
	}(ctx, tmpConn)

	// 接收数据
	for {
		// 流式数据，由系统自己读取并路由给接收者
		_, copyErr := io.Copy(protoBuffer, conn)
		if nil != copyErr {
			blog.Errorf("read error, will close the connection[%s] error information is %s", conn.RemoteAddr().String(), copyErr.Error())
			conn.Close()
			return
		}

		if nil == copyErr {
			blog.Warnf("finish read, close the connection[%s]", conn.RemoteAddr().String())
			conn.Close()
			return
		}

		blog.Fatalf("should not reach here, close connection[%s]", conn.RemoteAddr().String())
		conn.Close()
		return

	}
}

// Start 启动服务
func (cli *tcpServer) Start() error {

	blog.Info("TcpServer start")

	address := fmt.Sprintf("%s:%d", cli.listenIP, cli.listenPort)
	netListener, netErr := net.Listen("tcp", address)
	cli.netListener = netListener
	if nil != netErr {
		blog.Error("listen the address %s fail, error:%s", address, netErr.Error())
		return netErr
	}

	cli.headLength = protocol.HeadLength()

	for {

		conn, connErr := netListener.Accept()
		if nil != connErr {
			blog.Error("accept failed, error %s", connErr.Error())
			continue
		}

		// 启动goroutine 处理链接
		blog.Info("%s tcp connect success", conn.RemoteAddr().String())
		go cli.handleConnection(conn)
	}
}

// Stop 停止服务
func (cli *tcpServer) Stop() error {
	if nil != cli.netListener {
		cli.netListener.Close()
	}
	return nil
}
