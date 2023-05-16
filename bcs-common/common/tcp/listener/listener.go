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

package listener

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
)

// Connection 新的连接
type Connection struct {
	conn net.Conn
	err  error
}

// DualStackListener 双栈监听
type DualStackListener struct {
	// 取消
	ctx    context.Context
	cancel context.CancelFunc

	// subListeners 子listen
	subListeners []net.Listener
	// acceptOnce 启动子Listen的Accept()方法 只执行一次 doOnce() 方法
	acceptOnce *sync.Once
	// confirmClose 确认关闭
	confirmClose *sync.WaitGroup

	// connections chan
	connections chan *Connection
	// closeOnce 关闭 connections chan 只执行一次 close() 方法
	closeOnce *sync.Once
}

// NewDualStackListener 创建一个双栈监听
func NewDualStackListener() *DualStackListener {
	ctx, cancel := context.WithCancel(context.Background())
	return &DualStackListener{
		ctx:          ctx,
		cancel:       cancel,
		acceptOnce:   &sync.Once{},
		closeOnce:    &sync.Once{},
		confirmClose: &sync.WaitGroup{},
		connections:  make(chan *Connection, 1),
		subListeners: make([]net.Listener, 0),
	}
}

func (d *DualStackListener) addSubListener(address string) error {
	listen, err := net.Listen(types.TCP, address)
	if err != nil {
		if util.CheckBindError(err) {
			blog.Warn("unable to listen %s, err: %s", address, err.Error())
			return nil
		}
		return err
	}
	d.subListeners = append(d.subListeners, listen)
	return nil
}

// AddListener 添加监听
func (d *DualStackListener) AddListener(ip, port string) error {
	return d.addSubListener(net.JoinHostPort(ip, port))
}

// AddListenerWithAddr 添加监听
// 如：127.0.0.1:8000
func (d *DualStackListener) AddListenerWithAddr(address string) error {
	return d.addSubListener(address)
}

// Accept waits for and returns the next connection to the listener.
func (d *DualStackListener) Accept() (net.Conn, error) {
	// 所有子listener循环读取accept方法 (只执行一次！)
	d.acceptOnce.Do(d.doOnce)
	// 等待任意子listener新的连接
	select {
	case connection := <-d.connections:
		return connection.conn, connection.err
	case <-d.ctx.Done():
		return nil, http.ErrServerClosed
	}
}

// doOnce 执行一次
func (d *DualStackListener) doOnce() {
	if len(d.subListeners) == 0 {
		blog.Fatalf("DualStackListener doesn't have any child listeners available.")
	}
	if len(d.subListeners) == 1 {
		blog.Warnf("DualStackListener is running in single stack mode.")
	}
	for _, listener := range d.subListeners {
		d.confirmClose.Add(1)
		go func(l net.Listener) {
			d.subAccept(accept(l))
		}(listener)
	}
}

// accept 返回每个Listener的Accept结果
func accept(l net.Listener) func() *Connection {
	return func() *Connection {
		conn, err := l.Accept()
		return &Connection{
			conn: conn,
			err:  err,
		}
	}
}

// subAccept 每个子listener读取自己的Accept()
func (d *DualStackListener) subAccept(accept func() *Connection) {
	for {
		select {
		case <-d.ctx.Done():
			// 收到取消信号，打卡结束执行
			d.confirmClose.Done()
			return
		case d.connections <- accept():
			// 继续执行
		}
	}
}

// close 等待所有子listen打卡关闭，然后再去关闭 connections chan
func (d *DualStackListener) close() {
	// 等待所有 子listen 结束执行
	d.confirmClose.Wait()
	// 所有子listen执行结束后，再去关闭 chan
	d.closeOnce.Do(func() {
		close(d.connections)
	})
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (d *DualStackListener) Close() error {
	defer d.close()
	// 关闭子Listener
	for _, listener := range d.subListeners {
		if err := listener.Close(); err != nil {
			return errors.Wrapf(err, "close sub listener failed.addr:%v", listener.Addr())
		}
	}
	// 取消所有子Listener，停止读取accept
	d.cancel()
	return nil
}

// Addr returns the listener's network address.
func (d *DualStackListener) Addr() net.Addr {
	if len(d.subListeners) > 0 {
		return d.subListeners[0].Addr()
	}
	return nil
}
