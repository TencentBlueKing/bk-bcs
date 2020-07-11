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

package tcpclient

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/protocol"
	"io"
	"net"
	"sync"
	"time"
)

// tcpClient tcp 客户端
type tcpClient struct {
	serverIP   string             // 服务器IP
	serverPort uint               // 服务器通信端口
	lastTime   time.Time          // 上一次收到服务器返回心跳包的时间
	connLock   sync.RWMutex       // conn lock
	conn       net.Conn           // 与服务器通信的链接
	cancel     context.CancelFunc // 终止协程的方法
}

// New 实例化TcpClient
func New(serverIP string, serverPort uint) ClientIf {
	return &tcpClient{serverIP: serverIP, serverPort: serverPort}
}

// heartBeat 处理心跳
func (cli *tcpClient) heartBeat(ctx context.Context) {

	cli.lastTime = time.Now()

	defer cli.cancel()

	for {
		select {
		case <-ctx.Done():
			blog.Warn("stop heartbeat detect")
			return
		case <-time.After(10 * time.Second): // 15 秒发送一次心跳包
			fmt.Printf("\nsend heart beat\n")
			// 心跳超时检测，超时一分钟
			if time.Now().Sub(cli.lastTime) > time.Minute {
				blog.Error("heart beat timeout, will close the connection")
				cli.connLock.Lock()
				cli.conn.Close()
				cli.conn = nil
				cli.connLock.Unlock()
				return
			}

			msgHead := &protocol.MsgHead{}

			msgHead.Type = protocol.HeartBeatDetect
			msgHead.Magic = protocol.ProtocolMagicNum
			msgHead.Length = 0
			msgHead.Timestamp = time.Now().UnixNano()

			packageData, packageErr := protocol.Package(msgHead, nil)

			if nil != packageErr {
				blog.Errorf("package data failed, error information is %s", packageErr.Error())
				cli.connLock.Lock()
				cli.conn.Close()
				cli.conn = nil
				cli.connLock.Unlock()

				return
			}

			cli.connLock.Lock()
			_, sendErr := cli.conn.Write(packageData)
			if sendErr != nil {

				blog.Error("fail to send data to server[%s], error %s", fmt.Sprintf("%s:%d", cli.serverIP, cli.serverPort), sendErr.Error())

				cli.conn.Close()
				cli.conn = nil
				cli.connLock.Unlock()
				return
			}
			cli.connLock.Unlock()
		}
	}
}

//Write 实现HandlerIf 接口，接收网络发送来的数据
func (cli *tcpClient) Write(head *protocol.MsgHead, data []byte) (int, error) {

	switch head.Type {
	case protocol.HeartBeatDetect:
		cli.lastTime = time.Now()
		blog.Debug("read heart package")
	}
	return 0, nil
}

// handleRead 解析收到的数据
func (cli *tcpClient) handleRead() {

	defer cli.cancel()

	// 为每个conn 创建一个ProtocolBuffer的
	protoBuffer := protocol.CreateProtocolBuffer(cli)

	for {

		// 流式数据，由系统自己读取并路由给接收者
		_, copyErr := io.Copy(protoBuffer, cli.conn)
		if nil != copyErr {
			cli.connLock.Lock()
			if nil != cli.conn {

				cli.conn.Close()
				cli.conn = nil
			}
			cli.connLock.Unlock()
			blog.Errorf("read data failed, error information is %s", copyErr.Error())
			return
		}

		if nil == copyErr {
			blog.Warnf("finish read, close the connection[%s]", cli.conn.RemoteAddr().String())
			cli.connLock.Lock()
			cli.conn.Close()
			cli.conn = nil
			cli.connLock.Unlock()
			return
		}

		blog.Fatalf("should not reach here, close connection[%s]", cli.conn.RemoteAddr().String())
		cli.connLock.Lock()
		cli.conn.Close()
		cli.conn = nil
		cli.connLock.Unlock()
		return

	}
}

// IsAlive 存货性判断
func (cli *tcpClient) IsAlive() bool {
	return cli.conn != nil
}

// Connect 连接服务器
func (cli *tcpClient) Connect() error {

	address := fmt.Sprintf("%s:%d", cli.serverIP, cli.serverPort)
	conn, err := net.Dial("tcp", address)

	if err != nil {
		blog.Errorf("connect remote server[%s] failed, error information is %s", address, err.Error())
		return err
	}

	cli.conn = conn

	// 启动心跳探测
	ctx, cancel := context.WithCancel(context.Background())
	cli.cancel = cancel
	// 启动读
	go cli.handleRead()
	// 启动心跳探测
	go cli.heartBeat(ctx)
	return nil
}

// Send 发送数据
func (cli *tcpClient) Send(extID int, data []byte) (int, error) {

	cli.connLock.Lock()
	defer cli.connLock.Unlock()

	if nil == cli.conn {
		if conErr := cli.Connect(); nil != conErr {
			return 0, fmt.Errorf("the connection is invalid, failed to reconnect, error information is %s ", conErr.Error())
		}
		return 0, fmt.Errorf("the connection is invalid")
	}

	msgHead := &protocol.MsgHead{}

	msgHead.Type = protocol.BKDataPlugin
	msgHead.Magic = protocol.ProtocolMagicNum
	msgHead.Length = uint32(len(data))
	msgHead.Timestamp = time.Now().UnixNano()

	packageData, packageErr := protocol.Package(msgHead, data)

	if nil != packageErr {
		blog.Errorf("package data failed, error information is %s", packageErr.Error())
		return 0, packageErr
	}

	cnt, sendErr := cli.conn.Write(packageData)
	if sendErr != nil {

		blog.Error(" error %s", sendErr.Error())
		cli.conn.Close()
		cli.conn = nil

		return 0, sendErr
	}

	return cnt, nil
}

// DisConnect 断链
func (cli *tcpClient) DisConnect() error {

	blog.Info("disconnect")
	if nil == cli.conn {
		return nil
	}

	cli.cancel()
	cli.conn.Close()
	cli.conn = nil
	return nil
}
