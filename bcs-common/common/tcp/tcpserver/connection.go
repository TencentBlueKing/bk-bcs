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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/protocol"
	"net"
	"time"
)

// connection 每个链接保持一个connection
type connection struct {
	lastTime time.Time          // 上一次收到服务器返回心跳包的时间
	conn     net.Conn           // 客户端发起的连接
	handler  protocol.HandlerIf // 逻辑层数据转发接口
}

// 发送心跳数据
func (cli *connection) sendHeartBeat() error {

	msgHead := &protocol.MsgHead{}

	msgHead.Type = protocol.HeartBeatDetect
	msgHead.Magic = protocol.ProtocolMagicNum
	msgHead.Length = 0
	msgHead.Timestamp = time.Now().UnixNano()

	packageData, packageErr := protocol.Package(msgHead, nil)

	if nil != packageErr {
		blog.Errorf("package data failed, error information is %s", packageErr.Error())
		return packageErr
	}

	_, sendErr := cli.conn.Write(packageData)
	if sendErr != nil {
		blog.Error("fail to send data to client[%s], error %s", cli.conn.RemoteAddr().String(), sendErr.Error())
		return sendErr
	}

	return nil
}

//Write 实现HandlerIf 接口，接收网络发送来的数据
func (cli *connection) Write(head *protocol.MsgHead, data []byte) (int, error) {

	switch head.Type {
	case protocol.HeartBeatDetect:
		cli.lastTime = time.Now()
		heartErr := cli.sendHeartBeat()
		blog.Debug("read heart package")
		return protocol.HeadLength(), heartErr
	default:
		// 其余数据转给逻辑层
		return cli.handler.Write(head, data)
	}
}

// isAlive 检测链接存活性
func (cli *connection) isAlive() bool {

	// 心跳超时检测，超时一分钟
	duration := time.Now().Sub(cli.lastTime)
	if duration > 5*time.Minute {
		blog.Error("heart beat timeout:%d", duration)
		return false
	}

	return true
}
